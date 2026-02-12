package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
)

type MetricsRepository struct {
	db *gorm.DB
}

func NewMetricsRepository(db *gorm.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (r *MetricsRepository) GetQueueDepth(ctx context.Context, queueName string) (int, error) {
	return 0, nil // placeholder, use RabbitMQ management API in real impl
}

func (r *MetricsRepository) GetNotificationStats(ctx context.Context) (*port.NotificationStats, error) {
	stats := &port.NotificationStats{}
	type result struct {
		Status string
		Count  int64
	}
	var results []result
	err := r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	for _, res := range results {
		switch res.Status {
		case "pending":
			stats.Pending = res.Count
		case "queued":
			stats.Queued = res.Count
		case "sent":
			stats.Sent = res.Count
		case "failed":
			stats.Failed = res.Count
		}
		stats.Total += res.Count
	}
	return stats, nil
}
