@echo off
echo Building 64-bit Windows executable...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o media-server-windows-amd64.exe
if errorlevel 1 (
    echo Failed to build 64-bit executable
    exit /b 1
)

echo Build complete. Executables:
echo  - media-server-windows-amd64.exe
pause
