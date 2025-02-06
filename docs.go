// Package documentation and examples
package main

/*
NAFO Radio Local Network Messenger
================================

This file contains package documentation and usage examples.
See BUILD.md and README.md for detailed documentation.

Basic Usage:
    messenger --gui     Start in GUI mode (if available)
    messenger          Start in CLI mode

Example CLI Session:
    > help
    Available commands:
      help           - Show this help
      list           - List connected peers
      send <message> - Send text message
      file <path>    - Send file
      status         - Show network and statistics
      quit           - Exit the application

    > send "Hello World"
    Message sent to all peers

    > file "/path/to/file.txt"
    Sending file to all peers...

For full documentation, please refer to the README.md file.
*/ 