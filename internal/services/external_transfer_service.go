package services

import (
	"github.com/array/banking-api/internal/models"
	"github.com/array/banking-api/internal/repositories"
)

type ExternalTransferResultService struct {
	repository repositories.ExternalTransferResultRepositoryInterface
	webhook    *RegulatorWebhookService
}

func NewExternalTransferResultService(
	repository repositories.ExternalTransferResultRepositoryInterface,
	webhook *RegulatorWebhookService,
) *ExternalTransferResultService {
	return &ExternalTransferResultService{repository: repository, webhook: webhook}
}

func (s *ExternalTransferResultService) Record(result *models.ExternalTransferResult) error {
	if err := s.repository.Create(result); err != nil {
		return err
	}
	if s.webhook == nil {
		return nil
	}
	return s.webhook.QueueExternalTransferResult(result)
}
