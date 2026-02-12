package list

import (
	"time"

	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

type Query struct {
	Status   *notification.Status
	Channel  *notification.Channel
	FromTime *time.Time
	ToTime   *time.Time
	BatchID  *string
	Limit    int
	Offset   int
}
