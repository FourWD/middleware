package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)

const gaeVersionCheckInterval = 1 * time.Minute

type wakeUpResponse struct {
	Status int `json:"status"`
	Data   struct {
		AppVersion string `json:"app_version"`
	} `json:"data"`
}

// registerGAEVersionCheck polls the GAE service's /wake-up endpoint every minute
// and sends SIGTERM to the current process when the live app_version differs from
// this instance's SERVICE_VERSION, letting App.Run() drain traffic gracefully while
// GAE routes to the fresh instance.
//
// No-op outside Google App Engine (detected via GAE_SERVICE env).
func registerGAEVersionCheck(cfg CommonConfig, logger *Logger, hooks *[]func(context.Context) error) {
	service := GAEService()
	if service == "" {
		return
	}

	project := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if project == "" {
		project = cfg.GCPProjectID
	}
	if project == "" {
		logger.Warn(M("gae version check disabled: project id unknown"),
			WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("startup"))
		return
	}

	currentVersion := AppInfo.Version
	if currentVersion == "" {
		logger.Warn(M("gae version check disabled: APP_VERSION empty"),
			WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("startup"))
		return
	}

	wakeUpURL := fmt.Sprintf("https://%s-dot-%s.appspot.com/wake-up", service, project)

	ctx, cancel := context.WithCancel(context.Background())
	*hooks = append(*hooks, func(context.Context) error {
		cancel()
		return nil
	})

	go runGAEVersionLoop(ctx, logger, wakeUpURL, currentVersion)

	logger.Info(M("gae version check enabled"),
		WithField("gae_service", service),
		WithField("wake_up_url", wakeUpURL),
		WithField("current_version", currentVersion),
		WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("startup"))
}

func runGAEVersionLoop(ctx context.Context, logger *Logger, wakeUpURL, currentVersion string) {
	checkGAEVersion(ctx, logger, wakeUpURL, currentVersion)

	ticker := time.NewTicker(gaeVersionCheckInterval)
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
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wakeUpURL, nil)
	if err != nil {
		logger.Warn(M("gae version check: build request failed"),
			WithField("error", err.Error()),
			WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("runtime"))
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn(M("gae version check: request failed"),
			WithField("error", err.Error()),
			WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("runtime"))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		logger.Warn(M("gae version check: read body failed"),
			WithField("error", err.Error()),
			WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("runtime"))
		return
	}

	var wr wakeUpResponse
	if err := json.Unmarshal(body, &wr); err != nil {
		logger.Warn(M("gae version check: parse body failed"),
			WithField("error", err.Error()),
			WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("runtime"))
		return
	}

	if wr.Status != 1 {
		return
	}

	live := strings.TrimSpace(wr.Data.AppVersion)
	if live == "" || live == currentVersion {
		return
	}

	logger.Info(M("gae version mismatch, triggering shutdown"),
		WithField("live_version", live),
		WithField("current_version", currentVersion),
		WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("lifecycle"))

	if err := triggerShutdown(); err != nil {
		logger.Error(err, M("trigger shutdown failed"),
			WithComponent("app"), WithOperation("gae_version_check"), WithLogKind("lifecycle"))
	}
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
