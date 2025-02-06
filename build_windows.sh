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

# Use /tmp for building to avoid Windows filesystem permission issues
BUILD_DIR="/tmp/messenger-build-windows"
echo "Using build directory: ${BUILD_DIR}"

# Clean up any existing build artifacts
rm -rf "${BUILD_DIR}"
rm -rf build/windows

# Create build directories
mkdir -p "${BUILD_DIR}"
mkdir -p build/windows

# Set Windows as target OS and architecture
export GOOS=windows
export GOARCH=amd64

# Build the binary with console window
echo "Building Windows binary..."
go build -ldflags "-H windowsgui" -o "${BUILD_DIR}/messenger.exe" .

# Set permissions in Linux filesystem
chmod 755 "${BUILD_DIR}/messenger.exe"

# Copy to final location with proper permissions
cp "${BUILD_DIR}/messenger.exe" build/windows/

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Windows build completed successfully:"
    echo "  Binary: build/windows/messenger.exe"
    echo "Note: Run messenger.exe from command prompt for best experience"
else
    echo "Windows build failed"
    exit 1
fi

# Clean up
rm -rf "${BUILD_DIR}" 