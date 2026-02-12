package create

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

type UseCase struct {
	repo  port.NotificationRepository
	batch port.BatchRepository
	pub   port.EventPublisher
	idem  port.IdempotencyStore
	log   port.Logger
}

func NewUseCase(
	repo port.NotificationRepository,
	batch port.BatchRepository,
	pub port.EventPublisher,
	idem port.IdempotencyStore,
	log port.Logger,
) *UseCase {
	return &UseCase{
		repo:  repo,
		batch: batch,
		pub:   pub,
		idem:  idem,
		log:   log,
	}
}

// CreateNotification creates one notification.
func (u *UseCase) CreateNotification(ctx context.Context, cmd *Command) (*notification.Notification, error) {
	ch := notification.Channel(cmd.Channel)
	if !ch.Valid() {
		u.log.Warn(ctx, "invalid channel", port.F("channel", cmd.Channel))
		return nil, notification.ErrInvalidChannel
	}
	pr := notification.Priority(cmd.Priority)
	if !pr.Valid() {
		u.log.Warn(ctx, "invalid priority", port.F("priority", cmd.Priority))
		return nil, notification.ErrInvalidPriority
	}
	if len(cmd.Content) > notification.MaxContentLength(ch) || len(cmd.Content) == 0 {
		u.log.Warn(ctx, "invalid content", port.F("content_len", len(cmd.Content)), port.F("channel", cmd.Channel))
		return nil, notification.ErrInvalidContent
	}
	if len(cmd.Recipient) == 0 || len(cmd.Recipient) > notification.MaxRecipientLength {
		u.log.Warn(ctx, "invalid recipient", port.F("recipient_len", len(cmd.Recipient)))
		return nil, notification.ErrInvalidContent
	}
	// Idempotency check: Redis first (fast), DB fallback (guarantee)
	if cmd.IdempotencyKey != nil && *cmd.IdempotencyKey != "" {
		set, err := u.idem.SetIfNotExists(ctx, *cmd.IdempotencyKey, 7*24*3600)
		if err != nil {
			u.log.Warn(ctx, "redis idempotency check failed, falling back to DB",
				port.F("error", err), port.F("key", *cmd.IdempotencyKey))
			exists, dbErr := u.repo.ExistsByIdempotencyKey(ctx, *cmd.IdempotencyKey)
			if dbErr != nil {
				u.log.Error(ctx, "db idempotency check failed", port.F("error", dbErr), port.F("key", *cmd.IdempotencyKey))
				return nil, dbErr
			}
			if exists {
				u.log.Warn(ctx, "duplicate idempotency key (db fallback)", port.F("key", *cmd.IdempotencyKey))
				return nil, notification.ErrDuplicateRequest
			}
		} else if !set {
			u.log.Warn(ctx, "duplicate idempotency key (redis)", port.F("key", *cmd.IdempotencyKey))
			return nil, notification.ErrDuplicateRequest
		}
	}

	now := time.Now()
	id := uuid.New().String()

	n := &notification.Notification{
		ID:             id,
		Recipient:      cmd.Recipient,
		Channel:        ch,
		Content:        cmd.Content,
		Priority:       pr,
		Status:         notification.StatusPending,
		IdempotencyKey: cmd.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := u.repo.Create(ctx, n); err != nil {
		if cmd.IdempotencyKey != nil && *cmd.IdempotencyKey != "" && isUniqueViolation(err) {
			u.log.Warn(ctx, "duplicate idempotency key (db constraint)", port.F("key", *cmd.IdempotencyKey))
			return nil, notification.ErrDuplicateRequest
		}
		u.log.Error(ctx, "failed to create notification in DB", port.F("error", err), port.F("notification_id", id))
		return nil, err
	}

	u.log.Info(ctx, "notification created", port.F("notification_id", id), port.F("channel", ch), port.F("priority", pr))

	evt := &port.NotificationEvent{
		NotificationID: id,
		Recipient:      cmd.Recipient,
		Channel:        ch,
		Content:        cmd.Content,
		Priority:       pr,
		IdempotencyKey: cmd.IdempotencyKey,
		CreatedAt:      now.Format("2006-01-02T15:04:05Z07:00"),
	}

	if err := u.pub.Publish(ctx, evt); err != nil {
		u.log.Error(ctx, "failed to publish notification event", port.F("error", err), port.F("notification_id", id))
		return n, err
	}

	if err := u.repo.UpdateStatus(ctx, id, notification.StatusQueued, nil, nil); err != nil {
		u.log.Error(ctx, "failed to update status to queued", port.F("error", err), port.F("notification_id", id))
	}
	n.Status = notification.StatusQueued
	u.log.Info(ctx, "notification event published", port.F("notification_id", id))

	return n, nil
}

// CreateNotificationBatches creates a batch and its notifications.
func (u *UseCase) CreateNotificationBatches(ctx context.Context, cmd *BatchCommand) (*BatchResult, error) {
	if len(cmd.Items) == 0 || len(cmd.Items) > notification.MaxBatchSize {
		u.log.Warn(ctx, "batch size invalid", port.F("size", len(cmd.Items)))
		return nil, notification.ErrBatchTooLarge
	}

	batchID := uuid.New().String()
	now := time.Now()

	b := &notification.Batch{
		ID:             batchID,
		IdempotencyKey: cmd.IdempotencyKey,
		CreatedAt:      now,
	}

	if err := u.batch.Create(ctx, b); err != nil {
		u.log.Error(ctx, "failed to create batch", port.F("error", err), port.F("batch_id", batchID))
		return nil, err
	}

	u.log.Info(ctx, "batch created", port.F("batch_id", batchID), port.F("item_count", len(cmd.Items)))

	var notifications []*notification.Notification
	var events []*port.NotificationEvent
	skipped := 0

	// First pass: validate and build notification entities
	for _, item := range cmd.Items {
		ch := notification.Channel(item.Channel)
		if !ch.Valid() {
			skipped++
			continue
		}
		pr := notification.Priority(item.Priority)
		if !pr.Valid() {
			pr = notification.PriorityNormal
		}
		if len(item.Content) > notification.MaxContentLength(ch) || len(item.Content) == 0 || len(item.Recipient) == 0 {
			skipped++
			continue
		}

		id := uuid.New().String()
		n := &notification.Notification{
			ID:        id,
			BatchID:   &batchID,
			Recipient: item.Recipient,
			Channel:   ch,
			Content:   item.Content,
			Priority:  pr,
			Status:    notification.StatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}

		notifications = append(notifications, n)
		events = append(events, &port.NotificationEvent{
			NotificationID: id,
			BatchID:        &batchID,
			Recipient:      item.Recipient,
			Channel:        ch,
			Content:        item.Content,
			Priority:       pr,
			CreatedAt:      now.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Second pass: bulk insert all valid notifications
	if len(notifications) > 0 {
		if err := u.repo.CreateBatch(ctx, notifications); err != nil {
			u.log.Error(ctx, "failed to create batch notifications in DB", port.F("error", err), port.F("batch_id", batchID))
			return nil, err
		}
	}

	if skipped > 0 {
		u.log.Warn(ctx, "some batch items skipped", port.F("batch_id", batchID), port.F("skipped", skipped))
	}

	if err := u.pub.PublishBatch(ctx, events); err != nil {
		u.log.Error(ctx, "failed to publish batch events", port.F("error", err), port.F("batch_id", batchID))
		return &BatchResult{BatchID: batchID, Notifications: notifications}, err
	}

	for _, n := range notifications {
		if err := u.repo.UpdateStatus(ctx, n.ID, notification.StatusQueued, nil, nil); err != nil {
			u.log.Error(ctx, "failed to update status to queued", port.F("error", err), port.F("notification_id", n.ID))
		}
		n.Status = notification.StatusQueued
	}

	u.log.Info(ctx, "batch events published", port.F("batch_id", batchID), port.F("notification_count", len(notifications)))
	return &BatchResult{BatchID: batchID, Notifications: notifications}, nil
}

type BatchResult struct {
	BatchID       string
	Notifications []*notification.Notification
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	// PostgreSQL unique violation error code (23505)
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "23505")
}
