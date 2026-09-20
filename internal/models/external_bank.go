package models

type ValidationIssue struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Field    string `json:"field"`
	Severity string `json:"severity"`
}

type ValidationResult struct {
	Issues         []ValidationIssue      `json:"issues"`
	Metadata       map[string]interface{} `json:"metadata"`
	Valid          bool                   `json:"valid"`
	ValidationTime string                 `json:"validation_time"`
}

type ValidationResponse struct {
	Data       map[string]interface{} `json:"data"`
	Validation ValidationResult       `json:"validation"`
}

type ExternalAccountDetails struct {
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	InstitutionName   string `json:"institution_name"`
	RoutingNumber     string `json:"routing_number"`
}

type TransferRequest struct {
	Amount             int                    `json:"amount"`
	Currency           string                 `json:"currency"`
	Description        string                 `json:"description"`
	DestinationAccount ExternalAccountDetails `json:"destination_account"`
	Direction          string                 `json:"direction"`
	ReferenceNumber    string                 `json:"reference_number"`
	ScheduledDate      string                 `json:"scheduled_date"`
	SourceAccount      ExternalAccountDetails `json:"source_account"`
	TransferType       string                 `json:"transfer_type"`
}
