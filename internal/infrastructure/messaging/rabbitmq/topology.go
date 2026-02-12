package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName    = "notifications"
	DLXExchangeName = "notifications.dlx"

	QueueSMS   = "notifications.sms"
	QueueEmail = "notifications.email"
	QueuePush  = "notifications.push"

	QueueSMSDLQ   = "notifications.sms.dlq"
	QueueEmailDLQ = "notifications.email.dlq"
	QueuePushDLQ  = "notifications.push.dlq"
)

// DeclareTopology declares exchange and channel-based queues with priority and DLQ.
func DeclareTopology(ch *amqp.Channel) error {
	// Main exchange
	if err := ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	// Dead Letter Exchange (DLX)
	if err := ch.ExchangeDeclare(DLXExchangeName, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	// Main queues with priority and DLX
	mainQueues := []struct {
		name       string
		routingKey string
	}{
		{QueueSMS, RoutingKeySMS},
		{QueueEmail, RoutingKeyEmail},
		{QueuePush, RoutingKeyPush},
	}

	for _, q := range mainQueues {
		args := amqp.Table{
			"x-max-priority":         int32(4),
			"x-dead-letter-exchange": DLXExchangeName,
		}
		if _, err := ch.QueueDeclare(q.name, true, false, false, false, args); err != nil {
			return err
		}
		if err := ch.QueueBind(q.name, q.routingKey, ExchangeName, false, nil); err != nil {
			return err
		}
	}

	dlqQueues := []struct {
		name       string
		routingKey string
	}{
		{QueueSMSDLQ, RoutingKeySMS},
		{QueueEmailDLQ, RoutingKeyEmail},
		{QueuePushDLQ, RoutingKeyPush},
	}

	for _, q := range dlqQueues {
		if _, err := ch.QueueDeclare(q.name, true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(q.name, q.routingKey, DLXExchangeName, false, nil); err != nil {
			return err
		}
	}

	return nil
}
