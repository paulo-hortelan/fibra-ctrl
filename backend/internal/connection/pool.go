package connection

import (
    "errors"
    "sync"
)

// Pool represents a pool of connections
type Pool struct {
    connections map[string][]*Connection
    mu          sync.RWMutex
}

// NewPool creates a new connection pool
func NewPool() *Pool {
    return &Pool{
        connections: make(map[string][]*Connection),
    }
}

// GetConnection gets a connection from the pool, creating one if necessary
func (p *Pool) GetConnection(address string, protocol Protocol) (*Connection, error) {
    p.mu.RLock()
    
    // Check if we have any connections for this address
    if conns, ok := p.connections[address]; ok {
        // Look for an available connection of the right protocol
        for _, conn := range conns {
            if conn.Protocol == protocol && !conn.inUse {
                p.mu.RUnlock()
                return conn, nil
            }
        }
    }
    
    p.mu.RUnlock()
    
    // No available connection found, create a new one
    conn, err := NewConnection(address, protocol)
    if err != nil {
        return nil, err
    }
    
    // Add the new connection to the pool
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.connections[address] = append(p.connections[address], conn)
    
    return conn, nil
}

// ReleaseConnection releases a connection back to the pool
func (p *Pool) ReleaseConnection(conn *Connection) {
    conn.inUse = false
}

// CloseConnections closes all connections in the pool
func (p *Pool) CloseConnections() error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    var lastErr error
    
    for _, conns := range p.connections {
        for _, conn := range conns {
            if err := conn.Close(); err != nil {
                lastErr = err
            }
        }
    }
    
    p.connections = make(map[string][]*Connection)
    
    return lastErr
}