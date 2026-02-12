package postgres

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DSN string
}

type DB struct {
	*gorm.DB
}

func New(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("postgres: DSN required")
	}
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	if err := db.WithContext(ctx).AutoMigrate(
		&BatchModel{},
		&NotificationModel{},
		&DeliveryAttemptModel{},
	); err != nil {
		return nil, fmt.Errorf("postgres: migrate: %w", err)
	}
	return &DB{DB: db}, nil
}
