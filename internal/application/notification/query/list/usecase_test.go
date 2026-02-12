package list

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

type mockNotificationRepo struct {
	listFn func(ctx context.Context, filter port.ListFilter) (*port.ListResult, error)
}

func (m *mockNotificationRepo) List(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return &port.ListResult{
		Notifications: []*notification.Notification{},
		Total:         0,
	}, nil
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

func (m *mockNotificationRepo) CancelPending(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) CancelPendingByBatchID(ctx context.Context, batchID string) (int, error) {
	return 0, errors.New("not implemented")
}

func (m *mockNotificationRepo) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	return false, errors.New("not implemented")
}

func TestListByQuery_Success(t *testing.T) {
	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
			return &port.ListResult{
				Notifications: []*notification.Notification{
					{ID: "n1", Status: notification.StatusQueued},
					{ID: "n2", Status: notification.StatusQueued},
				},
				Total: 2,
			}, nil
		},
	}

	uc := NewUseCase(repo)

	status := notification.StatusQueued
	query := &Query{
		Status: &status,
		Limit:  10,
		Offset: 0,
	}

	result, err := uc.ListByQuery(context.Background(), query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.Notifications) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Notifications))
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
}

func TestListByQuery_WithFilters(t *testing.T) {
	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
			if filter.Status == nil {
				t.Error("expected status filter")
			}
			if filter.Channel == nil {
				t.Error("expected channel filter")
			}
			return &port.ListResult{Notifications: []*notification.Notification{}, Total: 0}, nil
		},
	}

	uc := NewUseCase(repo)

	status := notification.StatusQueued
	channel := notification.ChannelSMS
	query := &Query{
		Status:  &status,
		Channel: &channel,
		Limit:   50,
		Offset:  0,
	}

	_, err := uc.ListByQuery(context.Background(), query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestListByQuery_WithTimeRange(t *testing.T) {
	fromTime := time.Now().Add(-24 * time.Hour)
	toTime := time.Now()

	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
			if filter.FromTime == nil {
				t.Error("expected fromTime filter")
			}
			if filter.ToTime == nil {
				t.Error("expected toTime filter")
			}
			return &port.ListResult{Notifications: []*notification.Notification{}, Total: 0}, nil
		},
	}

	uc := NewUseCase(repo)

	query := &Query{
		FromTime: &fromTime,
		ToTime:   &toTime,
		Limit:    100,
		Offset:   0,
	}

	_, err := uc.ListByQuery(context.Background(), query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestListByQuery_WithBatchID(t *testing.T) {
	batchID := "batch-123"

	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
			if filter.BatchID == nil {
				t.Error("expected batchID filter")
			}
			if *filter.BatchID != batchID {
				t.Errorf("expected batchID %s, got %s", batchID, *filter.BatchID)
			}
			return &port.ListResult{Notifications: []*notification.Notification{}, Total: 0}, nil
		},
	}

	uc := NewUseCase(repo)

	query := &Query{
		BatchID: &batchID,
		Limit:   100,
		Offset:  0,
	}

	_, err := uc.ListByQuery(context.Background(), query)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestListByQuery_Error(t *testing.T) {
	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
			return nil, errors.New("database error")
		},
	}

	uc := NewUseCase(repo)

	query := &Query{
		Limit:  10,
		Offset: 0,
	}

	result, err := uc.ListByQuery(context.Background(), query)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestListByQuery_Pagination(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		offset int
	}{
		{"First page", 10, 0},
		{"Second page", 10, 10},
		{"Large limit", 100, 0},
		{"High offset", 10, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockNotificationRepo{
				listFn: func(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
					if filter.Limit != tt.limit {
						t.Errorf("expected limit %d, got %d", tt.limit, filter.Limit)
					}
					if filter.Offset != tt.offset {
						t.Errorf("expected offset %d, got %d", tt.offset, filter.Offset)
					}
					return &port.ListResult{Notifications: []*notification.Notification{}, Total: 0}, nil
				},
			}

			uc := NewUseCase(repo)

			query := &Query{
				Limit:  tt.limit,
				Offset: tt.offset,
			}

			_, err := uc.ListByQuery(context.Background(), query)

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
