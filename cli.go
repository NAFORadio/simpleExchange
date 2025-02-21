package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
    "os"
    "strings"
    "time"
)

const (
    clearScreen = "\033[H\033[2J"
    clearLine   = "\033[2K"
    moveUp      = "\033[1A"
    moveToStart = "\033[0G"
)

func showSplashScreen() {
    splash := `
    ███╗   ██╗ █████╗ ███████╗ ██████╗     ██████╗  █████╗ ██████╗ ██╗ ██████╗ 
    ████╗  ██║██╔══██╗██╔════╝██╔═══██╗    ██╔══██╗██╔══██╗██╔══██╗██║██╔═══██╗
    ██╔██╗ ██║███████║█████╗  ██║   ██║    ██████╔╝███████║██║  ██║██║██║   ██║
    ██║╚██╗██║██╔══██║██╔══╝  ██║   ██║    ██╔══██╗██╔══██║██║  ██║██║██║   ██║
    ██║ ╚████║██║  ██║██║     ╚██████╔╝    ██║  ██║██║  ██║██████╔╝██║╚██████╔╝
    ╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝      ╚═════╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝ ╚═╝ ╚═════╝ 
                                                                                  
    Local Network Messenger v1.0.0
    Created by NAFO Radio - For Educational Purposes Only
    Released under MIT License - See LICENSE file for details
    
    DISCLAIMER: This software is provided AS IS without warranty.
    Users assume all responsibility for how this software is used.
    
    Type 'help' for available commands
    ================================================================================
    `
    fmt.Println(splash)
}

func validateCommand(input string) error {
    if input == "" {
        return fmt.Errorf("empty command")
    }

    if strings.HasPrefix(input, "send ") {
        if len(input) <= 5 {
            return fmt.Errorf("empty message")
        }
    }

    if strings.HasPrefix(input, "file ") {
        if len(input) <= 5 {
            return fmt.Errorf("no file path provided")
        }
        filepath := input[5:]
        if _, err := os.Stat(filepath); os.IsNotExist(err) {
            return fmt.Errorf("file does not exist: %s", filepath)
        }
    }

    return nil
}

func startCLI(messenger *Messenger) {
    showSplashScreen()
    fmt.Printf("Your ID: %s\n\n", messenger.ID)

    // Reserve a line for status
    fmt.Println(messenger.getNetworkStatus())
    
    // Start status updater in background
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        for range ticker.C {
            if !messenger.running {
                return
            }
            // Move up one line, clear it, and write new status
            fmt.Printf("%s%s%s%s\n", 
                moveUp,
                clearLine,
                moveToStart,
                messenger.getNetworkStatus())
        }
    }()

    fmt.Println("\nEnter command (type 'help' for available commands):")
    
    scanner := bufio.NewScanner(os.Stdin)
    messenger.running = true
    
    for scanner.Scan() {
        input := scanner.Text()
        
        // Validate input
        if err := validateCommand(input); err != nil {
            fmt.Printf("Error: %v\n", err)
            fmt.Print("\nEnter command: ")
            continue
        }

        // Clear the input line after command
        fmt.Print(clearLine + moveToStart)
        
        switch {
        case input == "help":
            printHelp()
        
        case input == "list":
            listPeers(messenger)
        
        case input == "status":
            handleStatusCommand(messenger)
        
        case input == "quit":
            fmt.Println("Shutting down...")
            messenger.running = false
            return
        
        case strings.HasPrefix(input, "send "):
            handleSendCommand(messenger, input[5:])
        
        case strings.HasPrefix(input, "file "):
            handleFileCommand(messenger, input[5:])
        }
        
        fmt.Print("\nEnter command: ")
    }
}

func printHelp() {
    fmt.Println("\nAvailable commands:")
    fmt.Println("  help           - Show this help")
    fmt.Println("  list           - List connected peers")
    fmt.Println("  send <message> - Send text message")
    fmt.Println("  file <path>    - Send file")
    fmt.Println("  status         - Show network and statistics")
    fmt.Println("  quit           - Exit the application")
    fmt.Println()
}

func listPeers(messenger *Messenger) {
    fmt.Println("\nConnected peers:")
    messenger.peersMutex.RLock()
    defer messenger.peersMutex.RUnlock()
    
    for _, peer := range messenger.peers {
        fmt.Printf("  %s (%s) - Last seen: %s\n", 
            peer.ID, peer.Address, peer.LastSeen.Format("15:04:05"))
    }
    fmt.Println()
}

func handleSendCommand(messenger *Messenger, message string) {
    // Create message
    msg := Message{
        Type:      "text",
        Content:   message,
        Timestamp: time.Now(),
        SenderID:  messenger.ID,
        Size:      int64(len(message)),
    }

    // Broadcast to all peers except self
    messenger.peersMutex.RLock()
    peerCount := 0
    for _, peer := range messenger.peers {
        if peer.ID != messenger.ID {
            // Send message to peer
            if err := messenger.sendToPeer(peer, msg); err != nil {
                fmt.Printf("\nError sending to %s: %v\n", peer.ID, err)
                continue
            }
            peerCount++
        }
    }
    messenger.peersMutex.RUnlock()

    if peerCount == 0 {
        // No peers available, queue the message
        messenger.queueMessage(msg)
        fmt.Printf("\n%sNo peers available. Message queued for retry%s\n\nEnter command: ",
            clearLine, moveToStart)
        return
    }

    // Update statistics
    messenger.updateStats(msg, true)

    // Clear line and show status
    fmt.Printf("\n%sMessage sent to %d peers%s\n\nEnter command: ", 
        clearLine, peerCount, moveToStart)
}

func handleFileCommand(messenger *Messenger, filepath string) {
    // Check if file exists and memory is sufficient
    if err := messenger.handleLargeFile(filepath); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    // Read file
    data, err := os.ReadFile(filepath)
    if err != nil {
        fmt.Printf("Error reading file: %v\n", err)
        return
    }

    // Create message
    msg := Message{
        Type:      "file",
        Content:   filepath,
        Data:      data,
        Timestamp: time.Now(),
        SenderID:  messenger.ID,
        Size:      int64(len(data)),
    }

    // Broadcast to all peers except self
    messenger.peersMutex.RLock()
    peerCount := 0
    for _, peer := range messenger.peers {
        if peer.ID != messenger.ID {
            fmt.Printf("Sending file to peer %s...\n", peer.ID)
            if err := messenger.sendToPeer(peer, msg); err != nil {
                fmt.Printf("Error sending to %s: %v\n", peer.ID, err)
                continue
            }
            peerCount++
        }
    }
    messenger.peersMutex.RUnlock()

    if peerCount == 0 {
        // No peers available, queue the message
        messenger.queueMessage(msg)
        fmt.Printf("\n%sNo peers available. File queued for retry%s\n\nEnter command: ",
            clearLine, moveToStart)
        return
    }

    messenger.updateStats(msg, true)
    fmt.Printf("\n%sFile sent to %d peers%s\n\nEnter command: ", 
        clearLine, peerCount, moveToStart)
}

func handleStatusCommand(messenger *Messenger) {
    fmt.Print(clearScreen)  // Clear screen before showing full status
    fmt.Println("=== Status Report ===")
    fmt.Println(messenger.getNetworkStatus())
    fmt.Println(messenger.getStatistics())
    fmt.Println("Encryption: Enabled (AES-GCM)")
    fmt.Println("=====================================")
    fmt.Print("\nPress Enter to continue...")
    bufio.NewReader(os.Stdin).ReadString('\n')
    
    // Redraw the normal interface
    fmt.Print(clearScreen)
    showSplashScreen()
    fmt.Printf("Your ID: %s\n\n", messenger.ID)
    fmt.Println(messenger.getNetworkStatus())
    fmt.Print("\nEnter command: ")
}

func (m *Messenger) sendToPeer(peer *Peer, msg Message) error {
    // Encrypt the message
    data, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %v", err)
    }

    encrypted, err := m.encrypt(data)
    if err != nil {
        return fmt.Errorf("failed to encrypt message: %v", err)
    }

    // Create UDP connection to peer
    addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", peer.Address, messagePort))
    if err != nil {
        return fmt.Errorf("failed to resolve peer address: %v", err)
    }

    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        return fmt.Errorf("failed to connect to peer: %v", err)
    }
    defer conn.Close()

    // Send the encrypted message
    _, err = conn.Write(encrypted)
    if err != nil {
        return fmt.Errorf("failed to send message: %v", err)
    }

    return nil
} 