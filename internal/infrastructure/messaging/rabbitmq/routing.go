package rabbitmq

// Routing keys: channel name -> queue notifications.sms, .email, .push
const (
	RoutingKeySMS   = "sms"
	RoutingKeyEmail = "email"
	RoutingKeyPush  = "push"
)

func QueueNames() []string {
	return []string{QueueSMS, QueueEmail, QueuePush}
}
