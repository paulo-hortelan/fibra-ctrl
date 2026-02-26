package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/paulo-hortelan/fibra-ctrl/internal/domain"
)

type CommandExecutionModel struct {
	ID                string             `gorm:"primaryKey"`
	CommandID         string             `gorm:"column:command_id;not null;index"`
	ConnectionID      string             `gorm:"column:connection_id;not null;index"`
	Command           string             `gorm:"column:command;not null;type:text"`
	Status            string             `gorm:"column:status;not null"`
	RawResponse       string             `gorm:"column:raw_response;type:text"`
	FormattedResponse string             `gorm:"column:formatted_response;type:text"`
	ErrorMessage      string             `gorm:"column:error_message;type:text"`
	ExecutedAt        time.Time          `gorm:"column:executed_at;not null"`
	RespondedAt       time.Time          `gorm:"column:responded_at;not null"`
	ResponseTimeMs    int64              `gorm:"column:response_time_ms;not null"`
	TotalTimeMs       int64              `gorm:"column:total_time_ms;not null"`
	Connection        OLTConnectionModel `gorm:"foreignKey:ConnectionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func (CommandExecutionModel) TableName() string { return "command_executions" }

func (r *dbOLTRepository) CreateCommandExecution(entry *domain.CommandExecution) (string, error) {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	model := &CommandExecutionModel{
		ID:                entry.ID,
		CommandID:         entry.CommandID,
		ConnectionID:      entry.ConnectionID,
		Command:           entry.Command,
		Status:            entry.Status,
		RawResponse:       entry.RawResponse,
		FormattedResponse: entry.FormattedResponse,
		ErrorMessage:      entry.ErrorMessage,
		ExecutedAt:        entry.ExecutedAt,
		RespondedAt:       entry.RespondedAt,
		ResponseTimeMs:    entry.ResponseTimeMs,
		TotalTimeMs:       entry.TotalTimeMs,
	}

	if err := r.db.Create(model).Error; err != nil {
		return "", err
	}

	return entry.ID, nil
}
