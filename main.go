package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "path/filepath"
    "sync"
    "time"
    "container/list"
)

const (
    discoveryPort = 35001
    messagePort   = 35002
    maxFileSize   = 6 * 1024 * 1024 * 1024 // 6GB limit
)

type Message struct {
    Type      string    `json:"type"`      // "text" or "file"
    Content   string    `json:"content"`    // text content or file name
    Data      []byte    `json:"data"`      // file data if type is "file"
    Timestamp time.Time `json:"timestamp"`
    SenderID  string    `json:"sender_id"`
    Size      int64     `json:"size"`      // size in bytes for statistics
}

type Peer struct {
    ID        string
    Address   string
    LastSeen  time.Time
    Connected bool
}

type Statistics struct {
    BytesSent     int64
    BytesReceived int64
    MessagesSent  int64
    MessagesRecvd int64
    FilesSent     int64
    FilesRecvd    int64
    StartTime     time.Time
    mutex         sync.RWMutex
}

type QueuedMessage struct {
    Message   Message
    Attempts  int
    LastTry   time.Time
}

type Messenger struct {
    ID            string
    peers         map[string]*Peer
    peersMutex    sync.RWMutex
    encryptionKey []byte
    stats         Statistics
    running       bool
    shutdown      chan struct{}
    messageQueue  *list.List
    queueMutex   sync.RWMutex
}

func NewMessenger() *Messenger {
    // Generate random ID for this instance
    id := make([]byte, 8)
    rand.Read(id)
    
    // Generate encryption key
    key := make([]byte, 32)
    rand.Read(key)
    
    m := &Messenger{
        ID:            fmt.Sprintf("%x", id),
        peers:         make(map[string]*Peer),
        encryptionKey: key,
        shutdown:      make(chan struct{}),
        running:       true,
        messageQueue:  list.New(),
    }
    m.stats.StartTime = time.Now()
    
    // Start queue processor
    go m.processMessageQueue()
    
    return m
}

func (m *Messenger) startDiscovery() {
    addr := &net.UDPAddr{Port: discoveryPort}
    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Broadcast presence periodically
    go func() {
        for {
            m.broadcast(conn)
            time.Sleep(5 * time.Second)
        }
    }()

    // Listen for other peers
    buffer := make([]byte, 1024)
    for {
        n, remoteAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            continue
        }

        var peer Peer
        if err := json.Unmarshal(buffer[:n], &peer); err != nil {
            continue
        }

        m.peersMutex.Lock()
        peer.Address = remoteAddr.IP.String()
        peer.LastSeen = time.Now()
        m.peers[peer.ID] = &peer
        m.peersMutex.Unlock()
    }
}

func (m *Messenger) broadcast(conn *net.UDPConn) {
    peer := Peer{
        ID:        m.ID,
        LastSeen:  time.Now(),
        Connected: true,
    }

    data, err := json.Marshal(peer)
    if err != nil {
        return
    }

    addr := &net.UDPAddr{
        IP:   net.IPv4(255, 255, 255, 255),
        Port: discoveryPort,
    }
    conn.WriteToUDP(data, addr)
}

func (m *Messenger) encrypt(data []byte) ([]byte, error) {
    block, err := aes.NewCipher(m.encryptionKey)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    return gcm.Seal(nonce, nonce, data, nil), nil
}

func (m *Messenger) decrypt(data []byte) ([]byte, error) {
    block, err := aes.NewCipher(m.encryptionKey)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}

func verifyMemoryForFileSize(fileSize int64) error {
    availMem, err := checkAvailableMemory()
    if err != nil {
        return fmt.Errorf("unable to check memory: %v", err)
    }

    // Require 1.5x the file size to account for encryption overhead and processing
    requiredMem := uint64(float64(fileSize) * 1.5)
    if requiredMem > availMem {
        return fmt.Errorf("insufficient memory: need %d bytes, have %d bytes available",
            requiredMem, availMem)
    }
    return nil
}

func (m *Messenger) handleLargeFile(filePath string) error {
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return fmt.Errorf("unable to stat file: %v", err)
    }

    if fileInfo.Size() > maxFileSize {
        return fmt.Errorf("file too large: %d bytes (max: %d)", fileInfo.Size(), maxFileSize)
    }

    if err := verifyMemoryForFileSize(fileInfo.Size()); err != nil {
        return err
    }

    // If we get here, we have sufficient memory to process the file
    return nil
}

func (m *Messenger) updateStats(msg Message, sent bool) {
    m.stats.mutex.Lock()
    defer m.stats.mutex.Unlock()

    if sent {
        m.stats.BytesSent += msg.Size
        if msg.Type == "file" {
            m.stats.FilesSent++
        } else {
            m.stats.MessagesSent++
        }
    } else {
        m.stats.BytesReceived += msg.Size
        if msg.Type == "file" {
            m.stats.FilesRecvd++
        } else {
            m.stats.MessagesRecvd++
        }
    }
}

func (m *Messenger) getNetworkStatus() string {
    m.peersMutex.RLock()
    peerCount := len(m.peers)
    var activeCount int
    for _, p := range m.peers {
        if time.Since(p.LastSeen) < 10*time.Second {
            activeCount++
        }
    }
    m.peersMutex.RUnlock()
    return fmt.Sprintf("Network Status: %d peers (%d active)", peerCount, activeCount)
}

func (m *Messenger) getStatistics() string {
    m.stats.mutex.RLock()
    defer m.stats.mutex.RUnlock()
    
    uptime := time.Since(m.stats.StartTime).Round(time.Second)
    return fmt.Sprintf(`
Statistics:
  Uptime: %s
  Messages: Sent=%d, Received=%d
  Files: Sent=%d, Received=%d
  Data: Sent=%s, Received=%s`,
        uptime,
        m.stats.MessagesSent, m.stats.MessagesRecvd,
        m.stats.FilesSent, m.stats.FilesRecvd,
        formatBytes(m.stats.BytesSent), formatBytes(m.stats.BytesReceived))
}

func formatBytes(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (m *Messenger) Cleanup() {
    close(m.shutdown)
    m.running = false
    
    // Clean up peers
    m.peersMutex.Lock()
    for id, peer := range m.peers {
        peer.Connected = false
        delete(m.peers, id)
    }
    m.peersMutex.Unlock()
}

func (m *Messenger) startMessageListener() {
    addr := &net.UDPAddr{Port: messagePort}
    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create error channel for goroutine
    errChan := make(chan error, 1)
    
    // Start listener in goroutine
    go func() {
        buffer := make([]byte, maxFileSize)
        for {
            select {
            case <-m.shutdown:
                return
            default:
                conn.SetReadDeadline(time.Now().Add(time.Second))
                n, _, err := conn.ReadFromUDP(buffer)
                if err != nil {
                    if !os.IsTimeout(err) {
                        errChan <- err
                    }
                    continue
                }

                // Process message
                if err := m.handleMessage(buffer[:n]); err != nil {
                    log.Printf("Error handling message: %v", err)
                }
            }
        }
    }()

    // Monitor for errors
    select {
    case err := <-errChan:
        log.Printf("Message listener error: %v", err)
    case <-m.shutdown:
        return
    }
}

func (m *Messenger) handleMessage(data []byte) error {
    // Decrypt and handle message
    decrypted, err := m.decrypt(data)
    if err != nil {
        return fmt.Errorf("failed to decrypt: %v", err)
    }

    var msg Message
    if err := json.Unmarshal(decrypted, &msg); err != nil {
        return fmt.Errorf("failed to unmarshal: %v", err)
    }

    // Update statistics
    m.updateStats(msg, false)

    // Handle based on message type
    switch msg.Type {
    case "text":
        fmt.Printf("\n%sReceived from %s: %s%s\nEnter command: ", 
            clearLine, msg.SenderID, msg.Content, moveToStart)
    
    case "file":
        // Create received files directory if it doesn't exist
        if err := os.MkdirAll("received_files", 0755); err != nil {
            return fmt.Errorf("failed to create received_files directory: %v", err)
        }

        // Save file with unique name
        filename := filepath.Base(msg.Content)
        savePath := filepath.Join("received_files", 
            fmt.Sprintf("%s_%s", msg.SenderID, filename))
        
        if err := os.WriteFile(savePath, msg.Data, 0644); err != nil {
            return fmt.Errorf("failed to save file: %v", err)
        }

        fmt.Printf("\n%sReceived file from %s: %s%s\nEnter command: ",
            clearLine, msg.SenderID, savePath, moveToStart)
    }

    return nil
}

func (m *Messenger) processMessageQueue() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-m.shutdown:
            return
        case <-ticker.C:
            m.retryQueuedMessages()
        }
    }
}

func (m *Messenger) retryQueuedMessages() {
    m.queueMutex.Lock()
    defer m.queueMutex.Unlock()

    for e := m.messageQueue.Front(); e != nil; {
        qm := e.Value.(*QueuedMessage)
        next := e.Next() // Store next before potential removal

        // Skip if not enough time has passed since last attempt
        if time.Since(qm.LastTry) < time.Second*5 {
            e = next
            continue
        }

        // Try to send the message
        sent := false
        m.peersMutex.RLock()
        for _, peer := range m.peers {
            if peer.ID != m.ID && time.Since(peer.LastSeen) < time.Second*10 {
                if err := m.sendToPeer(peer, qm.Message); err == nil {
                    sent = true
                }
            }
        }
        m.peersMutex.RUnlock()

        if sent {
            // Message sent successfully, remove from queue
            m.messageQueue.Remove(e)
            log.Printf("Successfully sent queued %s after %d attempts\n", 
                qm.Message.Type, qm.Attempts)
        } else {
            // Update attempt count and last try time
            qm.Attempts++
            qm.LastTry = time.Now()
            
            // Remove if too many attempts
            if qm.Attempts > 12 { // 1 minute of retries
                m.messageQueue.Remove(e)
                log.Printf("Dropping %s message after %d failed attempts\n", 
                    qm.Message.Type, qm.Attempts)
            }
        }

        e = next
    }
}

func (m *Messenger) queueMessage(msg Message) {
    m.queueMutex.Lock()
    defer m.queueMutex.Unlock()

    qm := &QueuedMessage{
        Message:  msg,
        Attempts: 0,
        LastTry:  time.Now(),
    }
    m.messageQueue.PushBack(qm)
}

func main() {
    var guiMode bool
    flag.BoolVar(&guiMode, "gui", false, "Start in GUI mode")
    flag.Parse()

    messenger := NewMessenger()
    go messenger.startDiscovery()
    go messenger.startMessageListener()

    // For now, always use CLI mode
    startCLI(messenger)
} 