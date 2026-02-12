package notification

import "testing"

func TestMaxContentLength(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		want    int
	}{
		{"SMS channel", ChannelSMS, MaxContentLengthSMS},
		{"Email channel", ChannelEmail, MaxContentLengthEmail},
		{"Push channel", ChannelPush, MaxContentLengthPush},
		{"Invalid channel defaults to Email", Channel("invalid"), MaxContentLengthEmail},
		{"Empty channel defaults to Email", Channel(""), MaxContentLengthEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxContentLength(tt.channel); got != tt.want {
				t.Errorf("MaxContentLength(%v) = %v, want %v", tt.channel, got, tt.want)
			}
		})
	}
}

func TestMaxContentLength_Values(t *testing.T) {
	if MaxContentLengthSMS != 1600 {
		t.Errorf("MaxContentLengthSMS = %d, want 1600", MaxContentLengthSMS)
	}
	if MaxContentLengthEmail != 100_000 {
		t.Errorf("MaxContentLengthEmail = %d, want 100000", MaxContentLengthEmail)
	}
	if MaxContentLengthPush != 4_096 {
		t.Errorf("MaxContentLengthPush = %d, want 4096", MaxContentLengthPush)
	}
	if MaxRecipientLength != 512 {
		t.Errorf("MaxRecipientLength = %d, want 512", MaxRecipientLength)
	}
	if MaxBatchSize != 1000 {
		t.Errorf("MaxBatchSize = %d, want 1000", MaxBatchSize)
	}
}
