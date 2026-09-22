package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/array/banking-api/internal/config"
	"github.com/array/banking-api/internal/dto"
	"github.com/array/banking-api/internal/models"
	"github.com/labstack/echo/v4"
)

var cfg *config.Config = config.Load()

type ExternalBankAPIHandler struct {
	client         *http.Client
	baseURL        string
	apiKey         string
	resultRecorder ExternalTransferResultRecorder
}

type ExternalTransferResultRecorder interface {
	Record(result *models.ExternalTransferResult) error
}

func NewExternalAPIHandler(baseURL, apiKey string) *ExternalBankAPIHandler {
	return &ExternalBankAPIHandler{
		client:  &http.Client{Timeout: time.Duration(cfg.ExternalBank.Timeout) * time.Second},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (h *ExternalBankAPIHandler) SetExternalTransferResultRecorder(recorder ExternalTransferResultRecorder) {
	h.resultRecorder = recorder
}

func (h *ExternalBankAPIHandler) newExternalRequest(c echo.Context, payload interface{}, endpoint string) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to encode request")
	}

	httpReq, err := http.NewRequestWithContext(
		c.Request().Context(),
		http.MethodPost,
		h.baseURL+endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.apiKey)

	return httpReq, nil
}

func (h *ExternalBankAPIHandler) ValidationAccount(c echo.Context) error {
	var req dto.AccountValidationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	httpReq, err := h.newExternalRequest(c, req, "/external/accounts/validate")
	if err != nil {
		return err
	}

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

func (h *ExternalBankAPIHandler) ValidateTransfer(c echo.Context) error {
	var req models.TransferRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	httpReq, err := h.newExternalRequest(c, req, "/external/transfers/validate")
	if err != nil {
		return err
	}

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

func (h *ExternalBankAPIHandler) InitiateExternalTransfer(c echo.Context) error {
	var req models.TransferRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	httpReq, err := h.newExternalRequest(c, req, "/external/transfers/initiate")
	if err != nil {
		return err
	}

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

func (h *ExternalBankAPIHandler) BatchTransfers(c echo.Context) error {
	var req dto.BatchTransferRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	type transferResult struct {
		index int
		item  dto.BatchTransferItem
	}

	results := make(chan transferResult, len(req.Transfers))
	for index, transfer := range req.Transfers {
		go func(index int, transfer models.TransferRequest) {
			results <- transferResult{
				index: index,
				item:  h.initiateBatchTransfer(c, transfer),
			}
		}(index, transfer)
	}

	response := dto.BatchTransferResponse{
		TotalTransfers: len(req.Transfers),
		Transfers:      make([]dto.BatchTransferItem, len(req.Transfers)),
	}

	for range req.Transfers {
		result := <-results
		response.Transfers[result.index] = result.item
		if result.item.Error == "" {
			response.Accepted++
		} else {
			response.Rejected++
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h *ExternalBankAPIHandler) initiateBatchTransfer(c echo.Context, transfer models.TransferRequest) (item dto.BatchTransferItem) {
	item = dto.BatchTransferItem{Status: "failed"}
	defer func() {
		if h.resultRecorder == nil {
			return
		}
		payload, err := json.Marshal(map[string]interface{}{
			"event_type":       "external_transfer.result",
			"transfer_id":      item.TransferID,
			"status":           item.Status,
			"reference_number": item.ReferenceNu,
			"error":            item.Error,
		})
		if err != nil {
			log.Printf("failed to marshal external transfer result: %v", err)
			return
		}
		if err := h.resultRecorder.Record(&models.ExternalTransferResult{
			ExternalTransferID: item.TransferID,
			ReferenceNumber:    item.ReferenceNu,
			Status:             item.Status,
			Error:              item.Error,
			Payload:            string(payload),
		}); err != nil {
			log.Printf("failed to persist external transfer result: %v", err)
		}
	}()

	httpReq, err := h.newExternalRequest(c, transfer, "/external/transfers/initiate")
	if err != nil {
		item.Error = err.Error()
		return item
	}

	response, err := h.client.Do(httpReq)
	if err != nil {
		item.Error = "external API unavailable"
		return item
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		item.Error = "failed to read external response"
		return item
	}

	var transferResponse dto.TransferStatusResponse
	if err := json.Unmarshal(responseBody, &transferResponse); err == nil {
		item.TransferID = transferResponse.TransferID
		item.ReferenceNu = transferResponse.ReferenceNumber
		item.Status = transferResponse.Status
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		item.Error = strings.TrimSpace(string(responseBody))
		if item.Error == "" {
			item.Error = response.Status
		}
		item.Status = "failed"
		return item
	}

	if item.Status == "" {
		item.Status = "accepted"
	}
	return item
}
