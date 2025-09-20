package worker

import (
    "github.com/paulo-hortelan/fibra-ctrl/internal/connection"
    "github.com/paulo-hortelan/fibra-ctrl/internal/olt"
    "github.com/paulo-hortelan/fibra-ctrl/internal/olt/nokia"
    "github.com/paulo-hortelan/fibra-ctrl/internal/queue"
)

// Worker represents a worker that executes commands
type Worker struct {
    id          int
    connPool    *connection.Pool
    queue       *queue.Queue
    stopCh      chan struct{}
    oltHandlers map[string]olt.OLT
}

// NewWorker creates a new worker
func NewWorker(id int, connPool *connection.Pool, queue *queue.Queue) *Worker {
    return &Worker{
        id:       id,
        connPool: connPool,
        queue:    queue,
        stopCh:   make(chan struct{}),
        oltHandlers: map[string]olt.OLT{
            "nokia_fx16": nokia.NewFX16Handler(),
        },
    }
}

// Start starts the worker
func (w *Worker) Start() {
    go w.process()
}

// Stop stops the worker
func (w *Worker) Stop() {
    close(w.stopCh)
}

// process processes commands from the queue
func (w *Worker) process() {
    notifyCh := w.queue.GetNotificationChannel()
    
    for {
        select {
        case <-w.stopCh:
            return
        case <-notifyCh:
            // Process commands until the queue is empty
            for {
                cmd := w.queue.Dequeue()
                if cmd == nil {
                    break
                }
                
                w.executeCommand(cmd)
            }
        }
    }
}

// executeCommand executes a command
func (w *Worker) executeCommand(cmd *queue.Command) {
    // Update command status
    cmd.SetStatus(queue.StatusInProgress)
    
    // Get the appropriate OLT handler
    oltHandler, ok := w.oltHandlers[cmd.OltType+"_"+cmd.OltModel]
    if !ok {
        cmd.SetError(olt.ErrUnsupportedOLT)
        return
    }
    
    // Get a connection from the pool
    protocol := connection.Protocol(cmd.Protocol)
    conn, err := w.connPool.GetConnection(cmd.OltIP, protocol)
    if err != nil {
        cmd.SetError(err)
        return
    }
    
    // Execute the command
    result, err := oltHandler.ExecuteCommand(conn, cmd.Command)
    if err != nil {
        cmd.SetError(err)
        return
    }
    
    // Process the result
    processedResult, err := oltHandler.ProcessResult(cmd.Command, result)
    if err != nil {
        cmd.SetError(err)
        return
    }
    
    // Update the command with the result
    cmd.SetResult(processedResult)
    
    // Release the connection back to the pool
    w.connPool.ReleaseConnection(conn)
}