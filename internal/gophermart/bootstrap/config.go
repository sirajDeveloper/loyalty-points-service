package bootstrap

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	RunAddress          string
	DatabaseURI         string
	AccrualSystemAddress string
	JWTSecret           string
	JWTExpiry           time.Duration
}

func ConfigLoad() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", getEnv("RUN_ADDRESS", "localhost:8080"), "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", getEnv("DATABASE_URI", ""), "database connection string")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", getEnv("ACCRUAL_SYSTEM_ADDRESS", ""), "accrual system address")
	flag.StringVar(&cfg.JWTSecret, "j", getEnv("JWT_SECRET", "your-secret-key-change-in-production"), "JWT secret key")
	
	expiryStr := getEnv("JWT_EXPIRY", "30m")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 30 * time.Minute
	}
	cfg.JWTExpiry = expiry

	flag.Parse()

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


