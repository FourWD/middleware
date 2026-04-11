package infra

import (
	goredis "github.com/redis/go-redis/v9"
)

// RedisClient is a type alias for the go-redis Client.
type RedisClient = goredis.Client

// RedisConfig holds Redis connection configuration.
// Use LoadRedisConfig() to populate from environment variables.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// LoadRedisConfig reads Redis configuration from environment variables.
func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     GetEnv("REDIS_ADDR", "127.0.0.1:6379"),
		Password: GetEnv("REDIS_PASSWORD", ""),
		DB:       GetEnvInt("REDIS_DB", 0),
	}
}

// NewRedisClient creates a new Redis client from configuration.
func NewRedisClient(cfg RedisConfig) *RedisClient {
	return goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
