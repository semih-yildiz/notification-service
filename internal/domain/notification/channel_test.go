package notification

import "testing"

func TestChannel_Valid(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		want    bool
	}{
		{"SMS channel", ChannelSMS, true},
		{"Email channel", ChannelEmail, true},
		{"Push channel", ChannelPush, true},
		{"Invalid channel", Channel("invalid"), false},
		{"Empty channel", Channel(""), false},
		{"Uppercase SMS", Channel("SMS"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.channel.Valid(); got != tt.want {
				t.Errorf("Channel.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannel_String(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		want    string
	}{
		{"SMS", ChannelSMS, "sms"},
		{"Email", ChannelEmail, "email"},
		{"Push", ChannelPush, "push"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.channel.String(); got != tt.want {
				t.Errorf("Channel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
