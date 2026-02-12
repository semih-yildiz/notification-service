package create

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

type mockNotificationRepo struct {
	createFn                 func(ctx context.Context, n *notification.Notification) error
	createBatchFn            func(ctx context.Context, notifications []*notification.Notification) error
	updateStatusFn           func(ctx context.Context, id string, status notification.Status, sentAt *time.Time, reason *string) error
	existsByIdempotencyKeyFn func(ctx context.Context, key string) (bool, error)
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *notification.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, n)
	}
	return nil
}

func (m *mockNotificationRepo) CreateBatch(ctx context.Context, notifications []*notification.Notification) error {
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, notifications)
	}
	return nil
}

func (m *mockNotificationRepo) UpdateStatus(ctx context.Context, id string, status notification.Status, sentAt *time.Time, reason *string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, sentAt, reason)
	}
	return nil
}

func (m *mockNotificationRepo) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	if m.existsByIdempotencyKeyFn != nil {
		return m.existsByIdempotencyKeyFn(ctx, key)
	}
	return false, nil
}

func (m *mockNotificationRepo) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	return nil, errors.New("not implemented")
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

type mockBatchRepo struct {
	createFn func(ctx context.Context, b *notification.Batch) error
}

func (m *mockBatchRepo) Create(ctx context.Context, b *notification.Batch) error {
	if m.createFn != nil {
		return m.createFn(ctx, b)
	}
	return nil
}

func (m *mockBatchRepo) GetByID(ctx context.Context, id string) (*notification.Batch, error) {
	return nil, errors.New("not implemented")
}

func (m *mockBatchRepo) GetNotificationsByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error) {
	return nil, errors.New("not implemented")
}

type mockPublisher struct {
	publishFn      func(ctx context.Context, evt *port.NotificationEvent) error
	publishBatchFn func(ctx context.Context, events []*port.NotificationEvent) error
}

func (m *mockPublisher) Publish(ctx context.Context, evt *port.NotificationEvent) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, evt)
	}
	return nil
}

func (m *mockPublisher) PublishBatch(ctx context.Context, events []*port.NotificationEvent) error {
	if m.publishBatchFn != nil {
		return m.publishBatchFn(ctx, events)
	}
	return nil
}

type mockIdempotencyStore struct {
	setIfNotExistsFn func(ctx context.Context, key string, ttl int) (bool, error)
	existsFn         func(ctx context.Context, key string) (bool, error)
}

func (m *mockIdempotencyStore) SetIfNotExists(ctx context.Context, key string, ttl int) (bool, error) {
	if m.setIfNotExistsFn != nil {
		return m.setIfNotExistsFn(ctx, key, ttl)
	}
	return true, nil
}

func (m *mockIdempotencyStore) Exists(ctx context.Context, key string) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, key)
	}
	return false, nil
}

type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...port.Field)  {}
func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...port.Field)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, fields ...port.Field) {}

func TestCreateNotification_Success(t *testing.T) {
	repo := &mockNotificationRepo{}
	batch := &mockBatchRepo{}
	pub := &mockPublisher{}
	idem := &mockIdempotencyStore{}
	log := &mockLogger{}

	uc := NewUseCase(repo, batch, pub, idem, log)

	cmd := &Command{
		Recipient: "+905551234567",
		Channel:   "sms",
		Content:   "Test message",
		Priority:  "high",
	}

	result, err := uc.CreateNotification(context.Background(), cmd)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Status != notification.StatusQueued {
		t.Errorf("expected status %s, got %s", notification.StatusQueued, result.Status)
	}
}

func TestCreateNotification_InvalidChannel(t *testing.T) {
	uc := NewUseCase(&mockNotificationRepo{}, &mockBatchRepo{}, &mockPublisher{}, &mockIdempotencyStore{}, &mockLogger{})

	cmd := &Command{
		Recipient: "+905551234567",
		Channel:   "invalid",
		Content:   "Test message",
		Priority:  "high",
	}

	_, err := uc.CreateNotification(context.Background(), cmd)

	if !errors.Is(err, notification.ErrInvalidChannel) {
		t.Errorf("expected ErrInvalidChannel, got %v", err)
	}
}

func TestCreateNotification_InvalidPriority(t *testing.T) {
	uc := NewUseCase(&mockNotificationRepo{}, &mockBatchRepo{}, &mockPublisher{}, &mockIdempotencyStore{}, &mockLogger{})

	cmd := &Command{
		Recipient: "+905551234567",
		Channel:   "sms",
		Content:   "Test message",
		Priority:  "invalid",
	}

	_, err := uc.CreateNotification(context.Background(), cmd)

	if !errors.Is(err, notification.ErrInvalidPriority) {
		t.Errorf("expected ErrInvalidPriority, got %v", err)
	}
}

func TestCreateNotification_EmptyContent(t *testing.T) {
	uc := NewUseCase(&mockNotificationRepo{}, &mockBatchRepo{}, &mockPublisher{}, &mockIdempotencyStore{}, &mockLogger{})

	cmd := &Command{
		Recipient: "+905551234567",
		Channel:   "sms",
		Content:   "",
		Priority:  "high",
	}

	_, err := uc.CreateNotification(context.Background(), cmd)

	if !errors.Is(err, notification.ErrInvalidContent) {
		t.Errorf("expected ErrInvalidContent, got %v", err)
	}
}

func TestCreateNotification_EmptyRecipient(t *testing.T) {
	uc := NewUseCase(&mockNotificationRepo{}, &mockBatchRepo{}, &mockPublisher{}, &mockIdempotencyStore{}, &mockLogger{})

	cmd := &Command{
		Recipient: "",
		Channel:   "sms",
		Content:   "Test message",
		Priority:  "high",
	}

	_, err := uc.CreateNotification(context.Background(), cmd)

	if !errors.Is(err, notification.ErrInvalidContent) {
		t.Errorf("expected ErrInvalidContent, got %v", err)
	}
}

func TestCreateNotification_DuplicateIdempotencyKey_Redis(t *testing.T) {
	idem := &mockIdempotencyStore{
		setIfNotExistsFn: func(ctx context.Context, key string, ttl int) (bool, error) {
			return false, nil
		},
	}

	uc := NewUseCase(&mockNotificationRepo{}, &mockBatchRepo{}, &mockPublisher{}, idem, &mockLogger{})

	key := "test-key-123"
	cmd := &Command{
		Recipient:      "+905551234567",
		Channel:        "sms",
		Content:        "Test message",
		Priority:       "high",
		IdempotencyKey: &key,
	}

	_, err := uc.CreateNotification(context.Background(), cmd)

	if !errors.Is(err, notification.ErrDuplicateRequest) {
		t.Errorf("expected ErrDuplicateRequest, got %v", err)
	}
}

func TestCreateNotification_DuplicateIdempotencyKey_DBFallback(t *testing.T) {
	idem := &mockIdempotencyStore{
		setIfNotExistsFn: func(ctx context.Context, key string, ttl int) (bool, error) {
			return false, errors.New("redis error")
		},
	}

	repo := &mockNotificationRepo{
		existsByIdempotencyKeyFn: func(ctx context.Context, key string) (bool, error) {
			return true, nil
		},
	}

	uc := NewUseCase(repo, &mockBatchRepo{}, &mockPublisher{}, idem, &mockLogger{})

	key := "test-key-123"
	cmd := &Command{
		Recipient:      "+905551234567",
		Channel:        "sms",
		Content:        "Test message",
		Priority:       "high",
		IdempotencyKey: &key,
	}

	_, err := uc.CreateNotification(context.Background(), cmd)

	if !errors.Is(err, notification.ErrDuplicateRequest) {
		t.Errorf("expected ErrDuplicateRequest, got %v", err)
	}
}

func TestCreateNotification_PublishError(t *testing.T) {
	pub := &mockPublisher{
		publishFn: func(ctx context.Context, evt *port.NotificationEvent) error {
			return errors.New("publish failed")
		},
	}

	uc := NewUseCase(&mockNotificationRepo{}, &mockBatchRepo{}, pub, &mockIdempotencyStore{}, &mockLogger{})

	cmd := &Command{
		Recipient: "+905551234567",
		Channel:   "sms",
		Content:   "Test message",
		Priority:  "high",
	}

	result, err := uc.CreateNotification(context.Background(), cmd)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result == nil {
		t.Error("expected result even with publish error")
	}
}

func TestCreateNotificationBatches_Success(t *testing.T) {
	repo := &mockNotificationRepo{}
	batch := &mockBatchRepo{}
	pub := &mockPublisher{}

	uc := NewUseCase(repo, batch, pub, &mockIdempotencyStore{}, &mockLogger{})

	cmd := &BatchCommand{
		Items: []BatchItem{
			{Recipient: "+905551234567", Channel: "sms", Content: "Test 1", Priority: "high"},
			{Recipient: "+905551234568", Channel: "email", Content: "Test 2", Priority: "normal"},
		},
	}

	result, err := uc.CreateNotificationBatches(context.Background(), cmd)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.Notifications) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(result.Notifications))
	}
}

func TestCreateNotificationBatches_EmptyBatch(t *testing.T) {
	uc := NewUseCase(&mockNotificationRepo{}, &mockBatchRepo{}, &mockPublisher{}, &mockIdempotencyStore{}, &mockLogger{})

	cmd := &BatchCommand{Items: []BatchItem{}}

	_, err := uc.CreateNotificationBatches(context.Background(), cmd)

	if !errors.Is(err, notification.ErrBatchTooLarge) {
		t.Errorf("expected ErrBatchTooLarge, got %v", err)
	}
}

func TestCreateNotificationBatches_TooLarge(t *testing.T) {
	uc := NewUseCase(&mockNotificationRepo{}, &mockBatchRepo{}, &mockPublisher{}, &mockIdempotencyStore{}, &mockLogger{})

	items := make([]BatchItem, 1001)
	for i := range items {
		items[i] = BatchItem{Recipient: "+90555", Channel: "sms", Content: "Test", Priority: "normal"}
	}

	cmd := &BatchCommand{Items: items}

	_, err := uc.CreateNotificationBatches(context.Background(), cmd)

	if !errors.Is(err, notification.ErrBatchTooLarge) {
		t.Errorf("expected ErrBatchTooLarge, got %v", err)
	}
}

func TestCreateNotificationBatches_SkipsInvalidItems(t *testing.T) {
	repo := &mockNotificationRepo{}
	batch := &mockBatchRepo{}
	pub := &mockPublisher{}

	uc := NewUseCase(repo, batch, pub, &mockIdempotencyStore{}, &mockLogger{})

	cmd := &BatchCommand{
		Items: []BatchItem{
			{Recipient: "+905551234567", Channel: "sms", Content: "Valid", Priority: "high"},
			{Recipient: "+905551234568", Channel: "invalid", Content: "Invalid channel", Priority: "high"},
			{Recipient: "", Channel: "sms", Content: "Empty recipient", Priority: "high"},
		},
	}

	result, err := uc.CreateNotificationBatches(context.Background(), cmd)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result.Notifications) != 1 {
		t.Errorf("expected 1 valid notification, got %d", len(result.Notifications))
	}
}

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"unique constraint", errors.New("unique constraint violation"), true},
		{"duplicate key", errors.New("duplicate key value"), true},
		{"23505 code", errors.New("ERROR: duplicate key value violates unique constraint (23505)"), true},
		{"other error", errors.New("connection timeout"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUniqueViolation(tt.err)
			if result != tt.expected {
				t.Errorf("isUniqueViolation(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}
