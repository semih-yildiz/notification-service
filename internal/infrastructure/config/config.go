package config

type Config struct {
	Env      string
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	Webhook  WebhookConfig
}

type AppConfig struct {
	Port string
}

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr string
}

type RabbitMQConfig struct {
	URL            string
	ManagementURL  string
	ManagementUser string
	ManagementPass string
}

type WebhookConfig struct {
	URL string
}
