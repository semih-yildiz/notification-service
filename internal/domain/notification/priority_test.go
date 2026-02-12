package notification

import "testing"

func TestPriority_Valid(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     bool
	}{
		{"High priority", PriorityHigh, true},
		{"Normal priority", PriorityNormal, true},
		{"Low priority", PriorityLow, true},
		{"Invalid priority", Priority("invalid"), false},
		{"Empty priority", Priority(""), false},
		{"Uppercase HIGH", Priority("HIGH"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.Valid(); got != tt.want {
				t.Errorf("Priority.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriority_String(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     string
	}{
		{"High", PriorityHigh, "high"},
		{"Normal", PriorityNormal, "normal"},
		{"Low", PriorityLow, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.String(); got != tt.want {
				t.Errorf("Priority.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriority_RabbitMQPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		want     uint8
	}{
		{"High priority", PriorityHigh, 3},
		{"Normal priority", PriorityNormal, 2},
		{"Low priority", PriorityLow, 1},
		{"Invalid priority defaults to normal", Priority("invalid"), 2},
		{"Empty priority defaults to normal", Priority(""), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.RabbitMQPriority(); got != tt.want {
				t.Errorf("Priority.RabbitMQPriority() = %v, want %v", got, tt.want)
			}
		})
	}
}
