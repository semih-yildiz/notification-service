package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpserver "github.com/semih-yildiz/notification-service/internal/http"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/cache/redis"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/config"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/semih-yildiz/notification-service/internal/infrastructure/persistence/postgres"
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

	// RabbitMQ Management API client
	mqManagement := rabbitmq.NewManagementClient(
		cfg.RabbitMQ.ManagementURL,
		cfg.RabbitMQ.ManagementUser,
		cfg.RabbitMQ.ManagementPass,
	)
	_ = mqManagement

	// Initialize Echo server
	e := httpserver.NewEcho()
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
