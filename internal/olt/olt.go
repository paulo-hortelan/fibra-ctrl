package olt

import (
    "errors"

    "github.com/paulo-hortelan/fibra-ctrl/internal/connection"
)

// Common errors
var (
    ErrUnsupportedOLT = errors.New("unsupported OLT type")
)

// OLT defines the interface for OLT-specific operations
type OLT interface {
    // ExecuteCommand executes a command on the OLT
    ExecuteCommand(conn *connection.Connection, command string) (string, error)
    
    // ProcessResult processes the result of a command
    ProcessResult(command, result string) (string, error)
}