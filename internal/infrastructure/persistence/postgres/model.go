package postgres

import (
	"time"

	"gorm.io/gorm"
)

type BatchModel struct {
	ID             string         `gorm:"type:text;primaryKey"`
	IdempotencyKey *string        `gorm:"type:text;uniqueIndex"`
	CreatedAt      time.Time      `gorm:"not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (BatchModel) TableName() string { return "batches" }

type NotificationModel struct {
	ID             string         `gorm:"type:text;primaryKey"`
	BatchID        *string        `gorm:"type:text;index"`
	Recipient      string         `gorm:"type:text;not null"`
	Channel        string         `gorm:"type:text;not null"`
	Content        string         `gorm:"type:text;not null"`
	Priority       string         `gorm:"type:text;not null"`
	Status         string         `gorm:"type:text;not null"`
	IdempotencyKey *string        `gorm:"type:text;uniqueIndex"`
	CreatedAt      time.Time      `gorm:"not null"`
	UpdatedAt      time.Time      `gorm:"not null"`
	SentAt         *time.Time     `gorm:"type:timestamptz"`
	FailureReason  *string        `gorm:"type:text"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (NotificationModel) TableName() string { return "notifications" }

type DeliveryAttemptModel struct {
	ID             string         `gorm:"type:text;primaryKey"`
	NotificationID string         `gorm:"type:text;not null;index"`
	AttemptNumber  int            `gorm:"not null"`
	Success        bool           `gorm:"not null"`
	StatusCode     int            `gorm:"not null"`
	ResponseBody   string         `gorm:"type:text"`
	ErrorMessage   *string        `gorm:"type:text"`
	CreatedAt      time.Time      `gorm:"not null"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (DeliveryAttemptModel) TableName() string { return "delivery_attempts" }
