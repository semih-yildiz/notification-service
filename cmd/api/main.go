package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/semih-yildiz/notification-service/internal/application/notification/command/cancel"
	"github.com/semih-yildiz/notification-service/internal/application/notification/command/create"
	"github.com/semih-yildiz/notification-service/internal/application/notification/query/get"
	"github.com/semih-yildiz/notification-service/internal/application/notification/query/list"
	httpserver "github.com/semih-yildiz/notification-service/internal/http"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/cache/redis"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/config"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/persistence/postgres"
	"github.com/semih-yildiz/notification-service/internal/shared/logger"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize database connection
	db, err := postgres.New(ctx, postgres.Config{DSN: cfg.DB.DSN})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatalf("db sql: %v", err)
	}
	defer sqlDB.Close()

	// Redis
	rdb, err := redis.NewClient(redis.Config{Addr: cfg.Redis.Addr})
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	// RabbitMQ Publisher
	pub, err := rabbitmq.NewPublisher(rabbitmq.Config{URL: cfg.RabbitMQ.URL})
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer pub.Close()

	// RabbitMQ Management API client (for metrics)
	mqManagement := rabbitmq.NewManagementClient(
		cfg.RabbitMQ.ManagementURL,
		cfg.RabbitMQ.ManagementUser,
		cfg.RabbitMQ.ManagementPass,
	)

	// Repositories
	notifRepo := postgres.NewNotificationRepository(db.DB)
	batchRepo := postgres.NewBatchRepository(db.DB)
	metricsRepo := postgres.NewMetricsRepository(db.DB)
	idemStore := redis.NewIdempotencyStore(rdb)
	appLogger := logger.New()

	// Application layer: usecase
	createUsecase := create.NewUseCase(notifRepo, batchRepo, pub, idemStore, appLogger)
	cancelUsecase := cancel.NewUseCase(notifRepo)
	getUsecase := get.NewUseCase(notifRepo, batchRepo)
	listUsecase := list.NewUseCase(notifRepo)

	// HTTP layer: handle
	notificationHandler := httpserver.NewNotificationHandler(createUsecase, cancelUsecase, getUsecase, listUsecase)
	healthHandler := httpserver.NewHealthHandler(sqlDB, rdb, metricsRepo, mqManagement)

	// Initialize Echo server
	e := httpserver.NewEcho(notificationHandler, healthHandler, "")
	e.Server.Addr = ":" + cfg.App.Port

	// Start server
	go func() {
		log.Printf("api listening on :%s (env=%s)", cfg.App.Port, cfg.Env)
		if err := e.Start(e.Server.Addr); err != nil && err != http.ErrServerClosed {
			log.Printf("serve: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	if err := e.Shutdown(context.Background()); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("api shutdown")
}
