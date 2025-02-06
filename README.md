# NAFO Radio Local Network Messenger

A secure, cross-platform peer-to-peer messaging system for local network communication.

## Features

### Security
- End-to-end AES-GCM encryption
- No external servers or cloud dependencies
- Peer-to-peer architecture
- No data persistence

### Communication
- Real-time text messaging
- File transfers up to 6GB
- Automatic peer discovery
- Network status monitoring

### Cross-Platform Support
- Windows (.exe)
- Linux (.deb)
- macOS (.app)

## Quick Start

### Installation

#### Windows
```
# Run the Windows build script
./build_windows.sh
# Output: build/windows/messenger.exe
```

#### Linux
```
# Run the Linux build script
./build_linux.sh
# Install the .deb package
sudo dpkg -i build/linux/messenger.deb
```

#### macOS
```
# Run the macOS build script
./build_mac.sh
# Install the .app bundle
cp -r build/macos/Messenger.app /Applications/
```

### Usage

Start the messenger:
```
messenger     # CLI mode
```

Available commands:
```
help           - Show available commands
list           - List connected peers
send <message> - Send text message
file <path>    - Send file
status         - Show network and statistics
quit           - Exit application
```

## Network Requirements

- UDP ports required:
  - 35001 (peer discovery)
  - 35002 (messaging)
- Local network with UDP broadcast enabled
- Firewall rules allowing application traffic

## System Requirements

- Memory: 8GB RAM recommended
- Storage: Space for message/file handling
- Network: Local network access
- Permissions: Network and file system access

## Security Notes

- All communications encrypted with AES-GCM
- No authentication (designed for trusted networks)
- No persistent storage of messages or files
- Local network only, no internet required

## Limitations

- Maximum file size: 6GB
- Local network only
- No message persistence
- No user authentication
- No offline messaging

## Troubleshooting

### Common Issues

1. No peers visible
   - Check network connectivity
   - Verify UDP broadcast is enabled
   - Check firewall settings

2. Message send failure
   - Verify peer is still connected
   - Check network connectivity
   - Ensure sufficient memory

3. File transfer issues
   - Check available memory
   - Verify file permissions
   - Ensure file size < 6GB

## License

MIT License - See LICENSE file for details.

## Disclaimer

This software is provided for EDUCATIONAL PURPOSES ONLY. The creators and contributors:
- Make no warranties about functionality or suitability
- Are not responsible for any consequences of use
- Provide this as a learning tool only

## Created By

NAFO Radio - For Educational Purposes Only 