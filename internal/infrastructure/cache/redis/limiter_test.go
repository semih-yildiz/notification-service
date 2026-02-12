package redis

import (
	"testing"

	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

func TestRateLimitConstants(t *testing.T) {
	if rateLimitMaxPerSecond != 100 {
		t.Errorf("expected rateLimitMaxPerSecond=100, got %d", rateLimitMaxPerSecond)
	}

	if rateLimitKeyPrefix != "ratelimit:channel:" {
		t.Errorf("expected rateLimitKeyPrefix='ratelimit:channel:', got '%s'", rateLimitKeyPrefix)
	}
}

func TestRateLimitLogic(t *testing.T) {
	tests := []struct {
		name     string
		count    int64
		expected bool
	}{
		{"Under limit", 50, true},
		{"At limit", 100, true},
		{"Just over limit", 101, false},
		{"Way over limit", 200, false},
		{"Zero", 0, true},
		{"One", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.count <= rateLimitMaxPerSecond
			if allowed != tt.expected {
				t.Errorf("count=%d: expected allowed=%v, got %v", tt.count, tt.expected, allowed)
			}
		})
	}
}

func TestRateLimitKeyGeneration(t *testing.T) {
	tests := []struct {
		name    string
		channel notification.Channel
		prefix  string
	}{
		{"SMS channel", notification.ChannelSMS, "ratelimit:channel:sms"},
		{"Email channel", notification.ChannelEmail, "ratelimit:channel:email"},
		{"Push channel", notification.ChannelPush, "ratelimit:channel:push"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := rateLimitKeyPrefix + tt.channel.String()
			if key != tt.prefix {
				t.Errorf("expected key='%s', got '%s'", tt.prefix, key)
			}
		})
	}
}

func TestRateLimitWindow(t *testing.T) {
	if rateLimitWindow.Seconds() != 1.0 {
		t.Errorf("expected rateLimitWindow=1s, got %v", rateLimitWindow)
	}
}
