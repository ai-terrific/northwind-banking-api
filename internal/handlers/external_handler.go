package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/array/banking-api/internal/config"
	"github.com/array/banking-api/internal/dto"
	"github.com/labstack/echo/v4"
)

var cfg *config.Config = config.Load()

type ExternalBankAPIHandler struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func NewExternalAPIHandler(baseURL, apiKey string) *ExternalBankAPIHandler {
	return &ExternalBankAPIHandler{
		client:  &http.Client{Timeout: time.Duration(cfg.ExternalBank.Timeout) * time.Second},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (h *ExternalBankAPIHandler) ValidationAccount(c echo.Context) error {
	var req dto.AccountValidationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	body, err := json.Marshal(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to encode request")
	}

	httpReq, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		h.baseURL+"/external/accounts/validate",
		bytes.NewReader(body),
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.apiKey)

	response, err := h.client.Do(httpReq)
	log.Println(h.baseURL)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadGateway,
			"external API unavailable",
		)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadGateway,
			"failed to read external response",
		)
	}

	return c.Blob(
		response.StatusCode,
		response.Header.Get("Content-Type"),
		responseBody,
	)
}
