package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
)

type OLTConnectionModel struct {
	ID       string `gorm:"primaryKey"`
	OLTID    string `gorm:"column:olt_id;not null;index"`
	Protocol string `gorm:"not null"`
	IP       string `gorm:"not null"`
	Username string `gorm:"not null"`
	Password string `gorm:"not null"`
}

func (OLTConnectionModel) TableName() string { return "olt_connections" }

func (r *dbOLTRepository) ListConnections(oltID string) ([]*domain.OLTConnection, error) {
	var models []OLTConnectionModel
	if err := r.db.Where("olt_id = ?", oltID).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.OLTConnection, len(models))
	for i, m := range models {
		result[i] = toConnectionDetails(&m)
	}
	return result, nil
}

func (r *dbOLTRepository) GetConnectionByID(id string) (*domain.OLTConnection, error) {
	var m OLTConnectionModel
	if err := r.db.First(&m, "id = ?", id).Error; err != nil {
		return nil, ErrConnectionNotFound
	}
	return toConnectionDetails(&m), nil
}

func (r *dbOLTRepository) AddConnection(conn *domain.OLTConnection) (string, error) {
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}
	m := fromConnectionDetails(conn)
	if err := r.db.Create(m).Error; err != nil {
		return "", err
	}
	return conn.ID, nil
}

func (r *dbOLTRepository) UpdateConnection(conn *domain.OLTConnection) error {
	result := r.db.Model(&OLTConnectionModel{}).Where("id = ?", conn.ID).Updates(map[string]interface{}{
		"protocol": conn.Protocol,
		"ip":       conn.IP,
		"username": conn.Username,
		"password": conn.Password,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

func (r *dbOLTRepository) DeleteConnection(id string) error {
	result := r.db.Delete(&OLTConnectionModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConnectionNotFound
	}
	return nil
}

func toConnectionDetails(m *OLTConnectionModel) *domain.OLTConnection {
	return &domain.OLTConnection{
		ID:       m.ID,
		OLTID:    m.OLTID,
		Protocol: m.Protocol,
		IP:       m.IP,
		Username: m.Username,
		Password: m.Password,
	}
}

func fromConnectionDetails(c *domain.OLTConnection) *OLTConnectionModel {
	return &OLTConnectionModel{
		ID:       c.ID,
		OLTID:    c.OLTID,
		Protocol: c.Protocol,
		IP:       c.IP,
		Username: c.Username,
		Password: c.Password,
	}
}
