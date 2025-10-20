package config

import (
	"os"
	"strconv"
)

// Config holds all configuration loaded from environment variables
type Config struct {
	DatabaseURL        string  `env:"DATABASE_URL"`
	Port               string  `env:"PORT"`
	RateLimitRPS       int     `env:"RATE_LIMIT_RPS"`
	RateLimitBurst     int     `env:"RATE_LIMIT_BURST"`
	CORSAllowedOrigins string  `env:"CORS_ALLOWED_ORIGINS"`

	SMTPHost       string  `env:"SMTP_HOST"`
	SMTPPort       int     `env:"SMTP_PORT"`
	SMTPUser       string  `env:"SMTP_USER"`
	SMTPPass       string  `env:"SMTP_PASS"`
	EmailFrom      string  `env:"EMAIL_FROM"`

	JWTSecret      string  `env:"JWT_SECRET"`
	JWTExpiryHours int     `env:"JWT_EXPIRY_HOURS"`
}

// Load loads configuration from environment and provides defaults if missing
func Load() (*Config, error) {
	rps, _ := strconv.Atoi(getenv("RATE_LIMIT_RPS", "5"))
	burst, _ := strconv.Atoi(getenv("RATE_LIMIT_BURST", "10"))
	smtpPort, _ := strconv.Atoi(getenv("SMTP_PORT", "587"))
	jwtHours, _ := strconv.Atoi(getenv("JWT_EXPIRY_HOURS", "24"))

	cfg := &Config{
		DatabaseURL:        getenv("DATABASE_URL", "postgres://user:password@postgres:5432/testdb?sslmode=disable"),
		Port:               getenv("PORT", "8081"),
		RateLimitRPS:       rps,
		RateLimitBurst:     burst,
		CORSAllowedOrigins: getenv("CORS_ALLOWED_ORIGINS", "*"),

		SMTPHost:       getenv("SMTP_HOST", "smtp.example.com"),
		SMTPPort:       smtpPort,
		SMTPUser:       getenv("SMTP_USER", "user@example.com"),
		SMTPPass:       getenv("SMTP_PASS", "password"),
		EmailFrom:      getenv("EMAIL_FROM", "noreply@example.com"),

		JWTSecret:      getenv("JWT_SECRET", "supersecretjwtkey"),
		JWTExpiryHours: jwtHours,
	}

	return cfg, nil
}

// getenv reads an environment variable or returns default if not set
func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
