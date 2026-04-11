package infra

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadEnvFiles loads environment variables from .env files in priority order:
// .env → .env.local when APP_ENV is empty, otherwise .env.{APP_ENV}
// Existing OS env vars are never overwritten.
func LoadEnvFiles() error {
	appEnv := ""

	base, err := readEnvFile(".env")
	if err != nil {
		return err
	}
	if v, ok := base["APP_ENV"]; ok && v != "" {
		appEnv = v
	}
	if v := os.Getenv("APP_ENV"); v != "" {
		appEnv = v
	}

	merged := map[string]string{}
	mergeEnvMap(merged, base)

	envPath := ".env.local"
	if appEnv != "" {
		envPath = ".env." + appEnv
	}

	envByTarget, err := readEnvFile(envPath)
	if err != nil {
		return err
	}
	mergeEnvMap(merged, envByTarget)

	for key, value := range merged {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s: %w", key, err)
		}
	}

	return nil
}

func readEnvFile(path string) (map[string]string, error) {
	values, err := godotenv.Read(path)
	if err == nil {
		return values, nil
	}

	var pErr *os.PathError
	if errors.As(err, &pErr) && os.IsNotExist(pErr.Err) {
		return map[string]string{}, nil
	}
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}

	return nil, fmt.Errorf("read %s: %w", path, err)
}

func mergeEnvMap(dst, src map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
}

// GetEnv returns the value of the environment variable or the fallback.
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetEnvInt returns the integer value of the environment variable or the fallback.
func GetEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// GetEnvBool returns the boolean value of the environment variable or the fallback.
func GetEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// GetEnvFloat returns the float64 value of the environment variable or the fallback.
func GetEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}
