package worker

import (
	"log"
	"time"

	"github.com/paulo-hortelan/fibra-ctrl/internal/connection"
	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
	"github.com/paulo-hortelan/fibra-ctrl/internal/olt"
	"github.com/paulo-hortelan/fibra-ctrl/internal/olt/nokia"
	"github.com/paulo-hortelan/fibra-ctrl/internal/queue"
	"github.com/paulo-hortelan/fibra-ctrl/internal/repository"
)

// Worker represents a worker that executes commands
type Worker struct {
	id          int
	connPool    *connection.Pool
	queue       *queue.Queue
	stopCh      chan struct{}
	oltHandlers map[string]olt.OLT
	oltRepo     repository.OLTRepository
}

// NewWorker creates a new worker
func NewWorker(id int, connPool *connection.Pool, queue *queue.Queue, oltRepo repository.OLTRepository) *Worker {
	return &Worker{
		id:       id,
		connPool: connPool,
		queue:    queue,
		oltRepo:  oltRepo,
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
	executedAt := time.Now()
	responseDuration := time.Duration(0)
	rawResponse := ""
	formattedResponse := ""

	defer func() {
		w.persistExecution(cmd, executedAt, responseDuration, rawResponse, formattedResponse)
	}()

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
	defer w.connPool.ReleaseConnection(conn)

	// Execute the command
	responseStart := time.Now()
	result, err := oltHandler.ExecuteCommand(conn, cmd.Command)
	responseDuration = time.Since(responseStart)
	if err != nil {
		cmd.SetError(err)
		return
	}
	rawResponse = result

	// Process the result
	processedResult, err := oltHandler.ProcessResult(cmd.Command, result)
	if err != nil {
		cmd.SetError(err)
		return
	}
	formattedResponse = processedResult

	// Update the command with the result
	cmd.SetResult(processedResult)
}

func (w *Worker) persistExecution(
	cmd *queue.Command,
	executedAt time.Time,
	responseDuration time.Duration,
	rawResponse string,
	formattedResponse string,
) {
	if w.oltRepo == nil {
		return
	}

	errorMessage := ""
	if cmd.Error != nil {
		errorMessage = cmd.Error.Error()
	}

	respondedAt := cmd.UpdatedAt
	if respondedAt.IsZero() {
		respondedAt = time.Now()
	}

	entry := &domain.CommandExecution{
		CommandID:         cmd.ID,
		ConnectionID:      cmd.ConnectionID,
		Command:           cmd.Command,
		Status:            string(cmd.Status),
		RawResponse:       rawResponse,
		FormattedResponse: formattedResponse,
		ErrorMessage:      errorMessage,
		ExecutedAt:        executedAt,
		RespondedAt:       respondedAt,
		ResponseTimeMs:    responseDuration.Milliseconds(),
		TotalTimeMs:       respondedAt.Sub(executedAt).Milliseconds(),
	}

	if _, err := w.oltRepo.CreateCommandExecution(entry); err != nil {
		log.Printf("failed to persist command execution %s: %v", cmd.ID, err)
	}
}
