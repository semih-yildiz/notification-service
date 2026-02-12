package dto

import "fmt"

// NotificationItem represents a single notification (used in both single and batch requests).
type NotificationItem struct {
	Recipient      string  `json:"recipient"`
	Channel        string  `json:"channel"`
	Content        string  `json:"content"`
	Priority       string  `json:"priority"`
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

func (item *NotificationItem) Validate() error {
	var validationErrors []ValidationError

	if item.Recipient == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "recipient",
			Message: "recipient is required",
		})
	}

	if item.Channel == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "channel",
			Message: "channel is required",
		})
	} else if item.Channel != "sms" && item.Channel != "email" && item.Channel != "push" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "channel",
			Message: "channel must be one of: sms, email, push",
		})
	}

	if item.Content == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "content",
			Message: "content is required",
		})
	}

	// Set default priority if not provided
	if item.Priority == "" {
		item.Priority = "normal"
	} else if item.Priority != "high" && item.Priority != "normal" && item.Priority != "low" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "priority",
			Message: "priority must be one of: high, normal, low",
		})
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %d errors", len(validationErrors))
	}

	return nil
}
