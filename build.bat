@echo off
cd /d C:\GitHub\simple-uploader
echo Running go mod tidy...
go mod tidy
echo Building mounter.exe...
go build -ldflags="-H windowsgui -s -w" -o dist\mounter.exe .\cmd\mounter
echo Build complete!
dir dist\*.exe
pause
