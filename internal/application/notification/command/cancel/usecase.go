package cancel

import (
	"context"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
)

type UseCase struct {
	repo port.NotificationRepository
}

func NewUseCase(repo port.NotificationRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) CancelPendingNotification(ctx context.Context, cmd *Command) error {
	return u.repo.CancelPending(ctx, cmd.NotificationID)
}

func (u *UseCase) CancelPendingNotificationBatch(ctx context.Context, cmd *BatchCommand) (int, error) {
	return u.repo.CancelPendingByBatchID(ctx, cmd.BatchID)
}
