#!/bin/bash
# Build script for Linux

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
BUILD_DIR="/tmp/messenger-build"
echo "Using build directory: ${BUILD_DIR}"

# Clean up any existing build artifacts
rm -rf "${BUILD_DIR}"
rm -rf build/linux

# Create build directories with correct permissions
mkdir -p "${BUILD_DIR}/linux"
mkdir -p build/linux
chmod 755 "${BUILD_DIR}/linux"

# Set Linux as target OS and architecture
export GOOS=linux
export GOARCH=amd64

# Build the binary
echo "Building Linux binary..."
go build -o "${BUILD_DIR}/linux/messenger" .
chmod 755 "${BUILD_DIR}/linux/messenger"

# Create .deb package structure
echo "Creating package structure..."
mkdir -p "${BUILD_DIR}/linux/deb/DEBIAN"
mkdir -p "${BUILD_DIR}/linux/deb/usr/local/bin"

# Set ownership and permissions
chown -R root:root "${BUILD_DIR}/linux/deb"
find "${BUILD_DIR}/linux/deb" -type d -exec chmod 755 {} \;
find "${BUILD_DIR}/linux/deb" -type f -exec chmod 644 {} \;

# Create control file for .deb package
cat > "${BUILD_DIR}/linux/deb/DEBIAN/control" << EOF
Package: messenger
Version: 1.0.0
Section: base
Priority: optional
Architecture: amd64
Maintainer: NAFO Radio <example@example.com>
Description: Local Network Messenger
 A secure messaging system for local network communication.
EOF

# Copy binary and set its permissions
cp "${BUILD_DIR}/linux/messenger" "${BUILD_DIR}/linux/deb/usr/local/bin/"
chmod 755 "${BUILD_DIR}/linux/deb/usr/local/bin/messenger"

# Build .deb package
echo "Building .deb package..."
dpkg-deb --build "${BUILD_DIR}/linux/deb" "${BUILD_DIR}/linux/messenger.deb"

# Copy results back to the project directory
if [ $? -eq 0 ]; then
    cp "${BUILD_DIR}/linux/messenger" "build/linux/"
    cp "${BUILD_DIR}/linux/messenger.deb" "build/linux/"
    echo "Linux build completed successfully:"
    echo "  Binary: build/linux/messenger"
    echo "  Debian package: build/linux/messenger.deb"
else
    echo "Linux build failed"
    exit 1
fi

# Clean up
rm -rf "${BUILD_DIR}" 