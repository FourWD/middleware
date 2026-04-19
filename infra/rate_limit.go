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

// buildRateLimiter constructs the RateLimiter used by AppRuntimeDeps.
// When Redis is available it uses RedisRateLimitStore (distributed, atomic via Lua);
// otherwise falls back to InMemoryRateLimitStore (per-instance counters).
func buildRateLimiter(cfg CommonConfig, redis *RedisClient, hooks *[]func(context.Context) error) *RateLimiter {
	if !cfg.RateLimitEnabled {
		return NewRateLimiter(cfg, nil)
	}

	var store RateLimitStore
	if redis != nil {
		store = NewRedisRateLimitStore(redis)
	} else {
		inMem := NewInMemoryRateLimitStore()
		*hooks = append(*hooks, func(context.Context) error {
			inMem.Close()
			return nil
		})
		store = inMem
	}

	return NewRateLimiter(cfg, store)
}

// RateLimiter exposes tiered rate-limit middleware factories for use with
// route groups. Construct via NewRateLimiter; NewApp wires one into AppDeps.
//
// Tiers:
//   - Strict() : low cap per minute — auth endpoints (login, register, reset)
//   - Default(): high cap per second — regular API traffic
//   - skip     : do not attach any middleware (e.g. /metrics, /health)
type RateLimiter struct {
	store            RateLimitStore
	enabled          bool
	strictPerMinute  int
	defaultPerSecond int
	keyPrefix        string
}

// NewRateLimiter builds a RateLimiter from CommonConfig + a store.
// If cfg.RateLimitEnabled is false or store is nil, middleware becomes a no-op.
func NewRateLimiter(cfg CommonConfig, store RateLimitStore) *RateLimiter {
	return &RateLimiter{
		store:            store,
		enabled:          cfg.RateLimitEnabled && store != nil,
		strictPerMinute:  cfg.RateLimitStrictPerMinute,
		defaultPerSecond: cfg.RateLimitDefaultPerSecond,
		keyPrefix:        cfg.AppID,
	}
}

// Strict returns a middleware with "RateLimitStrictPerMinute" cap per minute.
// Intended for auth-sensitive endpoints.
func (r *RateLimiter) Strict() fiber.Handler {
	return r.build("strict", r.strictPerMinute, time.Minute)
}

// Default returns a middleware with "RateLimitDefaultPerSecond" cap per second.
// Intended for regular API routes.
func (r *RateLimiter) Default() fiber.Handler {
	return r.build("default", r.defaultPerSecond, time.Second)
}

func (r *RateLimiter) build(tier string, max int, window time.Duration) fiber.Handler {
	if r == nil || !r.enabled || max <= 0 {
		return func(c fiber.Ctx) error { return c.Next() }
	}

	return func(c fiber.Ctx) error {
		client := c.IP()
		if auth := c.Get("Authorization"); auth != "" {
			client = auth
		}
		key := fmt.Sprintf("%s:%s:%s:%s", r.keyPrefix, tier, client, c.Path())

		count, err := r.store.IncrWithExpiry(c.Context(), key, window)
		if err != nil {
			// fail-open on store error — prefer availability over strict enforcement
			return c.Next()
		}
		if count > int64(max) {
			return WriteErrorEnvelope(c, fiber.StatusTooManyRequests, "rate_limited", "too many requests")
		}
		return c.Next()
	}
}

