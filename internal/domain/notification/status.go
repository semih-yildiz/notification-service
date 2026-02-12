package notification

// Status is the lifecycle state of a notification.
type Status string

const (
	StatusPending   Status = "pending"   // created, not yet queued
	StatusQueued    Status = "queued"    // published to queue
	StatusSent      Status = "sent"      // delivered successfully
	StatusFailed    Status = "failed"    // delivery failed after retries
	StatusCancelled Status = "cancelled" // cancelled before/during processing
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusQueued, StatusSent, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

func (s Status) String() string { return string(s) }

// Terminal returns true if no further processing should occur.
func (s Status) Terminal() bool {
	return s == StatusSent || s == StatusFailed || s == StatusCancelled
}
