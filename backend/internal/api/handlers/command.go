package api

import (
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v5"
    "github.com/paulo-hortelan/fibra-ctrl/internal/queue"
    "github.com/paulo-hortelan/fibra-ctrl/internal/repository"
)

type CommandRequest struct {
    OltID        string                 `json:"oltId"`         // ID of the registered OLT
    FunctionName string                 `json:"functionName"`  // Name of the function to execute (e.g., "list-unregistered-onus")
    Parameters   map[string]interface{} `json:"parameters,omitempty"` // Optional parameters for the function
}

type CommandResponse struct {
    ID     string `json:"id"`
    Status string `json:"status"`
}

func (s *Server) handleCommandSubmission(c *gin.Context) {
    var req CommandRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate the request
    if req.OltID == "" || req.FunctionName == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: oltId and functionName"})
        return
    }

    // Get OLT details from database
    oltDetails, err := s.oltRepo.GetOLTByID(req.OltID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "OLT not found"})
        return
    }

    // Map function name to actual command(s)
    commands, err := s.mapFunctionToCommands(req.FunctionName, req.Parameters, oltDetails)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid function: %s", err.Error())})
        return
    }

    // Create a job that contains one or more commands
    job := queue.NewJob(req.OltID, req.FunctionName)
    
    // Add commands to the job
    for _, cmdInfo := range commands {
        cmd := queue.NewCommand(
            oltDetails.IP, 
            oltDetails.Type, 
            oltDetails.Model, 
            cmdInfo.CommandStr, 
            oltDetails.Protocol,
        )
        job.AddCommand(cmd)
    }

    // Add the job to the queue
    jobID := s.queue.EnqueueJob(job)

    // Return the job ID for status tracking
    c.JSON(http.StatusAccepted, CommandResponse{
        ID:     jobID,
        Status: "queued",
    })
}

// mapFunctionToCommands maps a function name to OLT-specific commands
func (s *Server) mapFunctionToCommands(functionName string, params map[string]interface{}, olt *OLTDetails) ([]CommandInfo, error) {
    // Convert dash-case to camelCase for function name mapping
    funcName := convertToCamelCase(functionName)
    
    // Get function definition for the specific OLT type
    functionDef, err := s.oltRepo.GetOLTFunction(olt.Type, funcName)
    if err != nil {
        return nil, fmt.Errorf("function not supported for this OLT: %s", err)
    }
    
    // Build the commands using the function definition and parameters
    commands := make([]CommandInfo, 0, len(functionDef.Commands))
    
    for i, cmdTemplate := range functionDef.Commands {
        // Replace parameters in command template if needed
        cmdStr := cmdTemplate
        
        // Apply parameters to command template
        for name, value := range params {
            placeholder := fmt.Sprintf("{{%s}}", name)
            cmdStr = strings.Replace(cmdStr, placeholder, fmt.Sprintf("%v", value), -1)
        }
        
        commands = append(commands, CommandInfo{
            CommandStr: cmdStr,
            Order:      i,
        })
    }
    
    return commands, nil
}

// convertToCamelCase converts dash-case to camelCase
func convertToCamelCase(input string) string {
    parts := strings.Split(input, "-")
    for i := 1; i < len(parts); i++ {
        if len(parts[i]) > 0 {
            parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
        }
    }
    return strings.Join(parts, "")
}

// handleCommandStatus handles status queries for commands
func (s *Server) handleCommandStatus(c *gin.Context) {
    id := c.Param("id")
    
    // Look up job status in the queue
    status, result, err := s.queue.GetJobStatus(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":     id,
        "status": status,
        "result": result,
    })
}

