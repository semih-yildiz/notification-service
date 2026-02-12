package process

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
	updateStatusFn func(ctx context.Context, id string, status notification.Status, sentAt *time.Time, reason *string) error
}

func (m *mockNotificationRepo) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockNotificationRepo) UpdateStatus(ctx context.Context, id string, status notification.Status, sentAt *time.Time, reason *string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, sentAt, reason)
	}
	return nil
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *notification.Notification) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) CreateBatch(ctx context.Context, notifications []*notification.Notification) error {
	return errors.New("not implemented")
}

func (m *mockNotificationRepo) GetByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error) {
	return nil, errors.New("not implemented")
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

type mockDeliveryAttemptRepo struct {
	createFn func(ctx context.Context, da *notification.DeliveryAttempt) error
}

func (m *mockDeliveryAttemptRepo) Create(ctx context.Context, da *notification.DeliveryAttempt) error {
	if m.createFn != nil {
		return m.createFn(ctx, da)
	}
	return nil
}

func (m *mockDeliveryAttemptRepo) GetByNotificationID(ctx context.Context, notificationID string) ([]*notification.DeliveryAttempt, error) {
	return nil, errors.New("not implemented")
}

type mockRateLimiter struct {
	allowFn func(ctx context.Context, channel notification.Channel) (bool, error)
}

func (m *mockRateLimiter) Allow(ctx context.Context, channel notification.Channel) (bool, error) {
	if m.allowFn != nil {
		return m.allowFn(ctx, channel)
	}
	return true, nil
}

type mockDeliveryClient struct {
	deliverFn func(ctx context.Context, req *port.DeliveryRequest) (*port.DeliveryResponse, int, error)
}

func (m *mockDeliveryClient) Deliver(ctx context.Context, req *port.DeliveryRequest) (*port.DeliveryResponse, int, error) {
	if m.deliverFn != nil {
		return m.deliverFn(ctx, req)
	}
	return &port.DeliveryResponse{MessageID: "test-msg-id", Status: "accepted", Timestamp: time.Now().Format(time.RFC3339)}, 202, nil
}

type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...port.Field)  {}
func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...port.Field)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, fields ...port.Field) {}

func TestExecute_Success(t *testing.T) {
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test message",
				Priority:  notification.PriorityHigh,
				Status:    notification.StatusPending,
			}, nil
		},
	}

	attemptRepo := &mockDeliveryAttemptRepo{}
	rateLimiter := &mockRateLimiter{}
	deliveryClient := &mockDeliveryClient{}
	logger := &mockLogger{}

	uc := NewUseCase(notifRepo, attemptRepo, rateLimiter, deliveryClient, logger)

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestExecute_NotificationNotFound(t *testing.T) {
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return nil, errors.New("not found")
		},
	}

	uc := NewUseCase(notifRepo, &mockDeliveryAttemptRepo{}, &mockRateLimiter{}, &mockDeliveryClient{}, &mockLogger{})

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestExecute_AlreadyTerminal(t *testing.T) {
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test message",
				Status:    notification.StatusSent,
			}, nil
		},
	}

	uc := NewUseCase(notifRepo, &mockDeliveryAttemptRepo{}, &mockRateLimiter{}, &mockDeliveryClient{}, &mockLogger{})

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err != nil {
		t.Errorf("expected no error for terminal status, got %v", err)
	}
}

func TestExecute_RateLimitExceeded(t *testing.T) {
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test message",
				Status:    notification.StatusPending,
			}, nil
		},
	}

	rateLimiter := &mockRateLimiter{
		allowFn: func(ctx context.Context, channel notification.Channel) (bool, error) {
			return false, nil
		},
	}

	uc := NewUseCase(notifRepo, &mockDeliveryAttemptRepo{}, rateLimiter, &mockDeliveryClient{}, &mockLogger{})

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err == nil {
		t.Error("expected rate limit error, got nil")
	}
}

func TestExecute_RateLimiterError(t *testing.T) {
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test message",
				Status:    notification.StatusPending,
			}, nil
		},
	}

	rateLimiter := &mockRateLimiter{
		allowFn: func(ctx context.Context, channel notification.Channel) (bool, error) {
			return false, errors.New("redis error")
		},
	}

	uc := NewUseCase(notifRepo, &mockDeliveryAttemptRepo{}, rateLimiter, &mockDeliveryClient{}, &mockLogger{})

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestExecute_DeliveryFailureRetries(t *testing.T) {
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test message",
				Status:    notification.StatusPending,
			}, nil
		},
	}

	attemptCount := 0
	deliveryClient := &mockDeliveryClient{
		deliverFn: func(ctx context.Context, req *port.DeliveryRequest) (*port.DeliveryResponse, int, error) {
			attemptCount++
			if attemptCount < 3 {
				return nil, 500, errors.New("delivery failed")
			}
			return &port.DeliveryResponse{MessageID: "success", Status: "accepted", Timestamp: time.Now().Format(time.RFC3339)}, 202, nil
		},
	}

	uc := NewUseCase(notifRepo, &mockDeliveryAttemptRepo{}, &mockRateLimiter{}, deliveryClient, &mockLogger{})

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err != nil {
		t.Errorf("expected success after retries, got %v", err)
	}
	if attemptCount != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}
}

func TestExecute_AllDeliveryAttemptsFail(t *testing.T) {
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test message",
				Status:    notification.StatusPending,
			}, nil
		},
	}

	attemptCount := 0
	deliveryClient := &mockDeliveryClient{
		deliverFn: func(ctx context.Context, req *port.DeliveryRequest) (*port.DeliveryResponse, int, error) {
			attemptCount++
			return nil, 500, errors.New("delivery failed")
		},
	}

	attemptsSaved := 0
	attemptRepo := &mockDeliveryAttemptRepo{
		createFn: func(ctx context.Context, da *notification.DeliveryAttempt) error {
			attemptsSaved++
			return nil
		},
	}

	uc := NewUseCase(notifRepo, attemptRepo, &mockRateLimiter{}, deliveryClient, &mockLogger{})

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err == nil {
		t.Error("expected error after all attempts fail, got nil")
	}
	if attemptCount != 5 {
		t.Errorf("expected 5 attempts, got %d", attemptCount)
	}
	if attemptsSaved != 5 {
		t.Errorf("expected 5 attempts saved, got %d", attemptsSaved)
	}
}

func TestExecute_DeliveryAttemptSaveError(t *testing.T) {
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test message",
				Status:    notification.StatusPending,
			}, nil
		},
	}

	attemptRepo := &mockDeliveryAttemptRepo{
		createFn: func(ctx context.Context, da *notification.DeliveryAttempt) error {
			return errors.New("db error")
		},
	}

	uc := NewUseCase(notifRepo, attemptRepo, &mockRateLimiter{}, &mockDeliveryClient{}, &mockLogger{})

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err != nil {
		t.Errorf("should not fail even if attempt save fails, got %v", err)
	}
}

func TestExecute_UpdateStatusError(t *testing.T) {
	statusUpdateCalled := false
	notifRepo := &mockNotificationRepo{
		getByIDFn: func(ctx context.Context, id string) (*notification.Notification, error) {
			return &notification.Notification{
				ID:        id,
				Recipient: "+905551234567",
				Channel:   notification.ChannelSMS,
				Content:   "Test message",
				Status:    notification.StatusPending,
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id string, status notification.Status, sentAt *time.Time, reason *string) error {
			statusUpdateCalled = true
			return errors.New("update error")
		},
	}

	uc := NewUseCase(notifRepo, &mockDeliveryAttemptRepo{}, &mockRateLimiter{}, &mockDeliveryClient{}, &mockLogger{})

	cmd := &Command{NotificationID: "test-id"}
	err := uc.Execute(context.Background(), cmd)

	if err != nil {
		t.Errorf("should not fail even if status update fails, got %v", err)
	}
	if !statusUpdateCalled {
		t.Error("expected status update to be called")
	}
}
