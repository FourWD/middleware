package infra

import "time"

// Build is the running binary's identity, for a service's own status or
// monitor endpoint. Version and GitSHA come from the environment (CI stamps
// them at deploy); StartedAt is process boot time.
type Build struct {
	Version   string    `json:"version"`
	GitSHA    string    `json:"git_sha"`
	StartedAt time.Time `json:"started_at"`
}

// processStartedAt is captured at package init, which for a normal service is
// close enough to process start — nothing before infra loads is worth timing.
var processStartedAt = time.Now()

// BuildInfo snapshots the identity at call time rather than caching it, so a
// handler marshals fresh values if the environment is reloaded.
func BuildInfo() Build {
	return Build{
		Version:   GetEnv("APP_VERSION", ""),
		GitSHA:    GetEnv("GIT_SHA", ""),
		StartedAt: processStartedAt,
	}
}

// UptimeSecs is whole seconds since process start.
func UptimeSecs() int64 {
	return int64(time.Since(processStartedAt).Seconds())
}
