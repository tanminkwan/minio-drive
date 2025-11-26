@echo off
cd /d C:\GitHub\simple-uploader
echo Building debug mounter...
go build -o dist\mounter_debug.exe .\cmd\mounter_debug
echo Done!
pause
