package bootstrap

import (
	"os"
	"time"
)

type Config struct {
	RunAddress  string
	DatabaseURI string
	JWTSecret   string
	JWTExpiry   time.Duration
}

func ConfigLoad() *Config {
	cfg := &Config{}

	cfg.RunAddress = getEnv("RUN_ADDRESS", "localhost:8081")
	cfg.DatabaseURI = getEnv("DATABASE_URI", "")
	cfg.JWTSecret = getEnv("JWT_SECRET", "your-secret-key-change-in-production")

	expiryStr := getEnv("JWT_EXPIRY", "30m")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 30 * time.Minute
	}
	cfg.JWTExpiry = expiry

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
