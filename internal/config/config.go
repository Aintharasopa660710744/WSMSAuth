package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Mode string // "debug" | "release"
}

type DatabaseConfig struct {
	DSN string
}

type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
}

func Load() *Config {
	accessExpiry, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_MINUTES", "15"))
	refreshExpiry, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "7"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8081"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "postgres://postgres:password@localhost:5432/authdb?sslmode=disable"),
		},
		JWT: JWTConfig{
			AccessSecret:       getEnv("JWT_ACCESS_SECRET", "your-access-secret-change-in-production"),
			RefreshSecret:      getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-change-in-production"),
			AccessTokenExpiry:  time.Duration(accessExpiry) * time.Minute,
			RefreshTokenExpiry: time.Duration(refreshExpiry) * 24 * time.Hour,
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
