package notification

// Priority controls queue ordering (high, normal, low).
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

func (p Priority) Valid() bool {
	switch p {
	case PriorityHigh, PriorityNormal, PriorityLow:
		return true
	default:
		return false
	}
}

func (p Priority) String() string { return string(p) }

// RabbitMQPriority returns numeric priority for RabbitMQ (high=3, normal=2, low=1).
func (p Priority) RabbitMQPriority() uint8 {
	switch p {
	case PriorityHigh:
		return 3
	case PriorityNormal:
		return 2
	case PriorityLow:
		return 1
	default:
		return 2
	}
}
