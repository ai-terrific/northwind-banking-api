package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/array/banking-api/internal/clients/regulator"
	"github.com/array/banking-api/internal/models"
	"github.com/array/banking-api/internal/repositories"
	"github.com/google/uuid"
)

const regulatorNotificationDeadline = 60 * time.Second

type regulatorTransferEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	TransferID uuid.UUID `json:"transfer_id"`
	Status     string    `json:"status"`
	Amount     string    `json:"amount"`
	Currency   string    `json:"currency"`
	OccurredAt time.Time `json:"occurred_at"`
	Error      string    `json:"error,omitempty"`
}

type RegulatorWebhookService struct {
	repository repositories.RegulatorEventRepositoryInterface
	client     *regulator.Client
	logger     *slog.Logger
}

func NewRegulatorWebhookService(
	repository repositories.RegulatorEventRepositoryInterface,
	client *regulator.Client,
	logger *slog.Logger,
) *RegulatorWebhookService {
	return &RegulatorWebhookService{repository: repository, client: client, logger: logger}
}

func (s *RegulatorWebhookService) QueueTransferEvent(transfer *models.Transfer) error {
	if transfer == nil {
		return fmt.Errorf("transfer cannot be nil")
	}

	eventType := "transfer.failed"
	if transfer.Status == models.TransferStatusCompleted {
		eventType = "transfer.succeeded"
	}

	event := regulatorTransferEvent{
		EventID:    "transfer-" + transfer.ID.String(),
		EventType:  eventType,
		TransferID: transfer.ID,
		Status:     transfer.Status,
		Amount:     transfer.Amount.String(),
		Currency:   "USD",
		OccurredAt: time.Now().UTC(),
	}
	if transfer.ErrorMessage != nil {
		event.Error = *transfer.ErrorMessage
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal regulator event: %w", err)
	}

	now := time.Now().UTC()
	return s.repository.Create(&models.RegulatorEvent{
		EventID:       event.EventID,
		TransferID:    transfer.ID,
		EventType:     eventType,
		Payload:       string(payload),
		Status:        models.RegulatorEventStatusPending,
		NextAttemptAt: now,
		DeadlineAt:    now.Add(regulatorNotificationDeadline),
	})
}

func (s *RegulatorWebhookService) Start(ctx context.Context) {
	go s.run(ctx)
}

func (s *RegulatorWebhookService) run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.processPending(ctx); err != nil {
				s.logger.Error("regulator webhook worker failed", "error", err)
			}
		}
	}
}

func (s *RegulatorWebhookService) processPending(ctx context.Context) error {
	events, err := s.repository.FindPending(50)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := s.processEvent(ctx, event); err != nil {
			s.logger.Error("failed to process regulator event", "error", err, "event_id", event.EventID)
		}
	}
	return nil
}

func (s *RegulatorWebhookService) processEvent(ctx context.Context, event *models.RegulatorEvent) error {
	if event == nil {
		return fmt.Errorf("regulator event cannot be nil")
	}

	if time.Now().UTC().After(event.DeadlineAt) {
		return s.repository.MarkFailed(event.ID, "regulator notification deadline exceeded")
	}

	sendCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	startedAt := time.Now().UTC()
	statusCode, sendErr := s.client.Send(sendCtx, event.EventID, []byte(event.Payload))
	finishedAt := time.Now().UTC()
	attempt := &models.RegulatorEventAttempt{
		RegulatorEventID: event.ID,
		AttemptNumber:    event.Attempts + 1,
		StartedAt:        startedAt,
		FinishedAt:       finishedAt,
	}
	if statusCode != 0 {
		attempt.ResponseStatus = &statusCode
	}
	if sendErr != nil {
		attempt.ErrorMessage = sendErr.Error()
	}
	if err := s.repository.RecordAttempt(attempt); err != nil {
		return err
	}

	if sendErr == nil {
		return s.repository.MarkDelivered(event.ID)
	}

	attempts := event.Attempts + 1
	nextAttempt := time.Now().UTC().Add(retryDelay(attempts))
	return s.repository.MarkRetry(event.ID, attempts, nextAttempt, sendErr.Error())
}

func retryDelay(attempt int) time.Duration {
	delays := []time.Duration{2 * time.Second, 5 * time.Second, 10 * time.Second, 15 * time.Second, 20 * time.Second}
	if attempt <= 0 {
		return 0
	}
	if attempt > len(delays) {
		return delays[len(delays)-1]
	}
	return delays[attempt-1]
}
