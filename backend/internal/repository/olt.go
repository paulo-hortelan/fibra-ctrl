package repository

import (
	"errors"
	"sync"

	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
)

var (
	ErrOLTNotFound      = errors.New("olt not found")
	ErrFunctionNotFound = errors.New("function not found")
)

// InMemoryOLTRepository implements OLTRepository with in-memory storage
type InMemoryOLTRepository struct {
	olts           map[string]*domain.OLTDetails
	functionsByOLT map[string]map[string]*domain.FunctionDefinition
	mu             sync.RWMutex
}

// NewInMemoryOLTRepository creates a new in-memory OLT repository
func NewInMemoryOLTRepository() *InMemoryOLTRepository {
	repo := &InMemoryOLTRepository{
		olts:           make(map[string]*domain.OLTDetails),
		functionsByOLT: make(map[string]map[string]*domain.FunctionDefinition),
	}
	
	// Initialize with some predefined functions for each OLT type
	repo.initPredefinedFunctions()
	
	return repo
}

// GetOLTByID retrieves an OLT by its ID
func (r *InMemoryOLTRepository) GetOLTByID(id string) (*domain.OLTDetails, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	olt, ok := r.olts[id]
	if !ok {
		return nil, ErrOLTNotFound
	}
	
	return olt, nil
}

// GetOLTFunction retrieves a function definition for an OLT type
func (r *InMemoryOLTRepository) GetOLTFunction(oltType, functionName string) (*domain.FunctionDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	functions, ok := r.functionsByOLT[oltType]
	if !ok {
		return nil, ErrFunctionNotFound
	}
	
	function, ok := functions[functionName]
	if !ok {
		return nil, ErrFunctionNotFound
	}
	
	return function, nil
}

// ListOLTs returns a list of all OLTs
func (r *InMemoryOLTRepository) ListOLTs() ([]*domain.OLTDetails, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	olts := make([]*domain.OLTDetails, 0, len(r.olts))
	for _, olt := range r.olts {
		olts = append(olts, olt)
	}
	
	return olts, nil
}

// ListFunctions returns all functions available for an OLT type
func (r *InMemoryOLTRepository) ListFunctions(oltType string) ([]*domain.FunctionDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	functions, ok := r.functionsByOLT[oltType]
	if !ok {
		return nil, ErrFunctionNotFound
	}
	
	result := make([]*domain.FunctionDefinition, 0, len(functions))
	for _, fn := range functions {
		result = append(result, fn)
	}
	
	return result, nil
}

// AddOLT adds a new OLT to the repository
func (r *InMemoryOLTRepository) AddOLT(olt *domain.OLTDetails) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Store the OLT
	r.olts[olt.ID] = olt
	
	return olt.ID, nil
}

// UpdateOLT updates an existing OLT
func (r *InMemoryOLTRepository) UpdateOLT(olt *domain.OLTDetails) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.olts[olt.ID]; !ok {
		return ErrOLTNotFound
	}
	
	r.olts[olt.ID] = olt
	return nil
}

// DeleteOLT removes an OLT from the repository
func (r *InMemoryOLTRepository) DeleteOLT(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.olts[id]; !ok {
		return ErrOLTNotFound
	}
	
	delete(r.olts, id)
	return nil
}

// initPredefinedFunctions initializes the repository with predefined functions for each OLT type
func (r *InMemoryOLTRepository) initPredefinedFunctions() {
	// Nokia FX16 functions
	nokiaFunctions := map[string]*domain.FunctionDefinition{
		"listUnregisteredOnus": {
			Name:        "listUnregisteredOnus",
			Description: "List all unregistered ONUs on the OLT",
			Commands:    []string{"show pon unprovision"},
			Parameters:  []string{},
		},
		"getOntStatus": {
			Name:        "getOntStatus",
			Description: "Get status of a specific ONT",
			Commands:    []string{"show interface gpon {{slot}}/{{port}} onu {{onuId}} status"},
			Parameters:  []string{"slot", "port", "onuId"},
		},
		"getOntSignal": {
			Name:        "getOntSignal",
			Description: "Get signal levels of a specific ONT",
			Commands:    []string{"show interface gpon {{slot}}/{{port}} onu {{onuId}} optical-info"},
			Parameters:  []string{"slot", "port", "onuId"},
		},
	}
	
	// Huawei MA5800 functions
	huaweiFunctions := map[string]*domain.FunctionDefinition{
		"listUnregisteredOnus": {
			Name:        "listUnregisteredOnus",
			Description: "List all unregistered ONUs on the OLT",
			Commands:    []string{"display ont autofind all"},
			Parameters:  []string{},
		},
		"getOntStatus": {
			Name:        "getOntStatus",
			Description: "Get status of a specific ONT",
			Commands:    []string{"display ont info {{frameId}} {{slot}} {{port}} {{onuId}}"},
			Parameters:  []string{"frameId", "slot", "port", "onuId"},
		},
		"getOntSignal": {
			Name:        "getOntSignal",
			Description: "Get signal levels of a specific ONT",
			Commands:    []string{"display ont optical-info {{frameId}} {{slot}} {{port}} {{onuId}}"},
			Parameters:  []string{"frameId", "slot", "port", "onuId"},
		},
	}
	
	r.functionsByOLT = map[string]map[string]*domain.FunctionDefinition{
		"nokia-fx16": nokiaFunctions,
		"huawei-ma5800": huaweiFunctions,
	}
}
