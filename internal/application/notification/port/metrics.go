package port

import "context"

// MetricsProvider returns system metrics (queue depth, success/failure rates, etc.).
type MetricsProvider interface {
	GetQueueDepth(ctx context.Context, queueName string) (int, error)
	GetNotificationStats(ctx context.Context) (*NotificationStats, error)
}

type NotificationStats struct {
	Pending int64
	Queued  int64
	Sent    int64
	Failed  int64
	Total   int64
}
