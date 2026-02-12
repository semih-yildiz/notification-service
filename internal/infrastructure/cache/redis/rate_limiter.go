package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

const (
	rateLimitMaxPerSecond = 100
	rateLimitKeyPrefix    = "ratelimit:channel:"
	rateLimitWindow       = time.Second
)

var _ port.RateLimiter = (*RateLimiter)(nil)

// RateLimiter implements port.RateLimiter (max 100/sec per channel).
type RateLimiter struct {
	client *redis.Client
}

// NewRateLimiter returns a new Redis rate limiter.
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow returns true if the channel is under the limit.
func (r *RateLimiter) Allow(ctx context.Context, channel notification.Channel) (bool, error) {
	key := rateLimitKeyPrefix + channel.String()
	now := time.Now().Truncate(rateLimitWindow)
	windowKey := fmt.Sprintf("%s:%d", key, now.Unix())
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, rateLimitWindow*2)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	return incr.Val() <= rateLimitMaxPerSecond, nil
}
