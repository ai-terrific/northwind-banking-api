package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/array/banking-api/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type externalTransferResultRecorder struct {
	results []*models.ExternalTransferResult
}

func (r *externalTransferResultRecorder) Record(result *models.ExternalTransferResult) error {
	r.results = append(r.results, result)
	return nil
}

func TestExternalBankAPIHandler_ValidateTransfer(t *testing.T) {
	expected := models.TransferRequest{
		Amount:       2500,
		Currency:     "USD",
		Description:  "Payment for services",
		Direction:    "INBOUND",
		TransferType: "ACH",
	}

	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/external/transfers/validate", r.URL.Path)
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req models.TransferRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, expected.Amount, req.Amount)
		require.Equal(t, expected.Currency, req.Currency)
		require.Equal(t, expected.Description, req.Description)
		require.Equal(t, expected.Direction, req.Direction)
		require.Equal(t, expected.TransferType, req.TransferType)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"valid":true,"validation":{"valid":true,"issues":[],"metadata":{},"validation_time":"2026-09-21T00:00:00Z"}}`))
		require.NoError(t, err)
	}))
	defer externalServer.Close()

	handler := NewExternalAPIHandler(externalServer.URL, "test-key")
	e := echo.New()
	payload := `{"amount":2500,"currency":"USD","description":"Payment for services","direction":"INBOUND","transfer_type":"ACH","source_account":{"account_holder_name":"Jane Doe","account_number":"123456789","institution_name":"Sample Bank","routing_number":"021000021"},"destination_account":{"account_holder_name":"John Doe","account_number":"987654321","institution_name":"Another Bank","routing_number":"011000015"},"reference_number":"REF123456","scheduled_date":"2026-01-01T12:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/external/transfers/validate", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ValidateTransfer(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"valid":true`)
}

func TestExternalBankAPIHandler_BatchTransfersRecordsSuccessAndFailure(t *testing.T) {
	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req models.TransferRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.Header().Set("Content-Type", "application/json")
		if req.ReferenceNumber == "FAIL" {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(`{"error":"insufficient funds"}`))
			require.NoError(t, err)
			return
		}
		_, err := w.Write([]byte(`{"transfer_id":"ext-123","reference_number":"REF-123","status":"completed"}`))
		require.NoError(t, err)
	}))
	defer externalServer.Close()

	recorder := &externalTransferResultRecorder{}
	handler := NewExternalAPIHandler(externalServer.URL, "test-key")
	handler.SetExternalTransferResultRecorder(recorder)
	e := echo.New()
	payload := `{"transfers":[{"amount":100,"currency":"USD","description":"success","reference_number":"OK"},{"amount":200,"currency":"USD","description":"failure","reference_number":"FAIL"}]}`
	req := httptest.NewRequest(http.MethodPost, "/external/transfers/batch", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.BatchTransfers(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, recorder.results, 2)

	statuses := map[string]bool{}
	for _, result := range recorder.results {
		statuses[result.Status] = true
		require.NotEmpty(t, result.Payload)
	}
	require.True(t, statuses["completed"])
	require.True(t, statuses["failed"])
}
