package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL        string
	Port               string
	RateLimitRPS       int
	RateLimitBurst     int
	CORSAllowedOrigins string
}

func Load() (*Config, error) {
	rps, _ := strconv.Atoi(getenv("RATE_LIMIT_RPS", "5"))
	burst, _ := strconv.Atoi(getenv("RATE_LIMIT_BURST", "10"))
	return &Config{
		DatabaseURL:        getenv("DATABASE_URL", "postgres://postgres:password@localhost:5432/testdb?sslmode=disable"),
		Port:               getenv("PORT", "8080"),
		RateLimitRPS:       rps,
		RateLimitBurst:     burst,
		CORSAllowedOrigins: getenv("CORS_ALLOWED_ORIGINS", "*"),
	}, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
