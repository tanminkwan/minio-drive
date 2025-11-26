package rclone

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"simple-uploader/internal/config"
	"strings"
	"syscall"
)

const remoteName = "minio"

type Manager struct {
	rclonePath string
	configPath string
	cfg        *config.Config
	serveCmd   *exec.Cmd  // WebDAV server process
	mountCmd   *exec.Cmd  // WinFsp mount process
}

// NewManager creates a new rclone manager
func NewManager(cfg *config.Config) (*Manager, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	exeDir := filepath.Dir(exePath)
	rclonePath := filepath.Join(exeDir, "rclone.exe")

	// Check if rclone exists
	if _, err := os.Stat(rclonePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("rclone.exe not found at %s", rclonePath)
	}

	// Config file in same directory
	configPath := filepath.Join(exeDir, "rclone.conf")

	return &Manager{
		rclonePath: rclonePath,
		configPath: configPath,
		cfg:        cfg,
	}, nil
}

// GenerateConfig creates rclone.conf for MinIO
func (m *Manager) GenerateConfig() error {
	protocol := "http"
	if m.cfg.MinIO.UseSSL {
		protocol = "https"
	}

	// Ensure endpoint doesn't have protocol prefix
	endpoint := m.cfg.MinIO.Endpoint
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	configContent := fmt.Sprintf(`[%s]
type = s3
provider = Minio
access_key_id = %s
secret_access_key = %s
endpoint = %s://%s
force_path_style = true
`,
		remoteName,
		m.cfg.MinIO.AccessKey,
		m.cfg.MinIO.SecretKey,
		protocol,
		endpoint,
	)

	return os.WriteFile(m.configPath, []byte(configContent), 0600)
}

// KillExistingProcesses kills any existing rclone processes
func (m *Manager) KillExistingProcesses() {
	// Kill rclone
	cmd := exec.Command("taskkill", "/F", "/IM", "rclone.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run()
}

// StartWebDAV starts the WebDAV server
func (m *Manager) StartWebDAV() error {
	// Kill any existing processes first
	m.KillExistingProcesses()

	if m.serveCmd != nil && m.serveCmd.Process != nil {
		m.serveCmd = nil
	}

	// Generate config file
	if err := m.GenerateConfig(); err != nil {
		return fmt.Errorf("failed to generate rclone config: %w", err)
	}

	remotePath := fmt.Sprintf("%s:%s", remoteName, m.cfg.MinIO.Bucket)
	addr := fmt.Sprintf("localhost:%d", m.cfg.Mount.Port)

	// Build serve webdav command
	m.serveCmd = exec.Command(m.rclonePath,
		"serve", "webdav",
		"--config", m.configPath,
		"--addr", addr,
		remotePath,
	)

	// Hide console window on Windows
	m.serveCmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	if err := m.serveCmd.Start(); err != nil {
		return fmt.Errorf("failed to start WebDAV server: %w", err)
	}

	return nil
}

// MountWinFsp mounts using WinFsp (rclone mount)
func (m *Manager) MountWinFsp() error {
	// Kill any existing processes first
	m.KillExistingProcesses()

	if m.mountCmd != nil && m.mountCmd.Process != nil {
		m.mountCmd = nil
	}

	// Generate config file
	if err := m.GenerateConfig(); err != nil {
		return fmt.Errorf("failed to generate rclone config: %w", err)
	}

	remotePath := fmt.Sprintf("%s:%s", remoteName, m.cfg.MinIO.Bucket)
	driveLetter := m.GetDriveLetter()

	// Build mount command
	m.mountCmd = exec.Command(m.rclonePath,
		"mount",
		"--config", m.configPath,
		"--vfs-cache-mode", "full",
		remotePath,
		driveLetter,
	)

	// Hide console window on Windows
	m.mountCmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	if err := m.mountCmd.Start(); err != nil {
		return fmt.Errorf("failed to mount drive: %w", err)
	}

	return nil
}

// UnmountWinFsp unmounts WinFsp drive
func (m *Manager) UnmountWinFsp() error {
	if m.mountCmd == nil || m.mountCmd.Process == nil {
		return nil
	}

	// Kill the rclone mount process
	if err := m.mountCmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill rclone mount: %w", err)
	}

	m.mountCmd = nil
	return nil
}

// IsMounted checks if WinFsp drive is mounted
func (m *Manager) IsMounted() bool {
	return m.mountCmd != nil && m.mountCmd.Process != nil
}

// StopWebDAV stops the WebDAV server
func (m *Manager) StopWebDAV() error {
	if m.serveCmd == nil || m.serveCmd.Process == nil {
		return nil
	}

	// Kill the rclone process
	if err := m.serveCmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill rclone process: %w", err)
	}

	m.serveCmd = nil
	return nil
}

// IsRunning checks if WebDAV server is running
func (m *Manager) IsRunning() bool {
	return m.serveCmd != nil && m.serveCmd.Process != nil
}

// GetWebDAVURL returns the WebDAV server URL
func (m *Manager) GetWebDAVURL() string {
	return fmt.Sprintf("http://localhost:%d", m.cfg.Mount.Port)
}

// GetDriveLetter returns the configured drive letter
func (m *Manager) GetDriveLetter() string {
	letter := m.cfg.Mount.DriveLetter
	if !strings.HasSuffix(letter, ":") {
		letter += ":"
	}
	return letter
}

// ConnectDrive connects WebDAV as network drive
func (m *Manager) ConnectDrive() error {
	driveLetter := m.GetDriveLetter()
	url := m.GetWebDAVURL()

	// Disconnect existing drive first
	m.DisconnectDrive()

	cmd := exec.Command("net", "use", driveLetter, url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to connect drive: %w", err)
	}
	return nil
}

// DisconnectDrive disconnects the network drive
func (m *Manager) DisconnectDrive() error {
	driveLetter := m.GetDriveLetter()

	cmd := exec.Command("net", "use", driveLetter, "/delete", "/y")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	_ = cmd.Run() // Ignore error if not connected
	return nil
}

