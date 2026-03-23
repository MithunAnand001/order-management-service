package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	AppPort          string
	JWTSecret        string
	AccessTokenExp   int // in minutes
	RefreshTokenExp  int // in hours
	RabbitMQURL      string
	RabbitExchange   string
	RabbitQueue      string
	RabbitBinding    string
	RabbitDLX        string
	EnableMigration  bool
	MaxRetryAttempts int
	RetryBaseDelay   int // in milliseconds
	SMTPHost         string
	SMTPPort         int
	SMTPUser         string
	SMTPPass         string
	FromEmail        string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	// Parse Access Token Exp (e.g., "30m" -> 30)
	accessTokenStr := getEnv("JWT_EXPIRES_IN", "30m")
	accessTokenExp, _ := strconv.Atoi(strings.TrimSuffix(accessTokenStr, "m"))
	if accessTokenExp == 0 {
		accessTokenExp = 30
	}

	// Parse Refresh Token Exp (e.g., "7d" -> 168 hours)
	refreshTokenStr := getEnv("JWT_REFRESH_EXPIRES_IN", "7d")
	days, _ := strconv.Atoi(strings.TrimSuffix(refreshTokenStr, "d"))
	refreshTokenExp := days * 24
	if refreshTokenExp == 0 {
		refreshTokenExp = 48 // fallback 2 days
	}

	maxRetry, _ := strconv.Atoi(getEnv("MAX_RETRY_ATTEMPTS", "3"))
	retryDelay, _ := strconv.Atoi(getEnv("RETRY_BASE_DELAY_MS", "5000"))
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "2525"))

	return &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USERNAME", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "admin123"),
		DBName:           getEnv("DB_DATABASE", "peerisland_orders"),
		AppPort:          getEnv("PORT", "3001"),
		JWTSecret:        getEnv("JWT_SECRET", "8f3c9a1b6d4e7f2a9c5e3b1d8f0a7c6e9d4b2c1f8a6e3d5c7b9a1e2f4c6d8b0"),
		AccessTokenExp:   accessTokenExp,
		RefreshTokenExp:  refreshTokenExp,
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://peerisland_user:peerisland_pass@localhost:5672"),
		RabbitExchange:   getEnv("RABBITMQ_EXCHANGE", "webhook.exchange"),
		RabbitQueue:      getEnv("RABBITMQ_QUEUE_WEBHOOK_EVENTS", "webhook.events"),
		RabbitBinding:    "order.created",
		RabbitDLX:        getEnv("RABBITMQ_QUEUE_RETRY", "webhook.events.retry"),
		EnableMigration:  getEnv("ENABLE_MIGRATION", "false") == "true",
		MaxRetryAttempts: maxRetry,
		RetryBaseDelay:   retryDelay,
		SMTPHost:         getEnv("SMTP_HOST", "smtp.mailtrap.io"),
		SMTPPort:         smtpPort,
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPass:         getEnv("SMTP_PASS", ""),
		FromEmail:        getEnv("FROM_EMAIL", "noreply@peerisland.com"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
