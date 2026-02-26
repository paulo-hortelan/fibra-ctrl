package connection

import (
    "errors"
    "net"
    "sync"
    "time"
)

// Protocol represents the connection protocol
type Protocol string

const (
    // ProtocolTelnet represents a Telnet connection
    ProtocolTelnet Protocol = "telnet"
    
    // ProtocolTL1 represents a TL1 connection
    ProtocolTL1 Protocol = "tl1"
)

// Connection represents a connection to an OLT
type Connection struct {
    ID       string
    Address  string
    Protocol Protocol
    Conn     net.Conn
    mu       sync.Mutex
    inUse    bool
}

// NewConnection creates a new connection
func NewConnection(address string, protocol Protocol) (*Connection, error) {
    // Create a connection based on the protocol
    var conn net.Conn
    var err error
    
    conn, err = net.DialTimeout("tcp", address, 10*time.Second)
    if err != nil {
        return nil, err
    }
    
    return &Connection{
        ID:       generateID(address),
        Address:  address,
        Protocol: protocol,
        Conn:     conn,
        inUse:    false,
    }, nil
}

// Execute executes a command on the connection
func (c *Connection) Execute(command string) (string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.inUse {
        return "", errors.New("connection is in use")
    }
    
    c.inUse = true
    defer func() { c.inUse = false }()
    
    // Set read/write deadlines
    err := c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
    if err != nil {
        return "", err
    }
    
    // Write the command to the connection
    _, err = c.Conn.Write([]byte(command + "\n"))
    if err != nil {
        return "", err
    }
    
    // Read the response
    buffer := make([]byte, 4096)
    n, err := c.Conn.Read(buffer)
    if err != nil {
        return "", err
    }
    
    return string(buffer[:n]), nil
}

// Close closes the connection
func (c *Connection) Close() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    return c.Conn.Close()
}

// generateID generates a unique ID for a connection
func generateID(address string) string {
    return address + "-" + time.Now().Format("20060102150405")
}