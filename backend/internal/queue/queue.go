package queue

import (
    "errors"
    "sync"
)

// Queue represents a command queue
type Queue struct {
    commands     []*Command
    commandMap   map[string]*Command
    mu           sync.RWMutex
    notifyCh     chan struct{}
}

// New creates a new command queue
func New() *Queue {
    return &Queue{
        commands:   make([]*Command, 0),
        commandMap: make(map[string]*Command),
        notifyCh:   make(chan struct{}, 1),
    }
}

// Enqueue adds a command to the queue
func (q *Queue) Enqueue(cmd *Command) {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    q.commands = append(q.commands, cmd)
    q.commandMap[cmd.ID] = cmd
    
    // Notify consumers that a new command is available
    select {
    case q.notifyCh <- struct{}{}:
    default:
    }
}

// Dequeue removes and returns the next command from the queue
func (q *Queue) Dequeue() *Command {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    if len(q.commands) == 0 {
        return nil
    }
    
    cmd := q.commands[0]
    q.commands = q.commands[1:]
    
    return cmd
}

// GetCommandStatus returns the status, result, and any error for a command
func (q *Queue) GetCommandStatus(id string) (CommandStatus, string, error) {
    q.mu.RLock()
    defer q.mu.RUnlock()
    
    cmd, ok := q.commandMap[id]
    if !ok {
        return "", "", errors.New("command not found")
    }
    
    var errStr string
    if cmd.Error != nil {
        errStr = cmd.Error.Error()
    }
    
    return cmd.Status, cmd.Result, nil
}

// GetNotificationChannel returns a channel that is notified when a command is added
func (q *Queue) GetNotificationChannel() <-chan struct{} {
    return q.notifyCh
}

// Length returns the number of commands in the queue
func (q *Queue) Length() int {
    q.mu.RLock()
    defer q.mu.RUnlock()
    
    return len(q.commands)
}