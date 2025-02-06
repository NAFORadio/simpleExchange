#!/bin/bash
# Build script for Windows

# Initialize Go module if it doesn't exist
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module..."
    go mod init messenger
    if [ $? -ne 0 ]; then
        echo "Failed to initialize Go module"
        exit 1
    fi
fi

# Ensure output directory exists
mkdir -p build/windows

# Set Windows as target OS and architecture
export GOOS=windows
export GOARCH=amd64

# Build the binary
echo "Building Windows binary..."
go build -o build/windows/messenger.exe .

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Windows build completed successfully: build/windows/messenger.exe"
else
    echo "Windows build failed"
    exit 1
fi 