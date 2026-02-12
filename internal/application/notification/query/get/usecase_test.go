package get

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

type mockNotificationRepo struct {
	getByIDFn      func(ctx context.Context, id string) (*notification.Notification, error)
	getByBatchIDFn func(ctx context.Context, batchID string) ([]*notification.Notification, error)
}

func (m *mockNotificationRepo) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockNotificationRepo) GetByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error) {
	if m.getByBatchIDFn != nil {
		return m.getByBatchIDFn(ctx, batchID)
	}
	return nil, errors.New("not found")
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *notification.Notification) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) CreateBatch(ctx context.Context, notifications []*notification.Notification) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) UpdateStatus(ctx context.Context, id string, status notification.Status, sentAt *time.Time, reason *string) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) List(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
	return nil, errors.New("not implemented")
}

func (m *mockNotificationRepo) CancelPending(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) CancelPendingByBatchID(ctx context.Context, batchID string) (int, error) {
	return 0, errors.New("not implemented")
}

func (m *mockNotificationRepo) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	return false, errors.New("not implemented")
}

type mockBatchRepo struct {
	getByIDFn func(ctx context.Context, id string) (*notification.Batch, error)
}

func (m *mockBatchRepo) GetByID(ctx context.Context, id string) (*notification.Batch, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockBatchRepo) Create(ctx context.Context, b *notification.Batch) error {
	return errors.New("not implemented")
}

func (m *mockBatchRepo) GetNotificationsByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error) {
	return nil, errors.New("not implemented")
}

func TestNotification_Success(t *testing.T) {
	repo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test",
				Status:    notification.StatusQueued,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	uc := NewUseCase(repo, &mockBatchRepo{})

	result, err := uc.Notification(context.Background(), &ByID{ID: "test-id"})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", result.ID)
	}
}

func TestNotification_NotFound(t *testing.T) {
	repo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return nil, errors.New("not found")
		},
	}

	uc := NewUseCase(repo, &mockBatchRepo{})

	result, err := uc.Notification(context.Background(), &ByID{ID: "non-existent"})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Error("expected nil result, got notification")
	}
}

func TestBatch_Success(t *testing.T) {
	batchRepo := &mockBatchRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Batch, error) {
			return &notification.Batch{
				ID:        id,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	notifRepo := &mockNotificationRepo{
		getByBatchIDFn: func(ctx context.Context, batchID string) ([]*notification.Notification, error) {
			return []*notification.Notification{
				{ID: "n1", BatchID: &batchID},
				{ID: "n2", BatchID: &batchID},
			}, nil
		},
	}

	uc := NewUseCase(notifRepo, batchRepo)

	batch, notifications, err := uc.Batch(context.Background(), &BatchByID{BatchID: "batch-id"})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if batch == nil {
		t.Fatal("expected batch, got nil")
	}
	if len(notifications) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(notifications))
	}
}

func TestBatch_NotFound(t *testing.T) {
	batchRepo := &mockBatchRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Batch, error) {
			return nil, errors.New("batch not found")
		},
	}

	uc := NewUseCase(&mockNotificationRepo{}, batchRepo)

	batch, notifications, err := uc.Batch(context.Background(), &BatchByID{BatchID: "non-existent"})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if batch != nil {
		t.Error("expected nil batch")
	}
	if notifications != nil {
		t.Error("expected nil notifications")
	}
}

func TestBatch_NotificationsError(t *testing.T) {
	batchRepo := &mockBatchRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Batch, error) {
			return &notification.Batch{
				ID:        id,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	notifRepo := &mockNotificationRepo{
		getByBatchIDFn: func(ctx context.Context, batchID string) ([]*notification.Notification, error) {
			return nil, errors.New("notifications query failed")
		},
	}

	uc := NewUseCase(notifRepo, batchRepo)

	batch, notifications, err := uc.Batch(context.Background(), &BatchByID{BatchID: "batch-id"})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if batch == nil {
		t.Error("expected batch even with notifications error")
	}
	if notifications != nil {
		t.Error("expected nil notifications on error")
	}
}
