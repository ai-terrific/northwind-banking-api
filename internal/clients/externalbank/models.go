package externalbank

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
