package main

import (
	"fmt"
	"os"
	"time"

	"simple-uploader/internal/config"
	"simple-uploader/internal/icon"
	"simple-uploader/internal/rclone"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

var (
	cfg     *config.Config
	manager *rclone.Manager
)

func main() {
	var err error

	// Load configuration
	cfg, err = config.Load()
	if err != nil {
		showError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	// Create rclone manager
	manager, err = rclone.NewManager(cfg)
	if err != nil {
		showError(fmt.Sprintf("Failed to initialize: %v", err))
		os.Exit(1)
	}

	// Run system tray
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.Data)

	// Set title based on mount type
	if cfg.IsWinFsp() {
		systray.SetTitle("MinIO Mount")
		systray.SetTooltip("MinIO Cloud Drive (WinFsp)")
	} else {
		systray.SetTitle("MinIO WebDAV")
		systray.SetTooltip("MinIO Cloud WebDAV Server")
	}

	// Menu items
	mStart := systray.AddMenuItem("Start", "Start server/mount")
	mStop := systray.AddMenuItem("Stop", "Stop server/mount")
	mStop.Disable()

	systray.AddSeparator()

	mStatus := systray.AddMenuItem("Status: Stopped", "Current status")
	mStatus.Disable()

	mInfo := systray.AddMenuItem("", "Info")
	mInfo.Disable()
	mInfo.Hide()

	systray.AddSeparator()

	// Show mount type
	mountType := "WebDAV"
	if cfg.IsWinFsp() {
		mountType = "WinFsp"
	}
	mType := systray.AddMenuItem(fmt.Sprintf("Mode: %s", mountType), "Mount type")
	mType.Disable()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Auto-start if configured
	if cfg.Mount.AutoStart {
		go func() {
			if err := startMount(mStart, mStop, mStatus, mInfo); err != nil {
				showError(fmt.Sprintf("Auto-start failed: %v", err))
			}
		}()
	}

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mStart.ClickedCh:
				if err := startMount(mStart, mStop, mStatus, mInfo); err != nil {
					showError(fmt.Sprintf("Start failed: %v", err))
				}
			case <-mStop.ClickedCh:
				if err := stopMount(mStart, mStop, mStatus, mInfo); err != nil {
					showError(fmt.Sprintf("Stop failed: %v", err))
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	if manager != nil {
		if cfg.IsWinFsp() {
			_ = manager.UnmountWinFsp()
		} else {
			_ = manager.DisconnectDrive()
			_ = manager.StopWebDAV()
		}
	}
}

func startMount(mStart, mStop, mStatus, mInfo *systray.MenuItem) error {
	if cfg.IsWinFsp() {
		// WinFsp mount
		if err := manager.MountWinFsp(); err != nil {
			return err
		}

		// Wait for mount
		time.Sleep(2 * time.Second)

		mStart.Disable()
		mStop.Enable()
		mStatus.SetTitle(fmt.Sprintf("Status: Mounted (%s)", manager.GetDriveLetter()))
		mInfo.SetTitle(fmt.Sprintf("Drive: %s", manager.GetDriveLetter()))
		mInfo.Show()

		_ = beeep.Notify("MinIO Mount",
			fmt.Sprintf("Mounted to %s", manager.GetDriveLetter()), "")
	} else {
		// WebDAV
		if err := manager.StartWebDAV(); err != nil {
			return err
		}

		// Wait for server to start
		time.Sleep(500 * time.Millisecond)

		// Connect network drive
		if err := manager.ConnectDrive(); err != nil {
			_ = beeep.Alert("MinIO WebDAV", fmt.Sprintf("Server started but drive connection failed: %v", err), "")
		}

		mStart.Disable()
		mStop.Enable()
		mStatus.SetTitle(fmt.Sprintf("Status: Running (%s)", manager.GetDriveLetter()))
		mInfo.SetTitle(fmt.Sprintf("Drive: %s", manager.GetDriveLetter()))
		mInfo.Show()

		_ = beeep.Notify("MinIO WebDAV",
			fmt.Sprintf("Connected to %s", manager.GetDriveLetter()), "")
	}

	return nil
}

func stopMount(mStart, mStop, mStatus, mInfo *systray.MenuItem) error {
	if cfg.IsWinFsp() {
		// WinFsp unmount
		if err := manager.UnmountWinFsp(); err != nil {
			return err
		}

		_ = beeep.Notify("MinIO Mount", "Unmounted", "")
	} else {
		// WebDAV
		_ = manager.DisconnectDrive()

		if err := manager.StopWebDAV(); err != nil {
			return err
		}

		_ = beeep.Notify("MinIO WebDAV", "Disconnected", "")
	}

	mStart.Enable()
	mStop.Disable()
	mStatus.SetTitle("Status: Stopped")
	mInfo.Hide()

	return nil
}

func showError(msg string) {
	_ = beeep.Alert("MinIO Error", msg, "")
}
