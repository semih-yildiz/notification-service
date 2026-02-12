package list

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

func (u *UseCase) ListByQuery(ctx context.Context, q *Query) (*port.ListResult, error) {
	filter := port.ListFilter{
		Status:   q.Status,
		Channel:  q.Channel,
		FromTime: q.FromTime,
		ToTime:   q.ToTime,
		BatchID:  q.BatchID,
		Limit:    q.Limit,
		Offset:   q.Offset,
	}

	return u.repo.List(ctx, filter)
}
