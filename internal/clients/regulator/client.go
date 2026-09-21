package regulator

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	url        string
	secret     string
	httpClient *http.Client
}

func NewClient(url, secret string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	return &Client{
		url:    url,
		secret: secret,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Send(ctx context.Context, eventID string, payload []byte) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create regulator request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Event-ID", eventID)
	request.Header.Set("X-Webhook-Signature", c.sign(payload))

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send regulator webhook: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("regulator returned status %d", response.StatusCode)
	}

	return nil
}

func (c *Client) sign(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(c.secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
