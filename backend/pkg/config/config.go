package config

import "os"

type Config struct {
	Port        string
	JWTSecret   string
	DatabaseURL string
	RedisURL    string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "kingdom-arena-secret-dev"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/kingdom_arena?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
