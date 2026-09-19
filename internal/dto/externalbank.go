package dto

type AccountValidationRequest struct {
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	RoutingNumber     string `json:"routing_number"`
}
