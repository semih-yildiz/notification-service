package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.RWMutex
}

type Config struct {
	URL string
}

func NewPublisher(cfg Config) (*Publisher, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	if err := DeclareTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq topology: %w", err)
	}
	return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) Publish(ctx context.Context, evt *port.NotificationEvent) error {
	return p.publish(ctx, evt)
}

func (p *Publisher) PublishBatch(ctx context.Context, events []*port.NotificationEvent) error {
	for _, evt := range events {
		if err := p.publish(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publisher) publish(ctx context.Context, evt *port.NotificationEvent) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.ch == nil {
		return fmt.Errorf("rabbitmq: channel closed")
	}

	routingKey := evt.Channel.String()
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	priority := evt.Priority.RabbitMQPriority()
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
		Headers:      amqp.Table{"priority": int(priority)},
		Priority:     priority,
	}

	return p.ch.PublishWithContext(ctx, ExchangeName, routingKey, false, false, msg)
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
	if p.conn != nil {
		err := p.conn.Close()
		p.conn = nil
		return err
	}
	return nil
}
