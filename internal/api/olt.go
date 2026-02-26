package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
)

// OLTRequest represents the payload for creating or updating an OLT
type OLTRequest struct {
	ID     string `json:"id,omitempty"`
	Title  string `json:"title" binding:"required"`
	IP     string `json:"ip" binding:"required"`
	Vendor string `json:"vendor" binding:"required"`
	Model  string `json:"model" binding:"required"`
}

// ConnectionRequest represents the payload for creating or updating a connection
type ConnectionRequest struct {
	Protocol string `json:"protocol" binding:"required"`
	IP       string `json:"ip" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// handleListOLTs returns all registered OLTs
func (s *Server) handleListOLTs(c *gin.Context) {
	olts, err := s.oltRepo.ListOLTs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, olts)
}

// handleGetOLT returns a single OLT by ID
func (s *Server) handleGetOLT(c *gin.Context) {
	id := c.Param("id")
	olt, err := s.oltRepo.GetOLTByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
		return
	}
	c.JSON(http.StatusOK, olt)
}

// handleAddOLT creates a new OLT
func (s *Server) handleAddOLT(c *gin.Context) {
	var req OLTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := domain.ValidateTypeModel(req.Vendor, req.Model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	olt := &domain.OLTDetails{
		ID:     req.ID,
		Title:  req.Title,
		IP:     req.IP,
		Vendor: req.Vendor,
		Model:  req.Model,
	}

	id, err := s.oltRepo.AddOLT(olt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	olt.ID = id
	c.JSON(http.StatusCreated, olt)
}

// handleUpdateOLT updates an existing OLT
func (s *Server) handleUpdateOLT(c *gin.Context) {
	id := c.Param("id")

	if _, err := s.oltRepo.GetOLTByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
		return
	}

	var req OLTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := domain.ValidateTypeModel(req.Vendor, req.Model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	olt := &domain.OLTDetails{
		ID:     id,
		Title:  req.Title,
		IP:     req.IP,
		Vendor: req.Vendor,
		Model:  req.Model,
	}

	if err := s.oltRepo.UpdateOLT(olt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, olt)
}

// handleDeleteOLT removes an OLT by ID
func (s *Server) handleDeleteOLT(c *gin.Context) {
	id := c.Param("id")

	if err := s.oltRepo.DeleteOLT(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ─── Connection handlers ─────────────────────────────────────────────────────

// handleListConnections returns all connections for an OLT
func (s *Server) handleListConnections(c *gin.Context) {
	oltID := c.Param("id")
	if _, err := s.oltRepo.GetOLTByID(oltID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
		return
	}
	conns, err := s.oltRepo.ListConnections(oltID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conns)
}

// handleAddConnection adds a new connection to an OLT
func (s *Server) handleAddConnection(c *gin.Context) {
	oltID := c.Param("id")
	if _, err := s.oltRepo.GetOLTByID(oltID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
		return
	}

	var req ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := &domain.OLTConnection{
		OLTID:    oltID,
		Protocol: req.Protocol,
		IP:       req.IP,
		Username: req.Username,
		Password: req.Password,
	}

	id, err := s.oltRepo.AddConnection(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	conn.ID = id
	c.JSON(http.StatusCreated, conn)
}

// handleUpdateConnection updates an existing connection
func (s *Server) handleUpdateConnection(c *gin.Context) {
	connID := c.Param("connId")
	if _, err := s.oltRepo.GetConnectionByID(connID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}

	var req ConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := &domain.OLTConnection{
		ID:       connID,
		OLTID:    c.Param("id"),
		Protocol: req.Protocol,
		IP:       req.IP,
		Username: req.Username,
		Password: req.Password,
	}

	if err := s.oltRepo.UpdateConnection(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conn)
}

// handleDeleteConnection removes a connection by ID
func (s *Server) handleDeleteConnection(c *gin.Context) {
	connID := c.Param("connId")
	if err := s.oltRepo.DeleteConnection(connID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ─── Function/OLT-type handlers ──────────────────────────────────────────────

// handleListFunctions lists available functions for an OLT type
func (s *Server) handleListFunctions(c *gin.Context) {
	oltType := c.Query("vendor")
	if oltType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OLT type is required"})
		return
	}

	functions, err := s.oltRepo.ListFunctions(oltType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Functions not found for OLT type"})
		return
	}

	c.JSON(http.StatusOK, functions)
}

// handleGetFunctionDetails returns details of a specific function for an OLT type
func (s *Server) handleGetFunctionDetails(c *gin.Context) {
	name := c.Param("name")
	oltType := c.Query("vendor")

	if oltType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OLT type is required"})
		return
	}

	function, err := s.oltRepo.GetOLTFunction(oltType, dashToCamel(name))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		return
	}

	c.JSON(http.StatusOK, function)
}

// handleListOLTTypes returns all valid OLT type/model combinations.
func (s *Server) handleListOLTTypes(c *gin.Context) {
	supported := domain.SupportedTypes()
	type modelEntry struct {
		Vendor string   `json:"vendor"`
		Models []string `json:"models"`
	}
	result := make([]modelEntry, 0, len(supported))
	for t, models := range supported {
		ms := make([]string, len(models))
		for i, m := range models {
			ms[i] = string(m)
		}
		result = append(result, modelEntry{Vendor: string(t), Models: ms})
	}
	c.JSON(http.StatusOK, result)
}

// dashToCamel converts dash-case to camelCase (e.g. "list-onus" → "listOnus")
func dashToCamel(s string) string {
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
