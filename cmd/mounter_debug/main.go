package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"simple-uploader/internal/config"
)

func main() {
	fmt.Println("=== MinIO Mounter Debug ===")

	// Check executable path
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("ERROR getting exe path: %v\n", err)
		waitExit()
		return
	}
	exeDir := filepath.Dir(exePath)
	fmt.Printf("Executable dir: %s\n", exeDir)

	// Check rclone exists
	rclonePath := filepath.Join(exeDir, "rclone.exe")
	fmt.Printf("\n[1] Checking rclone at: %s\n", rclonePath)
	if _, err := os.Stat(rclonePath); os.IsNotExist(err) {
		fmt.Println("ERROR: rclone.exe not found!")
		waitExit()
		return
	}
	fmt.Println("rclone.exe found")

	// Load config
	fmt.Println("\n[2] Loading config...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("ERROR loading config: %v\n", err)
		waitExit()
		return
	}

	mountType := cfg.Mount.Type
	if mountType == "" {
		mountType = "webdav"
	}
	fmt.Printf("Config: endpoint=%s, bucket=%s, type=%s, port=%d, drive=%s\n",
		cfg.MinIO.Endpoint, cfg.MinIO.Bucket, mountType, cfg.Mount.Port, cfg.Mount.DriveLetter)

	// Generate rclone config
	fmt.Println("\n[3] Generating rclone config...")
	rcloneConfigPath := filepath.Join(exeDir, "rclone.conf")

	protocol := "http"
	if cfg.MinIO.UseSSL {
		protocol = "https"
	}

	configContent := fmt.Sprintf(`[minio]
type = s3
provider = Minio
access_key_id = %s
secret_access_key = %s
endpoint = %s://%s
force_path_style = true
`,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		protocol,
		cfg.MinIO.Endpoint,
	)

	if err := os.WriteFile(rcloneConfigPath, []byte(configContent), 0600); err != nil {
		fmt.Printf("ERROR writing rclone config: %v\n", err)
		waitExit()
		return
	}
	fmt.Printf("Rclone config written to: %s\n", rcloneConfigPath)

	// Test rclone connection
	fmt.Println("\n[4] Testing rclone connection...")
	cmd := exec.Command(rclonePath, "--config", rcloneConfigPath, "lsd", "minio:")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("ERROR listing buckets: %v\n", err)
		fmt.Printf("Output: %s\n", output)
		waitExit()
		return
	}
	fmt.Printf("Buckets:\n%s\n", output)

	// Kill existing rclone processes
	fmt.Println("\n[5] Killing existing rclone processes...")
	killCmd := exec.Command("taskkill", "/F", "/IM", "rclone.exe")
	killCmd.Run() // Ignore error
	time.Sleep(500 * time.Millisecond)

	driveLetter := cfg.Mount.DriveLetter
	if len(driveLetter) == 1 {
		driveLetter += ":"
	}

	if cfg.IsWinFsp() {
		// WinFsp mount
		fmt.Println("\n[6] Starting WinFsp mount...")
		remotePath := fmt.Sprintf("minio:%s", cfg.MinIO.Bucket)

		fmt.Printf("Command: rclone mount %s %s\n", remotePath, driveLetter)

		mountCmd := exec.Command(rclonePath,
			"mount",
			"--config", rcloneConfigPath,
			"--vfs-cache-mode", "full",
			remotePath,
			driveLetter,
		)
		mountCmd.Stdout = os.Stdout
		mountCmd.Stderr = os.Stderr

		if err := mountCmd.Start(); err != nil {
			fmt.Printf("ERROR mounting: %v\n", err)
			waitExit()
			return
		}
		fmt.Printf("Mount started (PID: %d)\n", mountCmd.Process.Pid)

		fmt.Println("\n[7] Waiting for mount...")
		time.Sleep(3 * time.Second)

		fmt.Println("\n=== SUCCESS! ===")
		fmt.Printf("Drive %s mounted (WinFsp)\n", driveLetter)
		fmt.Println("\nPress Enter to unmount and exit...")
		fmt.Scanln()

		// Cleanup
		mountCmd.Process.Kill()
		fmt.Println("Unmounted.")
	} else {
		// WebDAV mode
		fmt.Println("\n[6] Starting WebDAV server...")
		addr := fmt.Sprintf("localhost:%d", cfg.Mount.Port)
		remotePath := fmt.Sprintf("minio:%s", cfg.MinIO.Bucket)

		fmt.Printf("Command: rclone serve webdav --addr %s %s\n", addr, remotePath)

		serveCmd := exec.Command(rclonePath,
			"serve", "webdav",
			"--config", rcloneConfigPath,
			"--addr", addr,
			remotePath,
		)
		serveCmd.Stdout = os.Stdout
		serveCmd.Stderr = os.Stderr

		if err := serveCmd.Start(); err != nil {
			fmt.Printf("ERROR starting WebDAV: %v\n", err)
			waitExit()
			return
		}
		fmt.Printf("WebDAV server started (PID: %d)\n", serveCmd.Process.Pid)

		// Wait for server to start
		fmt.Println("\n[7] Waiting for server...")
		time.Sleep(2 * time.Second)

		// Disconnect existing drive first
		fmt.Println("\n[8] Disconnecting existing drive...")
		disconnectCmd := exec.Command("net", "use", driveLetter, "/delete", "/y")
		disconnectCmd.Run() // Ignore error

		// Connect network drive
		fmt.Println("\n[9] Connecting network drive...")
		url := fmt.Sprintf("http://localhost:%d", cfg.Mount.Port)

		fmt.Printf("Command: net use %s %s\n", driveLetter, url)

		netCmd := exec.Command("net", "use", driveLetter, url)
		output, err = netCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("ERROR connecting drive: %v\n", err)
			fmt.Printf("Output: %s\n", output)
			fmt.Println("\nWebDAV server is still running. Press Enter to stop and exit...")
			fmt.Scanln()
			serveCmd.Process.Kill()
			return
		}
		fmt.Printf("Drive connected: %s\n", output)

		fmt.Println("\n=== SUCCESS! ===")
		fmt.Printf("WebDAV server running at %s\n", url)
		fmt.Printf("Drive %s connected\n", driveLetter)
		fmt.Println("\nPress Enter to disconnect and exit...")
		fmt.Scanln()

		// Cleanup
		exec.Command("net", "use", driveLetter, "/delete", "/y").Run()
		serveCmd.Process.Kill()
		fmt.Println("Disconnected and stopped.")
	}
}

func waitExit() {
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}
