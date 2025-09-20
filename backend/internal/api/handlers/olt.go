package api

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    "github.com/paulo-hortelan/fibra-ctrl/internal/domain"
)

// OLTRequest represents the payload for creating or updating an OLT
type OLTRequest struct {
    ID       string `json:"id,omitempty"`
    IP       string `json:"ip" binding:"required"`
    Type     string `json:"type" binding:"required"`
    Model    string `json:"model" binding:"required"`
    Protocol string `json:"protocol" binding:"required"`
}

// handleListOLTs handles requests to list all OLTs
func (s *Server) handleListOLTs(c *gin.Context) {
    olts, err := s.oltRepo.ListOLTs()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, olts)
}

// handleGetOLT handles requests to get a specific OLT by ID
func (s *Server) handleGetOLT(c *gin.Context) {
    id := c.Param("id")
    
    olt, err := s.oltRepo.GetOLTByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
        return
    }
    
    c.JSON(http.StatusOK, olt)
}

// handleAddOLT handles requests to add a new OLT
func (s *Server) handleAddOLT(c *gin.Context) {
    var req OLTRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Create a new OLT object
    olt := &domain.OLTDetails{
        ID:       req.ID, // If empty, the repository should generate an ID
        IP:       req.IP,
        Type:     req.Type,
        Model:    req.Model,
        Protocol: req.Protocol,
    }
    
    // Add the OLT to the repository
    id, err := s.oltRepo.AddOLT(olt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    // Return the created OLT with its ID
    olt.ID = id
    c.JSON(http.StatusCreated, olt)
}

// handleUpdateOLT handles requests to update an existing OLT
func (s *Server) handleUpdateOLT(c *gin.Context) {
    id := c.Param("id")
    
    // Check if the OLT exists
    _, err := s.oltRepo.GetOLTByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
        return
    }
    
    var req OLTRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Update the OLT
    olt := &domain.OLTDetails{
        ID:       id,
        IP:       req.IP,
        Type:     req.Type,
        Model:    req.Model,
        Protocol: req.Protocol,
    }
    
    if err := s.oltRepo.UpdateOLT(olt); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, olt)
}

// handleDeleteOLT handles requests to delete an OLT
func (s *Server) handleDeleteOLT(c *gin.Context) {
    id := c.Param("id")
    
    if err := s.oltRepo.DeleteOLT(id); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
        return
    }
    
    c.Status(http.StatusNoContent)
}

// handleListFunctions handles requests to list available functions for OLT types
func (s *Server) handleListFunctions(c *gin.Context) {
    oltType := c.Query("type")
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

// handleGetFunctionDetails handles requests to get details of a specific function
func (s *Server) handleGetFunctionDetails(c *gin.Context) {
    name := c.Param("name")
    oltType := c.Query("type")
    
    if oltType == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "OLT type is required"})
        return
    }
    
    // Convert the function name from dash-case to camelCase
    camelCaseName := convertToCamelCase(name)
    
    function, err := s.oltRepo.GetOLTFunction(oltType, camelCaseName)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
        return
    }
    
    c.JSON(http.StatusOK, function)
}

// convertToCamelCase converts a dash-case string to camelCase
func convertToCamelCase(s string) string {
    // Split the string by dash
    parts := strings.Split(s, "-")
    
    // Capitalize first letter of each part except the first one
    for i := 1; i < len(parts); i++ {
        if len(parts[i]) > 0 {
            parts[i] = strings.ToUpper(parts[i][0:1]) + parts[i][1:]
        }
    }
    
    // Join the parts back together
    return strings.Join(parts, "")
}
