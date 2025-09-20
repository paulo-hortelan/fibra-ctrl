package api

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/paulo-hortelan/fibra-ctrl/internal/queue"
    "github.com/paulo-hortelan/fibra-ctrl/internal/repository"
)

type Server struct {
    router   *gin.Engine
    addr     string
    queue    *queue.Queue
    oltRepo  repository.OLTRepository
    userRepo repository.UserRepository
}

func NewServer(addr string, queue *queue.Queue, oltRepo repository.OLTRepository, userRepo repository.UserRepository) *Server {
    server := &Server{
        router:   gin.Default(),
        addr:     addr,
        queue:    queue,
        oltRepo:  oltRepo,
        userRepo: userRepo,
    }

    server.setupRoutes()
    return server
}

func (s *Server) setupRoutes() {
    // Auth routes
    s.router.POST("/register", s.handleRegister)
    s.router.POST("/login", s.handleLogin)
    s.router.GET("/me", s.authMiddleware(), s.handleMe)
    s.router.POST("/logout", s.authMiddleware(), s.handleLogout)

    // OLT management routes
    s.router.GET("/olts", s.handleListOLTs)
    s.router.GET("/olts/:id", s.handleGetOLT)
    s.router.POST("/olts", s.handleAddOLT)
    s.router.PUT("/olts/:id", s.handleUpdateOLT)
    s.router.DELETE("/olts/:id", s.handleDeleteOLT)
    
    // Function information routes
    s.router.GET("/functions", s.handleListFunctions)
    s.router.GET("/functions/:name", s.handleGetFunctionDetails)
}

func (s *Server) Start() error {
    return s.router.Run(s.addr)
}