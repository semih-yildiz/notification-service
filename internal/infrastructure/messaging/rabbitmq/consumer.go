package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
)

type Consumer struct {
	conn    *amqp.Connection
	queues  []string
	mu      sync.Mutex
	connURL string // Store URL for reconnection
}

func NewConsumer(cfg Config) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := DeclareTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	_ = ch.Close()
	return &Consumer{
		conn:    conn,
		queues:  QueueNames(),
		connURL: cfg.URL,
	}, nil
}

// ProcessFunc is called for each message; return nil to ack.
type ProcessFunc func(ctx context.Context, evt *port.NotificationEvent) error

func (c *Consumer) Run(ctx context.Context, process ProcessFunc) error {
	var wg sync.WaitGroup
	for _, q := range c.queues {
		q := q
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.consumeQueueWithRestart(ctx, q, process)
		}()
	}
	wg.Wait()
	return nil
}

// consumeQueueWithRestart wraps consumeQueue with panic recovery and auto-restart.
func (c *Consumer) consumeQueueWithRestart(ctx context.Context, queue string, process ProcessFunc) {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Printf("rabbitmq consumer %s: shutting down (context cancelled)", queue)
			return
		default:
		}

		// Run with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("rabbitmq consumer panic recovered in queue %s: %v", queue, r)
				}
			}()
			c.consumeQueue(ctx, queue, process)
		}()

		// If we reach here, consumption ended (panic, connection loss, or normal exit)
		select {
		case <-ctx.Done():
			log.Printf("rabbitmq consumer %s: stopped", queue)
			return
		default:
			log.Printf("rabbitmq consumer %s: connection lost, attempting reconnect after %v", queue, backoff)
			time.Sleep(backoff)

			// Try to reconnect
			if err := c.reconnect(); err != nil {
				log.Printf("rabbitmq consumer %s: reconnect failed: %v", queue, err)

				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			} else {
				// Successful reconnection, reset backoff
				backoff = time.Second
				log.Printf("rabbitmq consumer %s: reconnected successfully, resuming consumption", queue)
			}
		}
	}
}

func (c *Consumer) consumeQueue(ctx context.Context, queue string, process ProcessFunc) {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Printf("rabbitmq consumer channel %s: %v", queue, err)
		return
	}
	defer ch.Close()

	if err := ch.Qos(10, 0, false); err != nil {
		log.Printf("rabbitmq consumer qos %s: %v", queue, err)
		return
	}

	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("rabbitmq consume %s: %v", queue, err)
		return
	}

	log.Printf("rabbitmq consumer %s: started", queue)

	for {
		select {
		case <-ctx.Done():
			log.Printf("rabbitmq consumer %s: context done", queue)
			return
		case d, ok := <-deliveries:
			if !ok {
				log.Printf("rabbitmq consumer %s: delivery channel closed, reconnecting...", queue)
				return
			}
			var evt port.NotificationEvent
			if json.Unmarshal(d.Body, &evt) != nil {
				// Invalid JSON: send to DLQ immediately
				_ = d.Nack(false, false)
				continue
			}

			// Check death count to prevent infinite retries
			deathCount := getDeathCount(d.Headers)
			const maxRetries = 3

			msgCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			err := process(msgCtx, &evt)
			cancel()

			if err != nil {
				log.Printf("rabbitmq process %s: %v (attempt %d/%d)", evt.NotificationID, err, deathCount+1, maxRetries)

				if deathCount >= maxRetries {
					log.Printf("rabbitmq max retries reached for %s, sending to DLQ", evt.NotificationID)
					_ = d.Nack(false, false)
				} else {
					_ = d.Nack(false, true)
				}
				continue
			}
			_ = d.Ack(false)
		}
	}
}

func getDeathCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	xDeath, ok := headers["x-death"].([]interface{})
	if !ok || len(xDeath) == 0 {
		return 0
	}
	if death, ok := xDeath[0].(amqp.Table); ok {
		if count, ok := death["count"].(int64); ok {
			return int(count)
		}
	}
	return 0
}

// reconnect attempts to reconnect to RabbitMQ.
func (c *Consumer) reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		_ = c.conn.Close()
	}

	conn, err := amqp.Dial(c.connURL)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	if err := DeclareTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return err
	}
	_ = ch.Close()

	c.conn = conn
	log.Printf("rabbitmq consumer: successfully reconnected")
	return nil
}

func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}
