package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/array/banking-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type regulatorEventRepository struct {
	db *gorm.DB
}

func NewRegulatorEventRepository(db *gorm.DB) RegulatorEventRepositoryInterface {
	return &regulatorEventRepository{db: db}
}

func (r *regulatorEventRepository) Create(event *models.RegulatorEvent) error {
	if event == nil {
		return errors.New("regulator event cannot be nil")
	}

	if err := r.db.Create(event).Error; err != nil {
		return fmt.Errorf("failed to create regulator event: %w", err)
	}

	return nil
}

func (r *regulatorEventRepository) FindPending(limit int) ([]*models.RegulatorEvent, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	var events []*models.RegulatorEvent
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Set("gorm:query_option", "FOR UPDATE SKIP LOCKED").
			Where("status = ? AND next_attempt_at <= ?", models.RegulatorEventStatusPending, time.Now().UTC()).
			Order("next_attempt_at ASC").
			Limit(limit).
			Find(&events).Error; err != nil {
			return fmt.Errorf("failed to find pending regulator events: %w", err)
		}

		if len(events) == 0 {
			return nil
		}

		ids := make([]uuid.UUID, len(events))
		for i, event := range events {
			ids[i] = event.ID
		}

		if err := tx.Model(&models.RegulatorEvent{}).
			Where("id IN ?", ids).
			Update("status", models.RegulatorEventStatusProcessing).Error; err != nil {
			return fmt.Errorf("failed to claim regulator events: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *regulatorEventRepository) MarkProcessing(id uuid.UUID) error {
	return r.updateStatus(id, models.RegulatorEventStatusProcessing)
}

func (r *regulatorEventRepository) MarkDelivered(id uuid.UUID) error {
	now := time.Now().UTC()
	result := r.db.Model(&models.RegulatorEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.RegulatorEventStatusDelivered,
			"delivered_at": now,
			"updated_at":   now,
			"last_error":   "",
		})
	if result.Error != nil {
		return fmt.Errorf("failed to mark regulator event delivered: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("regulator event %s not found", id)
	}
	return nil
}

func (r *regulatorEventRepository) MarkRetry(id uuid.UUID, attempts int, nextAttemptAt time.Time, lastError string) error {
	result := r.db.Model(&models.RegulatorEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          models.RegulatorEventStatusPending,
			"attempts":        attempts,
			"next_attempt_at": nextAttemptAt,
			"last_error":      lastError,
			"updated_at":      time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to schedule regulator event retry: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("regulator event %s not found", id)
	}
	return nil
}

func (r *regulatorEventRepository) MarkFailed(id uuid.UUID, lastError string) error {
	result := r.db.Model(&models.RegulatorEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     models.RegulatorEventStatusFailed,
			"last_error": lastError,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to mark regulator event failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("regulator event %s not found", id)
	}
	return nil
}

func (r *regulatorEventRepository) updateStatus(id uuid.UUID, status string) error {
	result := r.db.Model(&models.RegulatorEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update regulator event status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("regulator event %s not found", id)
	}
	return nil
}
