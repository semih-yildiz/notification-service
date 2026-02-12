package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) Deliver(ctx context.Context, req *port.DeliveryRequest) (*port.DeliveryResponse, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var out port.DeliveryResponse
	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK {
		_ = json.Unmarshal(data, &out)
		if out.Timestamp == "" {
			out.Timestamp = time.Now().Format(time.RFC3339)
		}
		return &out, resp.StatusCode, nil
	}
	return nil, resp.StatusCode, fmt.Errorf("delivery failed: status %d body %s", resp.StatusCode, string(data))
}
