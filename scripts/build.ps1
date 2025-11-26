# Build script for Simple Uploader
# Run from project root: .\scripts\build.ps1

$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$OutputDir = Join-Path $ProjectRoot "dist"

Write-Host "Building Simple Uploader..." -ForegroundColor Cyan
Write-Host "Output directory: $OutputDir" -ForegroundColor Gray

# Create output directory
if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Path $OutputDir | Out-Null

# Change to project root
Push-Location $ProjectRoot

try {
    # Download dependencies
    Write-Host "`nDownloading dependencies..." -ForegroundColor Yellow
    go mod tidy

    # Build uploader (GUI mode - no console window)
    Write-Host "`nBuilding uploader.exe..." -ForegroundColor Yellow
    go build -ldflags="-H windowsgui -s -w" -o "$OutputDir\uploader.exe" .\cmd\uploader

    # Build mounter (GUI mode - no console window)
    Write-Host "Building mounter.exe..." -ForegroundColor Yellow
    go build -ldflags="-H windowsgui -s -w" -o "$OutputDir\mounter.exe" .\cmd\mounter

    # Build installer (console mode for output)
    Write-Host "Building installer.exe..." -ForegroundColor Yellow
    go build -ldflags="-s -w" -o "$OutputDir\installer.exe" .\cmd\installer

    # Copy config template
    Write-Host "`nCopying config template..." -ForegroundColor Yellow
    Copy-Item "config.json" "$OutputDir\config.json"

    # Create README
    $ReadmeContent = @"
Simple Uploader - MinIO Cloud Integration
==========================================

INSTALLATION
------------
1. Install WinFsp from https://winfsp.dev/rel/ (required for drive mount)
2. Download rclone.exe from https://rclone.org/downloads/ and place it in this folder
3. Edit config.json with your MinIO settings
4. Run: installer.exe install (as Administrator)

USAGE
-----
- Right-click any file in Explorer -> "Upload to Cloud"
- mounter.exe runs in system tray to mount/unmount the drive

UNINSTALL
---------
Run: installer.exe uninstall (as Administrator)

CONFIGURATION (config.json)
---------------------------
{
  "minio": {
    "endpoint": "your-minio-server.com",
    "access_key": "YOUR_ACCESS_KEY",
    "secret_key": "YOUR_SECRET_KEY",
    "bucket": "your-bucket",
    "use_ssl": true
  },
  "mount": {
    "drive_letter": "M",
    "auto_mount": true
  }
}

FILES
-----
- uploader.exe  : Handles file uploads (called from context menu)
- mounter.exe   : System tray app for drive mounting
- installer.exe : Install/uninstall context menu and startup
- rclone.exe    : Required for drive mounting (download separately)
- config.json   : Your MinIO configuration
"@
    Set-Content -Path "$OutputDir\README.txt" -Value $ReadmeContent

    Write-Host "`n=== Build Complete ===" -ForegroundColor Green
    Write-Host "Output files in: $OutputDir" -ForegroundColor Cyan
    Write-Host "`nNext steps:" -ForegroundColor Yellow
    Write-Host "1. Download rclone.exe and place in dist folder"
    Write-Host "2. Edit dist\config.json with your MinIO settings"
    Write-Host "3. Run dist\installer.exe install as Administrator"

} finally {
    Pop-Location
}
