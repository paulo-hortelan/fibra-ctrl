package api

import (
	"github.com/gin-gonic/gin"
	"github.com/paulo-hortelan/fibra-ctrl/internal/queue"
	"github.com/paulo-hortelan/fibra-ctrl/internal/repository"
)

type Server struct {
	router  *gin.Engine
	addr    string
	queue   *queue.Queue
	oltRepo repository.OLTRepository
}

func NewServer(addr string, queue *queue.Queue, oltRepo repository.OLTRepository) *Server {
	server := &Server{
		router:  gin.Default(),
		addr:    addr,
		queue:   queue,
		oltRepo: oltRepo,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// OLT management routes
	s.router.GET("/olts", s.handleListOLTs)
	s.router.GET("/olts/types", s.handleListOLTTypes)
	s.router.GET("/olts/:id", s.handleGetOLT)
	s.router.POST("/olts", s.handleAddOLT)
	s.router.PUT("/olts/:id", s.handleUpdateOLT)
	s.router.DELETE("/olts/:id", s.handleDeleteOLT)

	// Connection management routes (nested under OLT)
	s.router.GET("/olts/:id/connections", s.handleListConnections)
	s.router.POST("/olts/:id/connections", s.handleAddConnection)
	s.router.PUT("/olts/:id/connections/:connId", s.handleUpdateConnection)
	s.router.DELETE("/olts/:id/connections/:connId", s.handleDeleteConnection)

	// Function information routes
	s.router.GET("/functions", s.handleListFunctions)
	s.router.GET("/functions/:name", s.handleGetFunctionDetails)

	// Command execution routes
	s.router.POST("/commands", s.handleCommandSubmission)
	s.router.GET("/commands/:id", s.handleCommandStatus)
}

func (s *Server) Start() error {
	return s.router.Run(s.addr)
}
