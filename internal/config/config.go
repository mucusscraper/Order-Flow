package config

import (
	"os"
)

type Config struct {
	AppEnv      string
	ServerPort  string
	DatabaseURL string
	JWTSecret   string
	RedisURL    string
	KafkaURL    string
}

func Load() *Config {
	return &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/orderflow?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "default-secret-key-change-me"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		KafkaURL:    getEnv("KAFKA_URL", "localhost:9092"),
	}
}

func getEnv(key, fallback string) string {
	val, exists := os.LookupEnv(key)
	if exists {
		return val
	}
	return fallback
}
