package repository

import (
	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
)

type OLTRepository interface {
	// OLT CRUD
	GetOLTByID(id string) (*domain.OLTDetails, error)
	ListOLTs() ([]*domain.OLTDetails, error)
	AddOLT(olt *domain.OLTDetails) (string, error)
	UpdateOLT(olt *domain.OLTDetails) error
	DeleteOLT(id string) error

	// Connection CRUD
	ListConnections(oltID string) ([]*domain.OLTConnection, error)
	GetConnectionByID(id string) (*domain.OLTConnection, error)
	AddConnection(conn *domain.OLTConnection) (string, error)
	UpdateConnection(conn *domain.OLTConnection) error
	DeleteConnection(id string) error

	// Function catalog
	GetOLTFunction(vendor, functionName string) (*domain.FunctionDefinition, error)
	ListFunctions(vendor string) ([]*domain.FunctionDefinition, error)

	// Command execution logs
	CreateCommandExecution(entry *domain.CommandExecution) (string, error)
}
