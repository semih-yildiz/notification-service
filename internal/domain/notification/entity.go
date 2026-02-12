package notification

import "time"

// Notification is the core aggregate: one notification request.
type Notification struct {
	ID             string
	BatchID        *string
	Recipient      string
	Channel        Channel
	Content        string
	Priority       Priority
	Status         Status
	IdempotencyKey *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	SentAt         *time.Time
	FailureReason  *string
}

// Batch represents a batch of notifications (up to MaxBatchSize).
type Batch struct {
	ID             string
	IdempotencyKey *string
	CreatedAt      time.Time
}

// DeliveryAttempt records one delivery attempt (retry and observability).
type DeliveryAttempt struct {
	ID             string
	NotificationID string
	AttemptNumber  int
	Success        bool
	StatusCode     int
	ResponseBody   string
	ErrorMessage   *string
	CreatedAt      time.Time
}
