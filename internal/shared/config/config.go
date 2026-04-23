package config

import (
	"os"
	"strconv"
)

// Config holds all static configuration loaded from environment variables.
// Business-logic values (pricing, algorithm weights, limits) are in DynamicConfig (database-backed).
// This struct holds infrastructure config only.
type Config struct {
	// Server
	Port string
	Env  string // development, staging, production

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Telegram
	TelegramBotToken  string
	TelegramWebAppURL string // URL where the Mini App is hosted
	UserbotURL        string // Python userbot service URL

	// Cloudflare R2 (S3-compatible)
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey  string
	R2BucketName      string
	R2PublicURL       string // CDN URL for serving media

	// AWS Rekognition (face verification)
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	// Paystack (event payments)
	PaystackSecretKey string
	PaystackPublicKey string

	// JWT
	JWTSecret string

	// Admin
	AdminTelegramIDs []int64 // Telegram user IDs allowed to access admin
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENV", "development"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/rooted?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),

		TelegramBotToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramWebAppURL: getEnv("TELEGRAM_WEBAPP_URL", ""),
		UserbotURL:        getEnv("USERBOT_URL", ""),

		R2AccountID:      getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:    getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:     getEnv("R2_BUCKET_NAME", "rooted-media"),
		R2PublicURL:      getEnv("R2_PUBLIC_URL", ""),

		AWSRegion:          getEnv("AWS_REGION", "eu-west-1"),
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),

		PaystackSecretKey: getEnv("PAYSTACK_SECRET_KEY", ""),
		PaystackPublicKey: getEnv("PAYSTACK_PUBLIC_KEY", ""),

		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),

		AdminTelegramIDs: parseIntList(getEnv("ADMIN_TELEGRAM_IDS", "")),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseIntList(s string) []int64 {
	if s == "" {
		return nil
	}
	var result []int64
	for _, part := range splitComma(s) {
		if v, err := strconv.ParseInt(part, 10, 64); err == nil {
			result = append(result, v)
		}
	}
	return result
}

func splitComma(s string) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
		} else if c != ' ' {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
