package queue

import (
    "time"

    "github.com/google/uuid"
)

// CommandStatus represents the status of a command
type CommandStatus string

const (
    // StatusQueued means the command is queued but not yet executed
    StatusQueued CommandStatus = "queued"
    
    // StatusInProgress means the command is currently being executed
    StatusInProgress CommandStatus = "in_progress"
    
    // StatusCompleted means the command has been executed successfully
    StatusCompleted CommandStatus = "completed"
    
    // StatusFailed means the command execution failed
    StatusFailed CommandStatus = "failed"
)

// Command represents a command to be executed on an OLT
type Command struct {
    ID        string
    OltIP     string
    OltType   string
    OltModel  string
    Command   string
    Protocol  string
    Status    CommandStatus
    Result    string
    Error     error
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NewCommand creates a new command
func NewCommand(oltIP, oltType, oltModel, commandStr, protocol string) *Command {
    return &Command{
        ID:        uuid.New().String(),
        OltIP:     oltIP,
        OltType:   oltType,
        OltModel:  oltModel,
        Command:   commandStr,
        Protocol:  protocol,
        Status:    StatusQueued,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

// SetStatus updates the command status
func (c *Command) SetStatus(status CommandStatus) {
    c.Status = status
    c.UpdatedAt = time.Now()
}

// SetResult sets the command result
func (c *Command) SetResult(result string) {
    c.Result = result
    c.Status = StatusCompleted
    c.UpdatedAt = time.Now()
}

// SetError sets the command error
func (c *Command) SetError(err error) {
    c.Error = err
    c.Status = StatusFailed
    c.UpdatedAt = time.Now()
}