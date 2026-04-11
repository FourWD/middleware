package infra

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	goredis "github.com/redis/go-redis/v9"
)

// rateLimitScript atomically increments a key and sets expiry on first increment.
// This avoids the race condition where INCR succeeds but a separate EXPIRE call fails,
// which would leave the key without a TTL (persisting forever).
var rateLimitScript = goredis.NewScript(`
	local count = redis.call("INCR", KEYS[1])
	if count == 1 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return count
`)

// RateLimitStore is the interface for rate limit backends (e.g. Redis).
// IncrWithExpiry atomically increments a key's counter and sets expiry on first increment.
// Returns the current count after increment.
type RateLimitStore interface {
	IncrWithExpiry(ctx context.Context, key string, window time.Duration) (count int64, err error)
}

// RedisRateLimitStore implements RateLimitStore using Redis.
type RedisRateLimitStore struct {
	client *RedisClient
}

// NewRedisRateLimitStore creates a Redis-backed rate limit store.
func NewRedisRateLimitStore(client *RedisClient) *RedisRateLimitStore {
	return &RedisRateLimitStore{client: client}
}

func (s *RedisRateLimitStore) IncrWithExpiry(ctx context.Context, key string, window time.Duration) (int64, error) {
	result, err := rateLimitScript.Run(ctx, s.client, []string{key}, int(window.Seconds())).Int64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// InMemoryRateLimitStore implements RateLimitStore using process memory.
type InMemoryRateLimitStore struct {
	mu       sync.Mutex
	counters map[string]inMemoryCounter
	stop     chan struct{}
}

type inMemoryCounter struct {
	count     int64
	expiresAt time.Time
}

func NewInMemoryRateLimitStore() *InMemoryRateLimitStore {
	s := &InMemoryRateLimitStore{
		counters: make(map[string]inMemoryCounter),
		stop:     make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

func (s *InMemoryRateLimitStore) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for key, counter := range s.counters {
				if !counter.expiresAt.After(now) {
					delete(s.counters, key)
				}
			}
			s.mu.Unlock()
		}
	}
}

// Close stops the background cleanup goroutine.
func (s *InMemoryRateLimitStore) Close() {
	close(s.stop)
}

func (s *InMemoryRateLimitStore) IncrWithExpiry(_ context.Context, key string, window time.Duration) (int64, error) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	counter, ok := s.counters[key]
	if !ok || !counter.expiresAt.After(now) {
		counter = inMemoryCounter{
			count:     1,
			expiresAt: now.Add(window),
		}
		s.counters[key] = counter
		return counter.count, nil
	}

	counter.count++
	s.counters[key] = counter
	return counter.count, nil
}

func registerRateLimit(app *fiber.App, cfg StackConfig) {
	if !cfg.RateLimitEnabled || cfg.RateLimitStore == nil {
		return
	}

	if cfg.RateLimitWindowSeconds <= 0 {
		cfg.RateLimitWindowSeconds = 60
	}

	window := time.Duration(cfg.RateLimitWindowSeconds) * time.Second

	app.Use(func(c fiber.Ctx) error {
		if cfg.RateLimitExemptHealth && c.Path() == "/healthz" {
			return c.Next()
		}

		key := fmt.Sprintf("%s:%s:%s", cfg.RateLimitKeyPrefix, c.IP(), c.Path())
		count, err := cfg.RateLimitStore.IncrWithExpiry(c.Context(), key, window)
		if err != nil {
			return WriteErrorEnvelope(c, fiber.StatusInternalServerError, "rate_limit_store_error", "rate limit store unavailable")
		}

		if count > int64(cfg.RateLimitMaxRequests) {
			return WriteErrorEnvelope(c, fiber.StatusTooManyRequests, "rate_limited", "too many requests")
		}

		return c.Next()
	})
}
