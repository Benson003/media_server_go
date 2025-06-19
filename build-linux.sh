#!/bin/bash
set -e

echo "Building media-server..."

echo "Building for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o media-server-linux-amd64
echo "Built media-server-linux-amd64"


echo "Building for Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o media-server-windows-amd64.exe
echo "Built media-server-windows-amd64.exe"


echo "All builds completed successfully."
