package port

import "context"

type IdempotencyStore interface {
	SetIfNotExists(ctx context.Context, key string, ttlSeconds int) (set bool, err error)
	Exists(ctx context.Context, key string) (bool, error)
}
