package notification

import "testing"

func TestStatus_Valid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"Pending status", StatusPending, true},
		{"Queued status", StatusQueued, true},
		{"Sent status", StatusSent, true},
		{"Failed status", StatusFailed, true},
		{"Cancelled status", StatusCancelled, true},
		{"Invalid status", Status("invalid"), false},
		{"Empty status", Status(""), false},
		{"Uppercase PENDING", Status("PENDING"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.want {
				t.Errorf("Status.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{"Pending", StatusPending, "pending"},
		{"Queued", StatusQueued, "queued"},
		{"Sent", StatusSent, "sent"},
		{"Failed", StatusFailed, "failed"},
		{"Cancelled", StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_Terminal(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"Pending is not terminal", StatusPending, false},
		{"Queued is not terminal", StatusQueued, false},
		{"Sent is terminal", StatusSent, true},
		{"Failed is terminal", StatusFailed, true},
		{"Cancelled is terminal", StatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Terminal(); got != tt.want {
				t.Errorf("Status.Terminal() = %v, want %v", got, tt.want)
			}
		})
	}
}
