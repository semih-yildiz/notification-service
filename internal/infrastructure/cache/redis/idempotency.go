package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
)

const idempotencyKeyPrefix = "idempotency:"
const defaultIdempotencyTTL = 86400 * 7

var _ port.IdempotencyStore = (*IdempotencyStore)(nil)

type IdempotencyStore struct {
	client *redis.Client
}

func NewIdempotencyStore(client *redis.Client) *IdempotencyStore {
	return &IdempotencyStore{client: client}
}

func (s *IdempotencyStore) SetIfNotExists(ctx context.Context, key string, ttlSeconds int) (bool, error) {
	if ttlSeconds <= 0 {
		ttlSeconds = defaultIdempotencyTTL
	}
	k := idempotencyKeyPrefix + key
	ok, err := s.client.SetNX(ctx, k, "1", time.Duration(ttlSeconds)*time.Second).Result()
	return ok, err
}

func (s *IdempotencyStore) Exists(ctx context.Context, key string) (bool, error) {
	k := idempotencyKeyPrefix + key
	n, err := s.client.Exists(ctx, k).Result()
	return n > 0, err
}
