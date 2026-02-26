package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
	"github.com/paulo-hortelan/fibra-ctrl/internal/queue"
)

type CommandRequest struct {
	OltID        string                 `json:"oltId"`
	ConnectionID string                 `json:"connectionId,omitempty"`
	FunctionName string                 `json:"functionName"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

type CommandResponse struct {
	IDs    []string `json:"ids"`
	Status string   `json:"status"`
}

// handleCommandSubmission submits one or more commands for async execution
func (s *Server) handleCommandSubmission(c *gin.Context) {
	var req CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OltID == "" || req.FunctionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "oltId and functionName are required"})
		return
	}

	oltDetails, err := s.oltRepo.GetOLTByID(req.OltID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
		return
	}

	// Resolve which connection to use.
	var conn *domain.OLTConnection
	if req.ConnectionID != "" {
		conn, err = s.oltRepo.GetConnectionByID(req.ConnectionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
			return
		}
	} else {
		conns, err := s.oltRepo.ListConnections(req.OltID)
		if err != nil || len(conns) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no connections configured for this OLT"})
			return
		}
		conn = conns[0]
	}

	commands, err := s.mapFunctionToCommands(req.FunctionName, req.Parameters, oltDetails)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid function: %s", err.Error())})
		return
	}

	ids := make([]string, 0, len(commands))
	for _, ci := range commands {
		cmd := queue.NewCommand(
			conn.ID,
			conn.IP,
			oltDetails.Vendor,
			oltDetails.Model,
			ci.CommandStr,
			conn.Protocol,
			conn.Username,
			conn.Password,
		)
		s.queue.Enqueue(cmd)
		ids = append(ids, cmd.ID)
	}

	c.JSON(http.StatusAccepted, CommandResponse{IDs: ids, Status: "queued"})
}

// handleCommandStatus returns the current status of a queued command
func (s *Server) handleCommandStatus(c *gin.Context) {
	id := c.Param("id")

	status, result, err := s.queue.GetCommandStatus(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "command not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "status": status, "result": result})
}

// mapFunctionToCommands resolves a function name to its OLT-specific command strings
func (s *Server) mapFunctionToCommands(
	functionName string,
	params map[string]interface{},
	olt *domain.OLTDetails,
) ([]domain.CommandInfo, error) {
	functionDef, err := s.oltRepo.GetOLTFunction(olt.Vendor, dashToCamel(functionName))
	if err != nil {
		return nil, fmt.Errorf("function not supported for this OLT: %w", err)
	}

	commands := make([]domain.CommandInfo, 0, len(functionDef.Commands))
	for i, template := range functionDef.Commands {
		cmdStr := template
		for name, value := range params {
			cmdStr = strings.ReplaceAll(cmdStr, fmt.Sprintf("{{%s}}", name), fmt.Sprintf("%v", value))
		}
		commands = append(commands, domain.CommandInfo{CommandStr: cmdStr, Order: i})
	}

	return commands, nil
}
