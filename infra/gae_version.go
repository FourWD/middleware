package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	maxWakeUpMinute     = 10
	defaultWakeUpMinute = 5

	gaeVersionCheckOp = "gae_version_check"
)

// wakeUpHTTPClient is shared across polling ticks to avoid allocating a new
// Transport per request.
var wakeUpHTTPClient = &http.Client{Timeout: 10 * time.Second}

// wakeUpInterval reads WAKE_UP_MINUTE.
// Empty/invalid → 5 min. "0" → disabled. Capped at 10.
func wakeUpInterval() time.Duration {
	raw := strings.TrimSpace(os.Getenv("WAKE_UP_MINUTE"))
	if raw == "" {
		return defaultWakeUpMinute * time.Minute
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return defaultWakeUpMinute * time.Minute
	}
	if n <= 0 {
		return 0
	}
	if n > maxWakeUpMinute {
		n = maxWakeUpMinute
	}
	return time.Duration(n) * time.Minute
}

type wakeUpResponse struct {
	Status int `json:"status"`
	Data   struct {
		AppVersion string `json:"app_version"`
	} `json:"data"`
}

// registerGAEVersionCheck polls /wake-up on the GAE service and SIGTERMs this
// process when the live app_version differs from ours, letting GAE route
// traffic to the new instance. No-op outside GAE.
func registerGAEVersionCheck(logger *Logger, hooks *[]func(context.Context) error) {
	service := GAEService()
	if service == "" {
		return
	}

	project, version, interval, ok := resolveGAECheckParams(logger)
	if !ok {
		return
	}

	wakeUpURL := fmt.Sprintf("https://%s-dot-%s.appspot.com/wake-up", service, project)

	ctx, cancel := context.WithCancel(context.Background())
	*hooks = append(*hooks, func(context.Context) error {
		cancel()
		return nil
	})

	go runWakeUpLoop(ctx, logger, wakeUpURL, version, interval)

	logger.LifecycleEvent("GAE_VERSION_CHECK_ENABLED", map[string]any{
		"gae_service":     service,
		"wake_up_url":     wakeUpURL,
		"current_version": version,
		"interval":        interval.String(),
	}, WithComponent(ComponentApp), WithOperation(gaeVersionCheckOp))
}

// resolveGAECheckParams returns (project, version, interval, ok). Logs the
// reason and returns ok=false when any prerequisite is missing.
func resolveGAECheckParams(logger *Logger) (project, version string, interval time.Duration, ok bool) {
	project = strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if project == "" {
		logCheckDisabled(logger, "GOOGLE_CLOUD_PROJECT not set")
		return "", "", 0, false
	}
	version = AppInfo.Version
	if version == "" {
		logCheckDisabled(logger, "APP_VERSION empty")
		return "", "", 0, false
	}
	interval = wakeUpInterval()
	if interval == 0 {
		logCheckDisabled(logger, "WAKE_UP_MINUTE explicitly set to 0")
		return "", "", 0, false
	}
	return project, version, interval, true
}

func logCheckDisabled(logger *Logger, reason string) {
	logger.LifecycleEvent("GAE_VERSION_CHECK_DISABLED", map[string]any{
		"reason": reason,
	}, WithComponent(ComponentApp), WithOperation(gaeVersionCheckOp))
}

func runWakeUpLoop(ctx context.Context, logger *Logger, wakeUpURL, currentVersion string, interval time.Duration) {
	checkGAEVersion(ctx, logger, wakeUpURL, currentVersion)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkGAEVersion(ctx, logger, wakeUpURL, currentVersion)
		}
	}
}

func checkGAEVersion(ctx context.Context, logger *Logger, wakeUpURL, currentVersion string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wakeUpURL, nil)
	if err != nil {
		logCheckFailure(logger, err, "GAE_VERSION_CHECK_BUILD_REQUEST_FAILURE")
		return
	}

	resp, err := wakeUpHTTPClient.Do(req)
	if err != nil {
		logCheckFailure(logger, err, "GAE_VERSION_CHECK_REQUEST_FAILURE")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		logCheckFailure(logger, err, "GAE_VERSION_CHECK_READ_BODY_FAILURE")
		return
	}

	var wr wakeUpResponse
	if err := json.Unmarshal(body, &wr); err != nil {
		logCheckFailure(logger, err, "GAE_VERSION_CHECK_PARSE_FAILURE")
		return
	}

	if wr.Status != 1 {
		return
	}

	live := strings.TrimSpace(wr.Data.AppVersion)
	if live == "" || live == currentVersion {
		return
	}

	logger.LifecycleEvent("GAE_VERSION_MISMATCH_SHUTDOWN", map[string]any{
		"live_version":    live,
		"current_version": currentVersion,
	}, WithComponent(ComponentApp), WithOperation(gaeVersionCheckOp))

	if err := triggerShutdown(); err != nil {
		logger.LifecycleError(err, "GAE_SHUTDOWN_TRIGGER_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation(gaeVersionCheckOp))
	}
}

// logCheckFailure uses LifecycleError so err is captured as a positional
// field (per CLAUDE.md — no duplicate err.Error() into the data map).
func logCheckFailure(logger *Logger, err error, label string) {
	logger.LifecycleError(err, label, nil,
		WithComponent(ComponentApp), WithOperation(gaeVersionCheckOp))
}

func triggerShutdown() error {
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM: %w", err)
	}
	return nil
}
