package postgres

import (
	"context"
	"time"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
	"gorm.io/gorm"
)

var _ port.NotificationRepository = (*NotificationRepository)(nil)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *notification.Notification) error {
	m := toNotificationModel(n)
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *NotificationRepository) CreateBatch(ctx context.Context, notifications []*notification.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	models := make([]*NotificationModel, len(notifications))
	for i, n := range notifications {
		models[i] = toNotificationModel(n)
	}

	// GORM CreateInBatches: inserts in chunks of 100 to avoid parameter limits
	return r.db.WithContext(ctx).CreateInBatches(models, 100).Error
}

func (r *NotificationRepository) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	var m NotificationModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, notification.ErrNotFound
		}
		return nil, err
	}
	return toNotificationDomain(&m), nil
}

func (r *NotificationRepository) GetByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error) {
	var list []NotificationModel
	err := r.db.WithContext(ctx).Where("batch_id = ?", batchID).Order("created_at").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]*notification.Notification, len(list))
	for i := range list {
		out[i] = toNotificationDomain(&list[i])
	}
	return out, nil
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id string, status notification.Status, sentAt *time.Time, failureReason *string) error {
	updates := map[string]interface{}{
		"status":     status.String(),
		"updated_at": time.Now(),
	}
	if sentAt != nil {
		updates["sent_at"] = sentAt
	}
	if failureReason != nil {
		updates["failure_reason"] = failureReason
	}
	res := r.db.WithContext(ctx).Model(&NotificationModel{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return notification.ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) List(ctx context.Context, filter port.ListFilter) (*port.ListResult, error) {
	q := r.db.WithContext(ctx).Model(&NotificationModel{})
	if filter.Status != nil {
		q = q.Where("status = ?", filter.Status.String())
	}
	if filter.Channel != nil {
		q = q.Where("channel = ?", filter.Channel.String())
	}
	if filter.FromTime != nil {
		q = q.Where("created_at >= ?", *filter.FromTime)
	}
	if filter.ToTime != nil {
		q = q.Where("created_at <= ?", *filter.ToTime)
	}
	if filter.BatchID != nil {
		q = q.Where("batch_id = ?", *filter.BatchID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	var list []NotificationModel
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*notification.Notification, len(list))
	for i := range list {
		out[i] = toNotificationDomain(&list[i])
	}
	return &port.ListResult{Notifications: out, Total: int(total)}, nil
}

func (r *NotificationRepository) CancelPending(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Model(&NotificationModel{}).
		Where("id = ? AND status IN ?", id, []string{"pending", "queued"}).
		Updates(map[string]interface{}{"status": "cancelled", "updated_at": time.Now()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return notification.ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) CancelPendingByBatchID(ctx context.Context, batchID string) (int, error) {
	res := r.db.WithContext(ctx).Model(&NotificationModel{}).
		Where("batch_id = ? AND status IN ?", batchID, []string{"pending", "queued"}).
		Updates(map[string]interface{}{"status": "cancelled", "updated_at": time.Now()})
	return int(res.RowsAffected), res.Error
}

func (r *NotificationRepository) ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&NotificationModel{}).Where("idempotency_key = ?", key).Limit(1).Count(&n).Error
	return n > 0, err
}

func toNotificationModel(n *notification.Notification) *NotificationModel {
	m := &NotificationModel{
		ID:        n.ID,
		Recipient: n.Recipient,
		Channel:   n.Channel.String(),
		Content:   n.Content,
		Priority:  n.Priority.String(),
		Status:    n.Status.String(),
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
	m.BatchID = n.BatchID
	m.IdempotencyKey = n.IdempotencyKey
	m.SentAt = n.SentAt
	m.FailureReason = n.FailureReason
	return m
}

func toNotificationDomain(m *NotificationModel) *notification.Notification {
	n := &notification.Notification{
		ID:        m.ID,
		Recipient: m.Recipient,
		Channel:   notification.Channel(m.Channel),
		Content:   m.Content,
		Priority:  notification.Priority(m.Priority),
		Status:    notification.Status(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	n.BatchID = m.BatchID
	n.IdempotencyKey = m.IdempotencyKey
	n.SentAt = m.SentAt
	n.FailureReason = m.FailureReason
	return n
}
