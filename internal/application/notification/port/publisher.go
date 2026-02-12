package port

import (
	"context"

	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

// NotificationEvent is the event payload published to the broker.
type NotificationEvent struct {
	NotificationID string
	BatchID        *string
	Recipient      string
	Channel        notification.Channel
	Content        string
	Priority       notification.Priority
	IdempotencyKey *string
	CreatedAt      string
}

// EventPublisher publishes notification events.
type EventPublisher interface {
	Publish(ctx context.Context, evt *NotificationEvent) error
	PublishBatch(ctx context.Context, events []*NotificationEvent) error
}
