package dto

import (
	"testing"
)

func TestNotificationItem_Validate_Success(t *testing.T) {
	item := &NotificationItem{
		Recipient: "+905551234567",
		Channel:   "sms",
		Content:   "Test message",
		Priority:  "high",
	}

	err := item.Validate()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNotificationItem_Validate_DefaultPriority(t *testing.T) {
	item := &NotificationItem{
		Recipient: "+905551234567",
		Channel:   "email",
		Content:   "Test message",
		Priority:  "",
	}

	err := item.Validate()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if item.Priority != "normal" {
		t.Errorf("expected priority to default to 'normal', got %s", item.Priority)
	}
}

func TestNotificationItem_Validate_MissingRecipient(t *testing.T) {
	item := &NotificationItem{
		Recipient: "",
		Channel:   "sms",
		Content:   "Test message",
		Priority:  "high",
	}

	err := item.Validate()

	if err == nil {
		t.Error("expected validation error for missing recipient")
	}
}

func TestNotificationItem_Validate_MissingChannel(t *testing.T) {
	item := &NotificationItem{
		Recipient: "+905551234567",
		Channel:   "",
		Content:   "Test message",
		Priority:  "high",
	}

	err := item.Validate()

	if err == nil {
		t.Error("expected validation error for missing channel")
	}
}

func TestNotificationItem_Validate_InvalidChannel(t *testing.T) {
	item := &NotificationItem{
		Recipient: "+905551234567",
		Channel:   "invalid",
		Content:   "Test message",
		Priority:  "high",
	}

	err := item.Validate()

	if err == nil {
		t.Error("expected validation error for invalid channel")
	}
}

func TestNotificationItem_Validate_MissingContent(t *testing.T) {
	item := &NotificationItem{
		Recipient: "+905551234567",
		Channel:   "sms",
		Content:   "",
		Priority:  "high",
	}

	err := item.Validate()

	if err == nil {
		t.Error("expected validation error for missing content")
	}
}

func TestNotificationItem_Validate_InvalidPriority(t *testing.T) {
	item := &NotificationItem{
		Recipient: "+905551234567",
		Channel:   "sms",
		Content:   "Test message",
		Priority:  "invalid",
	}

	err := item.Validate()

	if err == nil {
		t.Error("expected validation error for invalid priority")
	}
}

func TestNotificationItem_Validate_MultipleErrors(t *testing.T) {
	item := &NotificationItem{
		Recipient: "",
		Channel:   "invalid",
		Content:   "",
		Priority:  "invalid",
	}

	err := item.Validate()

	if err == nil {
		t.Error("expected validation error for multiple errors")
	}
}

func TestNotificationItem_Validate_AllChannels(t *testing.T) {
	channels := []string{"sms", "email", "push"}

	for _, channel := range channels {
		t.Run(channel, func(t *testing.T) {
			item := &NotificationItem{
				Recipient: "+905551234567",
				Channel:   channel,
				Content:   "Test message",
				Priority:  "high",
			}

			err := item.Validate()

			if err != nil {
				t.Errorf("expected no error for channel %s, got %v", channel, err)
			}
		})
	}
}

func TestNotificationItem_Validate_AllPriorities(t *testing.T) {
	priorities := []string{"high", "normal", "low"}

	for _, priority := range priorities {
		t.Run(priority, func(t *testing.T) {
			item := &NotificationItem{
				Recipient: "+905551234567",
				Channel:   "sms",
				Content:   "Test message",
				Priority:  priority,
			}

			err := item.Validate()

			if err != nil {
				t.Errorf("expected no error for priority %s, got %v", priority, err)
			}
		})
	}
}

func TestNotificationItem_Validate_WithIdempotencyKey(t *testing.T) {
	key := "test-key-123"
	item := &NotificationItem{
		Recipient:      "+905551234567",
		Channel:        "sms",
		Content:        "Test message",
		Priority:       "high",
		IdempotencyKey: &key,
	}

	err := item.Validate()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
