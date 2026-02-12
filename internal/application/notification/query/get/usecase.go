package get

import (
	"context"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

type UseCase struct {
	notifRepo port.NotificationRepository
	batchRepo port.BatchRepository
}

func NewUseCase(notifRepo port.NotificationRepository, batchRepo port.BatchRepository) *UseCase {
	return &UseCase{notifRepo: notifRepo, batchRepo: batchRepo}
}

func (u *UseCase) Notification(ctx context.Context, q *ByID) (*notification.Notification, error) {
	return u.notifRepo.GetByID(ctx, q.ID)
}

func (u *UseCase) Batch(ctx context.Context, q *BatchByID) (*notification.Batch, []*notification.Notification, error) {
	batch, err := u.batchRepo.GetByID(ctx, q.BatchID)
	if err != nil {
		return nil, nil, err
	}

	list, err := u.notifRepo.GetByBatchID(ctx, q.BatchID)
	if err != nil {
		return batch, nil, err
	}

	return batch, list, nil
}
