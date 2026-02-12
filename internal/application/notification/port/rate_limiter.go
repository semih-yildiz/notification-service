package port

import (
	"context"

	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

// RateLimiter limits send rate per channel.
type RateLimiter interface {
	Allow(ctx context.Context, channel notification.Channel) (allowed bool, err error)
}
