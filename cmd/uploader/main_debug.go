package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"simple-uploader/internal/config"
	"simple-uploader/internal/minio"
)

func main() {
	fmt.Println("=== Simple Uploader Debug ===")
	fmt.Printf("Args: %v\n", os.Args)

	if len(os.Args) < 2 {
		fmt.Println("ERROR: No files specified")
		fmt.Println("Usage: uploader.exe <file_path>")
		waitExit()
		return
	}

	// Load configuration
	fmt.Println("\n[1] Loading config...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("ERROR loading config: %v\n", err)

		// Try to show where it's looking
		exePath, _ := os.Executable()
		fmt.Printf("Executable path: %s\n", exePath)
		fmt.Printf("Looking for config at: %s\n", filepath.Join(filepath.Dir(exePath), "config.json"))
		waitExit()
		return
	}
	fmt.Printf("Config loaded: endpoint=%s, bucket=%s\n", cfg.MinIO.Endpoint, cfg.MinIO.Bucket)

	// Create MinIO client
	fmt.Println("\n[2] Connecting to MinIO...")
	client, err := minio.NewClient(&cfg.MinIO)
	if err != nil {
		fmt.Printf("ERROR connecting to MinIO: %v\n", err)
		waitExit()
		return
	}
	fmt.Println("MinIO client created")

	// Get file paths
	filePath := os.Args[1]
	fmt.Printf("\n[3] File to upload: %s\n", filePath)

	// Check file exists
	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("ERROR: File not found: %v\n", err)
		waitExit()
		return
	}
	if info.IsDir() {
		fmt.Println("ERROR: Path is a directory, not a file")
		waitExit()
		return
	}
	fmt.Printf("File size: %d bytes\n", info.Size())

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Ensure bucket exists
	fmt.Println("\n[4] Checking bucket...")
	if err := client.EnsureBucket(ctx); err != nil {
		fmt.Printf("ERROR with bucket: %v\n", err)
		waitExit()
		return
	}
	fmt.Println("Bucket OK")

	// Upload
	fmt.Println("\n[5] Uploading...")
	if err := client.UploadFile(ctx, filePath); err != nil {
		fmt.Printf("ERROR uploading: %v\n", err)
		waitExit()
		return
	}

	fmt.Println("\n=== SUCCESS! File uploaded ===")
	waitExit()
}

func waitExit() {
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}
