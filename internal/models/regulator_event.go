package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RegulatorEventStatusPending    = "pending"
	RegulatorEventStatusProcessing = "processing"
	RegulatorEventStatusDelivered  = "delivered"
	RegulatorEventStatusFailed     = "failed"
)

type RegulatorEvent struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	EventID       string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"event_id"`
	TransferID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"transfer_id"`
	EventType     string     `gorm:"type:varchar(100);not null" json:"event_type"`
	Payload       string     `gorm:"type:jsonb;not null" json:"payload"`
	Status        string     `gorm:"type:varchar(20);not null;index" json:"status"`
	Attempts      int        `gorm:"not null;default:0" json:"attempts"`
	NextAttemptAt time.Time  `gorm:"not null;index" json:"next_attempt_at"`
	DeadlineAt    time.Time  `gorm:"not null;index" json:"deadline_at"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	LastError     string     `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt     time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null" json:"updated_at"`
}

func (e *RegulatorEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	if e.Status == "" {
		e.Status = RegulatorEventStatusPending
	}

	now := time.Now().UTC()

	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}

	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = now
	}

	return nil
}

func (*RegulatorEvent) TableName() string {
	return "regulator_events"
}

type RegulatorEventAttempt struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RegulatorEventID uuid.UUID `gorm:"type:uuid;not null;index" json:"regulator_event_id"`
	AttemptNumber    int       `gorm:"not null" json:"attempt_number"`
	StartedAt        time.Time `gorm:"not null" json:"started_at"`
	FinishedAt       time.Time `gorm:"not null" json:"finished_at"`
	ResponseStatus   *int      `json:"response_status,omitempty"`
	ErrorMessage     string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt        time.Time `gorm:"not null" json:"created_at"`
}

func (a *RegulatorEventAttempt) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	return nil
}

func (*RegulatorEventAttempt) TableName() string {
	return "regulator_event_attempts"
}
