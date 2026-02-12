package notification

// Content and batch limits (assessment: character limits, required fields).
const (
	MaxContentLengthSMS   = 1600
	MaxContentLengthEmail = 100_000
	MaxContentLengthPush  = 4_096
	MaxRecipientLength    = 512
	MaxBatchSize          = 1000
)

// MaxContentLength returns the max content length for the channel.
func MaxContentLength(c Channel) int {
	switch c {
	case ChannelSMS:
		return MaxContentLengthSMS
	case ChannelEmail:
		return MaxContentLengthEmail
	case ChannelPush:
		return MaxContentLengthPush
	default:
		return MaxContentLengthEmail
	}
}
