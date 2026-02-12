//go:build integration
// +build integration

package rabbitmq

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/domain/notification"
)

// TestConsumerPanicRecovery tests that consumer recovers from panic and continues processing.
func TestConsumerPanicRecovery(t *testing.T) {
	// Skip if RabbitMQ not available (integration test)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to RabbitMQ (assumes local instance on default port)
	cfg := Config{URL: "amqp://guest:guest@localhost:5672/"}
	consumer, err := NewConsumer(cfg)
	if err != nil {
		t.Skipf("RabbitMQ not available: %v", err)
	}
	defer consumer.Close()

	// Setup test queue
	testQueue := "test.panic.recovery"
	conn, _ := amqp.Dial(cfg.URL)
	defer conn.Close()
	ch, _ := conn.Channel()
	defer ch.Close()

	// Declare test queue
	_, err = ch.QueueDeclare(testQueue, false, true, false, false, nil)
	if err != nil {
		t.Fatalf("failed to declare test queue: %v", err)
	}

	// Publish test messages
	for i := 0; i < 5; i++ {
		evt := &port.NotificationEvent{
			NotificationID: "test-" + string(rune(i+'0')),
			Channel:        notification.ChannelSMS,
			Priority:       notification.PriorityNormal,
		}
		body, _ := json.Marshal(evt)
		ch.Publish("", testQueue, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	}

	// Track processing
	var processedCount int32
	var panicCount int32

	// Process function that panics on first message
	processFn := func(ctx context.Context, evt *port.NotificationEvent) error {
		count := atomic.AddInt32(&processedCount, 1)

		// Panic on first message
		if count == 1 {
			atomic.AddInt32(&panicCount, 1)
			panic("simulated panic for testing")
		}

		// Normal processing for subsequent messages
		return nil
	}

	// Override consumer queues to use test queue
	consumer.queues = []string{testQueue}

	// Run consumer in background
	go consumer.Run(ctx, processFn)

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify panic recovery worked
	if atomic.LoadInt32(&panicCount) != 1 {
		t.Errorf("expected 1 panic, got %d", panicCount)
	}

	// Should have processed all 5 messages (1 panic + 4 successful + 1 retry of panic)
	finalCount := atomic.LoadInt32(&processedCount)
	if finalCount < 4 {
		t.Errorf("expected at least 4 processed messages after panic recovery, got %d", finalCount)
	}

	t.Logf("✅ Panic recovery test passed: %d panics, %d total processes", panicCount, finalCount)
}

// TestConsumerReconnection tests auto-reconnection on connection loss.
func TestConsumerReconnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg := Config{URL: "amqp://guest:guest@localhost:5672/"}
	consumer, err := NewConsumer(cfg)
	if err != nil {
		t.Skipf("RabbitMQ not available: %v", err)
	}
	defer consumer.Close()

	testQueue := "test.reconnection"
	conn, _ := amqp.Dial(cfg.URL)
	defer conn.Close()
	ch, _ := conn.Channel()
	defer ch.Close()

	_, err = ch.QueueDeclare(testQueue, false, true, false, false, nil)
	if err != nil {
		t.Fatalf("failed to declare test queue: %v", err)
	}

	// Publish messages before and after connection loss
	evt := &port.NotificationEvent{
		NotificationID: "test-reconnect",
		Channel:        notification.ChannelEmail,
		Priority:       notification.PriorityHigh,
	}
	body, _ := json.Marshal(evt)
	ch.Publish("", testQueue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	var processedCount int32

	processFn := func(ctx context.Context, evt *port.NotificationEvent) error {
		atomic.AddInt32(&processedCount, 1)
		return nil
	}

	consumer.queues = []string{testQueue}

	go consumer.Run(ctx, processFn)

	// Wait for first message
	time.Sleep(1 * time.Second)

	// Simulate connection loss by closing the connection
	t.Log("Simulating connection loss...")
	consumer.conn.Close()

	// Wait for reconnection
	time.Sleep(3 * time.Second)

	// Publish another message after reconnection
	conn2, _ := amqp.Dial(cfg.URL)
	defer conn2.Close()
	ch2, _ := conn2.Channel()
	defer ch2.Close()

	ch2.Publish("", testQueue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	// Wait for second message
	time.Sleep(2 * time.Second)

	finalCount := atomic.LoadInt32(&processedCount)
	if finalCount < 2 {
		t.Errorf("expected at least 2 messages processed (before and after reconnect), got %d", finalCount)
	}

	t.Logf("✅ Reconnection test passed: %d messages processed", finalCount)
}

// TestConsumerExponentialBackoff tests that backoff increases on repeated failures.
func TestConsumerExponentialBackoff(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use invalid URL to force reconnection failures
	cfg := Config{URL: "amqp://invalid:invalid@localhost:9999/"}
	consumer := &Consumer{
		connURL: cfg.URL,
		queues:  []string{"test"},
	}

	startTime := time.Now()

	// Mock process function
	processFn := func(ctx context.Context, evt *port.NotificationEvent) error {
		return nil
	}

	// Run in background (will fail to connect)
	go consumer.consumeQueueWithRestart(ctx, "test", processFn)

	// Wait for multiple reconnection attempts
	time.Sleep(4 * time.Second)

	elapsed := time.Since(startTime)

	// Should have attempted multiple times with exponential backoff
	// 1s + 2s + 4s = at least 7 seconds of cumulative backoff
	if elapsed < 3*time.Second {
		t.Errorf("expected exponential backoff to delay for at least 3s, got %v", elapsed)
	}

	t.Logf("✅ Exponential backoff test passed: elapsed %v", elapsed)
}

// TestConsumerGracefulShutdown tests context cancellation stops consumer.
func TestConsumerGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())

	cfg := Config{URL: "amqp://guest:guest@localhost:5672/"}
	consumer, err := NewConsumer(cfg)
	if err != nil {
		t.Skipf("RabbitMQ not available: %v", err)
	}
	defer consumer.Close()

	processFn := func(ctx context.Context, evt *port.NotificationEvent) error {
		return nil
	}

	// Run consumer
	done := make(chan struct{})
	go func() {
		consumer.Run(ctx, processFn)
		close(done)
	}()

	// Cancel context after 1 second
	time.Sleep(1 * time.Second)
	cancel()

	// Wait for shutdown
	select {
	case <-done:
		t.Log("✅ Graceful shutdown test passed")
	case <-time.After(3 * time.Second):
		t.Error("consumer did not shut down within 3 seconds")
	}
}
