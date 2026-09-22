package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExternalTransferResult stores the result returned by the external bank API.
type ExternalTransferResult struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ExternalTransferID string    `gorm:"type:varchar(255);index" json:"external_transfer_id"`
	ReferenceNumber    string    `gorm:"type:varchar(255)" json:"reference_number,omitempty"`
	Status             string    `gorm:"type:varchar(50);not null" json:"status"`
	Error              string    `gorm:"type:text" json:"error,omitempty"`
	Payload            string    `gorm:"type:jsonb;not null" json:"payload"`
	CreatedAt          time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt          time.Time `gorm:"not null" json:"updated_at"`
}

func (r *ExternalTransferResult) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	now := time.Now().UTC()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = now
	}
	return nil
}

func (*ExternalTransferResult) TableName() string {
	return "external_transfer_results"
}
