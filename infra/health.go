package infra

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type HealthCheckFunc func(ctx context.Context) error

type HealthComponentCheck struct {
	Name     string
	Required bool
	Check    HealthCheckFunc
}

type HealthComponentStatus struct {
	Status     string `json:"status"`
	Required   bool   `json:"required"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type HealthReport struct {
	Status     string                           `json:"status"`
	CheckedAt  time.Time                        `json:"checked_at"`
	Components map[string]HealthComponentStatus `json:"components"`
}

func (r HealthReport) Healthy() bool {
	return r.Status == "ok"
}

type HealthCheckOptions struct {
	Databases         Databases
	DatabaseRequired  bool
	Redis             *RedisClient
	RedisRequired     bool
	Mongo             *MongoClient
	MongoRequired     bool
	Firebase          *FirebaseClient
	FirestoreRequired bool
	FirestoreCheck    HealthCheckFunc
	Checks            []HealthComponentCheck
}

// HealthCheck verifies the configured dependencies and returns a per-component report.
func HealthCheck(opts HealthCheckOptions) HealthReport {
	ctx, cancel := context.WithTimeout(context.Background(), HealthTimeout)
	defer cancel()

	report := HealthReport{
		Status:     "ok",
		CheckedAt:  time.Now(),
		Components: map[string]HealthComponentStatus{},
	}

	for _, check := range buildHealthChecks(opts) {
		if check.Check == nil {
			report.Components[check.Name] = HealthComponentStatus{
				Status:   "skipped",
				Required: check.Required,
			}
			continue
		}

		start := time.Now()
		err := check.Check(ctx)
		status := HealthComponentStatus{
			Status:     "ok",
			Required:   check.Required,
			DurationMs: time.Since(start).Milliseconds(),
		}
		if err != nil {
			status.Error = err.Error()
			if check.Required {
				status.Status = "down"
				report.Status = "down"
			} else {
				status.Status = "degraded"
				if report.Status == "ok" {
					report.Status = "degraded"
				}
			}
		}
		report.Components[check.Name] = status
	}

	return report
}

func buildHealthChecks(opts HealthCheckOptions) []HealthComponentCheck {
	checks := make([]HealthComponentCheck, 0, 8)

	if opts.Databases.Primary != nil || opts.Databases.Secondary != nil {
		checks = append(checks, HealthComponentCheck{
			Name:     "database",
			Required: opts.DatabaseRequired || opts.Databases.Primary != nil,
			Check: func(ctx context.Context) error {
				if opts.Databases.Primary == nil {
					return fmt.Errorf("primary database is unavailable")
				}
				for _, db := range []*gorm.DB{opts.Databases.Primary, opts.Databases.Secondary} {
					if db == nil {
						continue
					}
					sqlDB, err := db.DB()
					if err != nil {
						return err
					}
					if err := sqlDB.PingContext(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		})
	}

	if opts.Redis != nil {
		checks = append(checks, HealthComponentCheck{
			Name:     "redis",
			Required: opts.RedisRequired || opts.Redis != nil,
			Check: func(ctx context.Context) error {
				return opts.Redis.Ping(ctx).Err()
			},
		})
	}

	if opts.Mongo != nil && opts.Mongo.client != nil {
		checks = append(checks, HealthComponentCheck{
			Name:     "mongo",
			Required: opts.MongoRequired,
			Check: func(ctx context.Context) error {
				return opts.Mongo.client.Ping(ctx, nil)
			},
		})
	}

	if opts.Firebase != nil && opts.Firebase.Firestore != nil {
		checks = append(checks, HealthComponentCheck{
			Name:     "firestore",
			Required: opts.FirestoreRequired,
			Check:    opts.FirestoreCheck,
		})
	}

	checks = append(checks, opts.Checks...)
	return checks
}
