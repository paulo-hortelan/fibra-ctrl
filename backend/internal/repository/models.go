package repository

import (
	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
)

type OLTRepository interface {
	GetOLTByID(id string) (*domain.OLTDetails, error)
	GetOLTFunction(oltType, functionName string) (*domain.FunctionDefinition, error)
	ListOLTs() ([]*domain.OLTDetails, error)
	ListFunctions(oltType string) ([]*domain.FunctionDefinition, error)
	AddOLT(olt *domain.OLTDetails) (string, error)
	UpdateOLT(olt *domain.OLTDetails) error
	DeleteOLT(id string) error
}
