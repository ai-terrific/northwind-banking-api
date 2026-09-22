package repositories

import (
	"fmt"

	"github.com/array/banking-api/internal/models"
	"gorm.io/gorm"
)

type externalTransferResultRepository struct {
	db *gorm.DB
}

func NewExternalTransferResultRepository(db *gorm.DB) ExternalTransferResultRepositoryInterface {
	return &externalTransferResultRepository{db: db}
}

func (r *externalTransferResultRepository) Create(result *models.ExternalTransferResult) error {
	if result == nil {
		return fmt.Errorf("external transfer result cannot be nil")
	}
	if err := r.db.Create(result).Error; err != nil {
		return fmt.Errorf("failed to create external transfer result: %w", err)
	}
	return nil
}
