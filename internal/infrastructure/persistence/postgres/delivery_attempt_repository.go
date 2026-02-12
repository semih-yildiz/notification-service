package postgres

import (
	"context"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
	"gorm.io/gorm"
)

var _ port.DeliveryAttemptRepository = (*DeliveryAttemptRepository)(nil)

type DeliveryAttemptRepository struct {
	db *gorm.DB
}

func NewDeliveryAttemptRepository(db *gorm.DB) *DeliveryAttemptRepository {
	return &DeliveryAttemptRepository{db: db}
}

func (r *DeliveryAttemptRepository) Create(ctx context.Context, a *notification.DeliveryAttempt) error {
	m := &DeliveryAttemptModel{
		ID:             a.ID,
		NotificationID: a.NotificationID,
		AttemptNumber:  a.AttemptNumber,
		Success:        a.Success,
		StatusCode:     a.StatusCode,
		ResponseBody:   a.ResponseBody,
		ErrorMessage:   a.ErrorMessage,
		CreatedAt:      a.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}
