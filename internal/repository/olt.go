package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
	"gorm.io/gorm"
)

var (
	ErrOLTNotFound      = errors.New("olt not found")
	ErrFunctionNotFound = errors.New("function not found")
)

type OLTModel struct {
	ID     string `gorm:"primaryKey"`
	Title  string `gorm:"column:title;not null;default:''"`
	IP     string `gorm:"column:ip;not null"`
	Vendor string `gorm:"column:vendor;not null"`
	Model  string `gorm:"not null"`
}

func (OLTModel) TableName() string { return "olts" }

type dbOLTRepository struct {
	db        *gorm.DB
	functions map[string]map[string]*domain.FunctionDefinition
}

func NewDBOLTRepository(db *gorm.DB) (OLTRepository, error) {
	if err := db.AutoMigrate(&OLTModel{}, &OLTConnectionModel{}, &CommandExecutionModel{}); err != nil {
		return nil, err
	}
	repo := &dbOLTRepository{db: db}
	repo.initFunctions()
	return repo, nil
}

func (r *dbOLTRepository) GetOLTByID(id string) (*domain.OLTDetails, error) {
	var m OLTModel
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		return nil, ErrOLTNotFound
	}
	details := toDetails(&m)

	conns, err := r.ListConnections(id)
	if err == nil {
		details.Connections = make([]domain.OLTConnection, len(conns))
		for i, c := range conns {
			details.Connections[i] = *c
		}
	}
	return details, nil
}

func (r *dbOLTRepository) ListOLTs() ([]*domain.OLTDetails, error) {
	var models []OLTModel
	if err := r.db.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.OLTDetails, len(models))
	for i, m := range models {
		d := toDetails(&m)
		conns, err := r.ListConnections(m.ID)
		if err == nil {
			d.Connections = make([]domain.OLTConnection, len(conns))
			for j, c := range conns {
				d.Connections[j] = *c
			}
		}
		result[i] = d
	}
	return result, nil
}

func (r *dbOLTRepository) AddOLT(olt *domain.OLTDetails) (string, error) {
	if olt.ID == "" {
		olt.ID = uuid.New().String()
	}
	m := fromDetails(olt)
	if err := r.db.Create(m).Error; err != nil {
		return "", err
	}
	return olt.ID, nil
}

func (r *dbOLTRepository) UpdateOLT(olt *domain.OLTDetails) error {
	result := r.db.Model(&OLTModel{}).Where("id = ?", olt.ID).Updates(map[string]interface{}{
		"title":  olt.Title,
		"ip":     olt.IP,
		"vendor": olt.Vendor,
		"model":  olt.Model,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOLTNotFound
	}
	return nil
}

func (r *dbOLTRepository) DeleteOLT(id string) error {
	// Remove child connections first.
	r.db.Delete(&OLTConnectionModel{}, "olt_id = ?", id)

	result := r.db.Delete(&OLTModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOLTNotFound
	}
	return nil
}

func (r *dbOLTRepository) GetOLTFunction(oltType, functionName string) (*domain.FunctionDefinition, error) {
	fns, ok := r.functions[oltType]
	if !ok {
		return nil, ErrFunctionNotFound
	}
	fn, ok := fns[functionName]
	if !ok {
		return nil, ErrFunctionNotFound
	}
	return fn, nil
}

func (r *dbOLTRepository) ListFunctions(oltType string) ([]*domain.FunctionDefinition, error) {
	fns, ok := r.functions[oltType]
	if !ok {
		return nil, ErrFunctionNotFound
	}
	result := make([]*domain.FunctionDefinition, 0, len(fns))
	for _, fn := range fns {
		result = append(result, fn)
	}
	return result, nil
}

func toDetails(m *OLTModel) *domain.OLTDetails {
	return &domain.OLTDetails{
		ID:     m.ID,
		Title:  m.Title,
		IP:     m.IP,
		Vendor: m.Vendor,
		Model:  m.Model,
	}
}

func fromDetails(d *domain.OLTDetails) *OLTModel {
	return &OLTModel{
		ID:     d.ID,
		Title:  d.Title,
		IP:     d.IP,
		Vendor: d.Vendor,
		Model:  d.Model,
	}
}

func (r *dbOLTRepository) initFunctions() {
	r.functions = map[string]map[string]*domain.FunctionDefinition{
		"nokia-fx16": {
			"listUnregisteredOnus": {
				Name:        "listUnregisteredOnus",
				Description: "Listar ONUs não registradas",
				Commands:    []string{"show pon unprovision"},
				Parameters:  []string{},
			},
			"getOntStatus": {
				Name:        "getOntStatus",
				Description: "Status de uma ONT específica",
				Commands:    []string{"show interface gpon {{slot}}/{{port}} onu {{onuId}} status"},
				Parameters:  []string{"slot", "port", "onuId"},
			},
			"getOntSignal": {
				Name:        "getOntSignal",
				Description: "Nível de sinal de uma ONT específica",
				Commands:    []string{"show interface gpon {{slot}}/{{port}} onu {{onuId}} optical-info"},
				Parameters:  []string{"slot", "port", "onuId"},
			},
		},
		"huawei-ma5800": {
			"listUnregisteredOnus": {
				Name:        "listUnregisteredOnus",
				Description: "Listar ONUs não registradas",
				Commands:    []string{"display ont autofind all"},
				Parameters:  []string{},
			},
			"getOntStatus": {
				Name:        "getOntStatus",
				Description: "Status de uma ONT específica",
				Commands:    []string{"display ont info {{frameId}} {{slot}} {{port}} {{onuId}}"},
				Parameters:  []string{"frameId", "slot", "port", "onuId"},
			},
			"getOntSignal": {
				Name:        "getOntSignal",
				Description: "Nível de sinal de uma ONT específica",
				Commands:    []string{"display ont optical-info {{frameId}} {{slot}} {{port}} {{onuId}}"},
				Parameters:  []string{"frameId", "slot", "port", "onuId"},
			},
		},
	}
}
