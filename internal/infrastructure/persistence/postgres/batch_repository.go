package postgres

import (
	"context"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
	"gorm.io/gorm"
)

var _ port.BatchRepository = (*BatchRepository)(nil)

type BatchRepository struct {
	db *gorm.DB
}

func NewBatchRepository(db *gorm.DB) *BatchRepository {
	return &BatchRepository{db: db}
}

func (r *BatchRepository) Create(ctx context.Context, b *notification.Batch) error {
	m := &BatchModel{
		ID:             b.ID,
		IdempotencyKey: b.IdempotencyKey,
		CreatedAt:      b.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *BatchRepository) GetByID(ctx context.Context, id string) (*notification.Batch, error) {
	var m BatchModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notification.ErrNotFound
		}
		return nil, err
	}
	return &notification.Batch{
		ID:             m.ID,
		IdempotencyKey: m.IdempotencyKey,
		CreatedAt:      m.CreatedAt,
	}, nil
}
