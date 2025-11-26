package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"simple-uploader/internal/config"
	"simple-uploader/internal/minio"

	"github.com/gen2brain/beeep"
)

func main() {
	if len(os.Args) < 2 {
		showNotification("Upload Error", "No files specified")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		showNotification("Upload Error", fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	// Create MinIO client
	client, err := minio.NewClient(&cfg.MinIO)
	if err != nil {
		showNotification("Upload Error", fmt.Sprintf("Failed to connect: %v", err))
		os.Exit(1)
	}

	// Get file paths from arguments
	filePaths := os.Args[1:]

	// Validate files exist
	var validFiles []string
	for _, path := range filePaths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			validFiles = append(validFiles, path)
		}
	}

	if len(validFiles) == 0 {
		showNotification("Upload Error", "No valid files to upload")
		os.Exit(1)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Ensure bucket exists
	if err := client.EnsureBucket(ctx); err != nil {
		showNotification("Upload Error", fmt.Sprintf("Bucket error: %v", err))
		os.Exit(1)
	}

	// Upload files
	successes, failures := client.UploadFiles(ctx, validFiles)

	// Show result notification
	if cfg.Upload.ShowNotification {
		showResult(successes, failures)
	}

	if len(failures) > 0 {
		os.Exit(1)
	}
}

func showNotification(title, message string) {
	_ = beeep.Notify(title, message, "")
}

func showResult(successes []string, failures map[string]error) {
	if len(failures) == 0 {
		// All successful
		if len(successes) == 1 {
			showNotification("Upload Complete", fmt.Sprintf("Uploaded: %s", filepath.Base(successes[0])))
		} else {
			showNotification("Upload Complete", fmt.Sprintf("Uploaded %d files successfully", len(successes)))
		}
	} else if len(successes) == 0 {
		// All failed
		var errMsgs []string
		for path, err := range failures {
			errMsgs = append(errMsgs, fmt.Sprintf("%s: %v", filepath.Base(path), err))
		}
		showNotification("Upload Failed", strings.Join(errMsgs, "\n"))
	} else {
		// Partial success
		showNotification("Upload Partial",
			fmt.Sprintf("%d succeeded, %d failed", len(successes), len(failures)))
	}
}
