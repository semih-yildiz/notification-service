package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ManagementClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

type QueueStats struct {
	Name     string `json:"name"`
	Messages int    `json:"messages"`
	Ready    int    `json:"messages_ready"`
	Unacked  int    `json:"messages_unacknowledged"`
}

func NewManagementClient(baseURL, username, password string) *ManagementClient {
	return &ManagementClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (m *ManagementClient) GetQueueStats(ctx context.Context, vhost, queueName string) (*QueueStats, error) {
	if vhost == "/" {
		vhost = "%2F"
	}

	url := fmt.Sprintf("%s/api/queues/%s/%s", m.baseURL, vhost, queueName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(m.username, m.password)
	req.Header.Set("Accept", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq management api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rabbitmq management api: status %d: %s", resp.StatusCode, string(body))
	}

	var stats QueueStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("rabbitmq management api decode: %w", err)
	}

	return &stats, nil
}

func (m *ManagementClient) GetAllQueueStats(ctx context.Context, vhost string, queueNames []string) (map[string]*QueueStats, error) {
	result := make(map[string]*QueueStats)

	for _, queueName := range queueNames {
		stats, err := m.GetQueueStats(ctx, vhost, queueName)
		if err != nil {
			result[queueName] = &QueueStats{Name: queueName, Messages: -1}
			continue
		}
		result[queueName] = stats
	}

	return result, nil
}

type QueueDepth struct {
	Queue   string `json:"queue"`
	Depth   int    `json:"depth"`
	Ready   int    `json:"ready"`
	Unacked int    `json:"unacked"`
}

func (m *ManagementClient) GetQueueDepths(ctx context.Context) ([]QueueDepth, error) {
	queues := []string{
		QueueSMS, QueueEmail, QueuePush,
		QueueSMSDLQ, QueueEmailDLQ, QueuePushDLQ,
	}

	stats, err := m.GetAllQueueStats(ctx, "/", queues)
	if err != nil {
		return nil, err
	}

	depths := make([]QueueDepth, 0, len(queues))
	for _, queueName := range queues {
		if s, ok := stats[queueName]; ok {
			depths = append(depths, QueueDepth{
				Queue:   queueName,
				Depth:   s.Messages,
				Ready:   s.Ready,
				Unacked: s.Unacked,
			})
		}
	}

	return depths, nil
}
