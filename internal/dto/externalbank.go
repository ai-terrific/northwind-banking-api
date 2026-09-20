package dto

import "github.com/array/banking-api/internal/models"

type AccountValidationRequest struct {
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	RoutingNumber     string `json:"routing_number"`
}

type AccountDetails struct {
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	InstitutionName   string `json:"institution_name"`
	RoutingNumber     string `json:"routing_number"`
}

type StatusHistory struct {
	Description string `json:"description"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
}

type TransferStatusResponse struct {
	Amount                 int             `json:"amount"`
	CompletedDate          string          `json:"completed_date"`
	Currency               string          `json:"currency"`
	Description            string          `json:"description"`
	DestinationAccount     AccountDetails  `json:"destination_account"`
	Direction              string          `json:"direction"`
	ErrorCode              string          `json:"error_code"`
	ErrorMessage           string          `json:"error_message"`
	ExchangeRate           int             `json:"exchange_rate"`
	ExceptedCompletionDate string          `json:"excepted_completion_date"`
	Fee                    float32         `json:"fee"`
	InitiatedDate          string          `json:"initiated_date"`
	ProcessingDate         string          `json:"processing_date"`
	ReferenceNumber        string          `json:"reference_number"`
	RetryCount             int             `json:"retry_count"`
	SourceAccount          AccountDetails  `json:"source_account"`
	Status                 string          `json:"status"`
	StatusHistory          []StatusHistory `json:"status_history"`
	TransferID             string          `json:"transfer_id"`
	TransferType           string          `json:"transfer_type"`
}

type BatchTransferRequest struct {
	Transfers []models.TransferRequest `json:"transfers"`
}

type BatchTransferItem struct {
	Error       string `json:"error"`
	ReferenceNu string `json:"reference_number"`
	Status      string `json:"status"`
	TransferID  string `json:"transfer_id"`
}

type BatchTransferResponse struct {
	Accepted       int                 `json:"accepted"`
	BatchID        string              `json:"batch_id"`
	Rejected       int                 `json:"rejected"`
	TotalTransfers int                 `json:"total_transfers"`
	Transfers      []BatchTransferItem `json:"transfer"`
}
