package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	menuName     = "Upload2Cloud"
	menuText     = "Upload to Cloud"
	shellKeyPath = `*\shell\` + menuName
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		if err := install(); err != nil {
			fmt.Printf("Installation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Installation completed successfully!")
		fmt.Println("Right-click any file to see 'Upload to Cloud' menu.")
	case "uninstall":
		if err := uninstall(); err != nil {
			fmt.Printf("Uninstallation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Uninstallation completed successfully!")
	case "status":
		checkStatus()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Simple Uploader Installer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  installer.exe install   - Install context menu and startup")
	fmt.Println("  installer.exe uninstall - Remove context menu and startup")
	fmt.Println("  installer.exe status    - Check installation status")
}

func install() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get uploader.exe path (same directory as installer)
	uploaderPath := filepath.Join(filepath.Dir(exePath), "uploader.exe")
	if _, err := os.Stat(uploaderPath); os.IsNotExist(err) {
		return fmt.Errorf("uploader.exe not found at %s", uploaderPath)
	}

	// Register context menu for all files
	if err := registerContextMenu(uploaderPath); err != nil {
		return err
	}

	// Register startup for mounter (optional)
	mounterPath := filepath.Join(filepath.Dir(exePath), "mounter.exe")
	if _, err := os.Stat(mounterPath); err == nil {
		if err := registerStartup(mounterPath); err != nil {
			fmt.Printf("Warning: Failed to register startup: %v\n", err)
		}
	}

	return nil
}

func uninstall() error {
	// Remove context menu
	if err := unregisterContextMenu(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	// Remove startup
	if err := unregisterStartup(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	return nil
}

func registerContextMenu(uploaderPath string) error {
	// Create shell key: HKEY_CLASSES_ROOT\*\shell\Upload2Cloud
	shellKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, shellKeyPath, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create shell key: %w", err)
	}
	defer shellKey.Close()

	// Set menu text
	if err := shellKey.SetStringValue("", menuText); err != nil {
		return fmt.Errorf("failed to set menu text: %w", err)
	}

	// Set icon
	if err := shellKey.SetStringValue("Icon", uploaderPath); err != nil {
		return fmt.Errorf("failed to set icon: %w", err)
	}

	// Create command key
	cmdKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, shellKeyPath+`\command`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create command key: %w", err)
	}
	defer cmdKey.Close()

	// Set command - %1 is the selected file path
	command := fmt.Sprintf(`"%s" "%%1"`, uploaderPath)
	if err := cmdKey.SetStringValue("", command); err != nil {
		return fmt.Errorf("failed to set command: %w", err)
	}

	return nil
}

func unregisterContextMenu() error {
	// Delete command subkey first
	err := registry.DeleteKey(registry.CLASSES_ROOT, shellKeyPath+`\command`)
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to delete command key: %w", err)
	}

	// Delete shell key
	err = registry.DeleteKey(registry.CLASSES_ROOT, shellKeyPath)
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to delete shell key: %w", err)
	}

	return nil
}

func registerStartup(mounterPath string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open Run key: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue("SimpleUploaderMounter", mounterPath); err != nil {
		return fmt.Errorf("failed to set startup value: %w", err)
	}

	return nil
}

func unregisterStartup() error {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		return nil // Key doesn't exist, nothing to remove
	}
	defer key.Close()

	err = key.DeleteValue("SimpleUploaderMounter")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to delete startup value: %w", err)
	}

	return nil
}

func checkStatus() {
	// Check context menu
	shellKey, err := registry.OpenKey(registry.CLASSES_ROOT, shellKeyPath, registry.QUERY_VALUE)
	if err != nil {
		fmt.Println("Context menu: NOT installed")
	} else {
		shellKey.Close()
		fmt.Println("Context menu: Installed")
	}

	// Check startup
	runKey, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err == nil {
		defer runKey.Close()
		_, _, err := runKey.GetStringValue("SimpleUploaderMounter")
		if err == nil {
			fmt.Println("Auto-start: Enabled")
		} else {
			fmt.Println("Auto-start: Disabled")
		}
	} else {
		fmt.Println("Auto-start: Disabled")
	}

	// Check WinFsp
	winfspKey, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\WinFsp`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Println("WinFsp: NOT installed (required for drive mount)")
		fmt.Println("  Download from: https://winfsp.dev/rel/")
	} else {
		winfspKey.Close()
		fmt.Println("WinFsp: Installed")
	}
}
