package config

import (
	"log"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Production should use real environment variables (Docker/K8s/CI).
func Load() *Config {
	env := getEnv("APP_ENV", "local")

	envFile := filepath.Join(".", ".env."+env)
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("config: .env file not found (%s), using system environment", envFile)
	} else {
		log.Printf("config: loaded %s", envFile)
	}

	cfg := &Config{
		Env: env,
		App: AppConfig{
			Port: mustGetEnv("APP_PORT"),
		},
		DB: DBConfig{
			DSN: mustGetEnv("DB_DSN"),
		},
		Redis: RedisConfig{
			Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:            getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			ManagementURL:  getEnv("RABBITMQ_MANAGEMENT_URL", "http://localhost:15672"),
			ManagementUser: getEnv("RABBITMQ_MANAGEMENT_USER", "guest"),
			ManagementPass: getEnv("RABBITMQ_MANAGEMENT_PASS", "guest"),
		},
		Webhook: WebhookConfig{
			URL: getEnv("WEBHOOK_URL", "https://webhook.site/unique-id"),
		},
	}

	log.Printf("config: environment=%s port=%s", cfg.Env, cfg.App.Port)

	return cfg
}
