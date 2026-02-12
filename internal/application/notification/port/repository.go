package port

import (
	"context"
	"time"

	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *notification.Notification) error
	CreateBatch(ctx context.Context, notifications []*notification.Notification) error
	GetByID(ctx context.Context, id string) (*notification.Notification, error)
	GetByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error)
	UpdateStatus(ctx context.Context, id string, status notification.Status, sentAt *time.Time, failureReason *string) error
	List(ctx context.Context, filter ListFilter) (*ListResult, error)
	CancelPending(ctx context.Context, id string) error
	CancelPendingByBatchID(ctx context.Context, batchID string) (int, error)
	ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error)
}

type BatchRepository interface {
	Create(ctx context.Context, b *notification.Batch) error
	GetByID(ctx context.Context, id string) (*notification.Batch, error)
}

type DeliveryAttemptRepository interface {
	Create(ctx context.Context, a *notification.DeliveryAttempt) error
}

type ListFilter struct {
	Status   *notification.Status
	Channel  *notification.Channel
	FromTime *time.Time
	ToTime   *time.Time
	BatchID  *string
	Limit    int
	Offset   int
}

type ListResult struct {
	Notifications []*notification.Notification
	Total         int
}
