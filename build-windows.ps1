# build.ps1 - Build Windows 64-bit and 32-bit binaries for media-server

Write-Host "Building 64-bit Windows executable..."
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "-s -w" -o "media-server-windows-amd64.exe"
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build 64-bit executable"
    exit $LASTEXITCODE
}

Write-Host "Building 32-bit Windows executable..."
$env:GOOS = "windows"
$env:GOARCH = "386"
go build -ldflags "-s -w" -o "media-server-windows-386.exe"
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build 32-bit executable"
    exit $LASTEXITCODE
}

Write-Host "Build complete. Executables:"
Write-Host " - media-server-windows-amd64.exe"
Write-Host " - media-server-windows-386.exe"
