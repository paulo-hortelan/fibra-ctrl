package worker

import (
    "github.com/paulo-hortelan/fibra-ctrl/internal/connection"
    "github.com/paulo-hortelan/fibra-ctrl/internal/queue"
)

// Pool represents a pool of workers
type Pool struct {
    size     int
    workers  []*Worker
    connPool *connection.Pool
    queue    *queue.Queue
}

// NewPool creates a new worker pool
func NewPool(size int, connPool *connection.Pool, queue *queue.Queue) *Pool {
    return &Pool{
        size:     size,
        workers:  make([]*Worker, 0, size),
        connPool: connPool,
        queue:    queue,
    }
}

// Start starts all workers in the pool
func (p *Pool) Start() {
    for i := 0; i < p.size; i++ {
        worker := NewWorker(i, p.connPool, p.queue)
        worker.Start()
        p.workers = append(p.workers, worker)
    }
}

// Stop stops all workers in the pool
func (p *Pool) Stop() {
    for _, worker := range p.workers {
        worker.Stop()
    }
}