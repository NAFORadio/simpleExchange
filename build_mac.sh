#!/bin/bash
# Build script for macOS

# Ensure output directory exists
mkdir -p build/macos

# Set macOS as target OS and architecture
export GOOS=darwin
export GOARCH=amd64

# Build the binary
echo "Building macOS binary..."
go build -o build/macos/messenger .

# Create basic app bundle structure
mkdir -p build/macos/Messenger.app/Contents/{MacOS,Resources}

# Create Info.plist
cat > build/macos/Messenger.app/Contents/Info.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>messenger</string>
    <key>CFBundleIdentifier</key>
    <string>com.example.messenger</string>
    <key>CFBundleName</key>
    <string>Messenger</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.10</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

# Copy binary to app bundle
cp build/macos/messenger build/macos/Messenger.app/Contents/MacOS/

# Create DMG (requires create-dmg tool)
if command -v create-dmg &> /dev/null; then
    create-dmg \
        --volname "Messenger" \
        --window-pos 200 120 \
        --window-size 800 400 \
        --icon-size 100 \
        --icon "Messenger.app" 200 190 \
        --hide-extension "Messenger.app" \
        --app-drop-link 600 185 \
        "build/macos/Messenger.dmg" \
        "build/macos/Messenger.app"
fi

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "macOS build completed successfully:"
    echo "  App bundle: build/macos/Messenger.app"
    if command -v create-dmg &> /dev/null; then
        echo "  DMG installer: build/macos/Messenger.dmg"
    fi
else
    echo "macOS build failed"
    exit 1
fi 