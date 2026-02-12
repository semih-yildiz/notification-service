package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/semih-yildiz/notification-service/internal/application/notification/command/process"
	"github.com/semih-yildiz/notification-service/internal/application/notification/port"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/cache/redis"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/config"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/persistence/postgres"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/provider/webhook"
	"github.com/semih-yildiz/notification-service/internal/shared/logger"
)

func main() {
	// Load configuration from environment (and optional .env file)
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Infrastructure: DB
	db, err := postgres.New(ctx, postgres.Config{DSN: cfg.DB.DSN})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	sqlDB, _ := db.DB.DB()
	defer sqlDB.Close()

	// Redis
	rdb, err := redis.NewClient(redis.Config{Addr: cfg.Redis.Addr})
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	// RabbitMQ Consumer
	consumer, err := rabbitmq.NewConsumer(rabbitmq.Config{URL: cfg.RabbitMQ.URL})
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer consumer.Close()

	// Repositories and services
	notifRepo := postgres.NewNotificationRepository(db.DB)
	attemptRepo := postgres.NewDeliveryAttemptRepository(db.DB)
	rateLimiter := redis.NewRateLimiter(rdb)
	deliveryClient := webhook.NewClient(cfg.Webhook.URL)
	appLogger := logger.New()

	processUseCase := process.NewUseCase(notifRepo, attemptRepo, rateLimiter, deliveryClient, appLogger)

	processFn := func(ctx context.Context, evt *port.NotificationEvent) error {
		return processUseCase.Execute(ctx, &process.Command{NotificationID: evt.NotificationID})
	}

	log.Printf("worker consuming (env=%s)", cfg.Env)
	_ = consumer.Run(ctx, processFn)
	log.Println("worker shutdown")
}
