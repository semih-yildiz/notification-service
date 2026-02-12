package cancel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

type mockNotificationRepo struct {
	cancelPendingFn          func(ctx context.Context, id string) error
	cancelPendingByBatchIDFn func(ctx context.Context, batchID string) (int, error)
}

func (m *mockNotificationRepo) CancelPending(ctx context.Context, id string) error {
	if m.cancelPendingFn != nil {
		return m.cancelPendingFn(ctx, id)
	}
	return nil
}

func (m *mockNotificationRepo) CancelPendingByBatchID(ctx context.Context, batchID string) (int, error) {
	if m.cancelPendingByBatchIDFn != nil {
		return m.cancelPendingByBatchIDFn(ctx, batchID)
	}
	return 0, nil
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *notification.Notification) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) CreateBatch(ctx context.Context, notifications []*notification.Notification) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	return nil, errors.New("not implemented")
}

func (m *mockNotificationRepo) UpdateStatus(ctx context.Context, id string, status notification.Status, sentAt *time.Time, reason *string) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) GetByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error) {
	return nil, errors.New("not implemented")
}

func (m *mockNotificationRepo) List(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockNotificationRepo) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	return false, errors.New("not implemented")
}

func TestCancelPendingNotification_Success(t *testing.T) {
	repo := &mockNotificationRepo{
		cancelPendingFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	uc := NewUseCase(repo)

	cmd := &Command{NotificationID: "test-id"}
	err := uc.CancelPendingNotification(context.Background(), cmd)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCancelPendingNotification_Error(t *testing.T) {
	repo := &mockNotificationRepo{
		cancelPendingFn: func(ctx context.Context, id string) error {
			return errors.New("cancel failed")
		},
	}

	uc := NewUseCase(repo)

	cmd := &Command{NotificationID: "test-id"}
	err := uc.CancelPendingNotification(context.Background(), cmd)

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCancelPendingNotificationBatch_Success(t *testing.T) {
	repo := &mockNotificationRepo{
		cancelPendingByBatchIDFn: func(ctx context.Context, batchID string) (int, error) {
			return 10, nil
		},
	}

	uc := NewUseCase(repo)

	cmd := &BatchCommand{BatchID: "batch-id"}
	count, err := uc.CancelPendingNotificationBatch(context.Background(), cmd)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if count != 10 {
		t.Errorf("expected count 10, got %d", count)
	}
}

func TestCancelPendingNotificationBatch_Error(t *testing.T) {
	repo := &mockNotificationRepo{
		cancelPendingByBatchIDFn: func(ctx context.Context, batchID string) (int, error) {
			return 0, errors.New("batch cancel failed")
		},
	}

	uc := NewUseCase(repo)

	cmd := &BatchCommand{BatchID: "batch-id"}
	count, err := uc.CancelPendingNotificationBatch(context.Background(), cmd)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}
