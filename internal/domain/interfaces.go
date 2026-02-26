package domain

import (
	"errors"
	"time"
)

type OLTType string

type OLTModel string

const (
	OLTTypeNokia OLTType = "nokia"
)

const (
	OLTModelFX16 OLTModel = "fx16"
)

// validModels lists the allowed model for each type.
var validModels = map[OLTType][]OLTModel{
	OLTTypeNokia: {OLTModelFX16},
}

// SupportedTypes returns each OLT type together with its accepted models.
func SupportedTypes() map[OLTType][]OLTModel {
	return validModels
}

// ValidateTypeModel returns an error if the type/model combination is not supported.
func ValidateTypeModel(oltType, model string) error {
	models, ok := validModels[OLTType(oltType)]
	if !ok {
		return errors.New("unsupported OLT type: " + oltType)
	}
	for _, m := range models {
		if m == OLTModel(model) {
			return nil
		}
	}
	return errors.New("unsupported model '" + model + "' for OLT type '" + oltType + "'")
}

type OLTDetails struct {
	ID          string
	Title       string
	IP          string
	Vendor      string
	Model       string
	Connections []OLTConnection
}

// OLTConnection represents a single connection endpoint for an OLT device.
// An OLT may have multiple connections (e.g. SSH on one IP, Telnet on another).
type OLTConnection struct {
	ID       string
	OLTID    string
	Protocol string
	IP       string
	Username string
	Password string
}

type FunctionDefinition struct {
	Name        string   // Function name (camelCase)
	Description string   // Human-readable description
	Commands    []string // List of command templates to execute in order
	Parameters  []string // List of required parameter names
}

type CommandInfo struct {
	CommandStr string
	Order      int // Execution order if multiple commands are needed
}

type CommandExecution struct {
	ID                string
	CommandID         string
	ConnectionID      string
	Command           string
	Status            string
	RawResponse       string
	FormattedResponse string
	ErrorMessage      string
	ExecutedAt        time.Time
	RespondedAt       time.Time
	ResponseTimeMs    int64
	TotalTimeMs       int64
}
