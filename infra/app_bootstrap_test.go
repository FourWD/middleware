package infra

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

// --- resolveHTTPAddress ----------------------------------------------------

func TestResolveHTTPAddress_UsesHTTPAddressEnv(t *testing.T) {
	t.Setenv("HTTP_ADDRESS", ":9090")
	t.Setenv("PORT", "")
	if got := resolveHTTPAddress(); got != ":9090" {
		t.Fatalf("got %q want %q", got, ":9090")
	}
}

func TestResolveHTTPAddress_NormalizesBareNumber(t *testing.T) {
	t.Setenv("HTTP_ADDRESS", "9090")
	t.Setenv("PORT", "")
	if got := resolveHTTPAddress(); got != ":9090" {
		t.Fatalf("got %q want %q", got, ":9090")
	}
}

func TestResolveHTTPAddress_FallsBackToPort(t *testing.T) {
	_ = os.Unsetenv("HTTP_ADDRESS")
	t.Setenv("PORT", "7000")
	if got := resolveHTTPAddress(); got != ":7000" {
		t.Fatalf("got %q want %q", got, ":7000")
	}
}

func TestResolveHTTPAddress_DefaultPort(t *testing.T) {
	_ = os.Unsetenv("HTTP_ADDRESS")
	_ = os.Unsetenv("PORT")
	if got := resolveHTTPAddress(); got != ":8080" {
		t.Fatalf("got %q want %q", got, ":8080")
	}
}

// --- normalizeHTTPAddress --------------------------------------------------

func TestNormalizeHTTPAddress(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{":8080", ":8080"},
		{"0.0.0.0:8080", "0.0.0.0:8080"},
		{"8080", ":8080"},
		{"localhost:9090", "localhost:9090"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := normalizeHTTPAddress(tc.in); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

// --- buildAppDeps -----------------------------------------------------------

func TestBuildAppDeps_MapsAllSections(t *testing.T) {
	cfg := CommonConfig{AppID: "x", AppEnv: "test"}
	logger := testLogger()
	var hooks []func(context.Context) error
	var workers []Worker
	rate := &RateLimiter{}
	clients := InfraClients{}

	deps := buildAppDeps(cfg, logger, &hooks, clients, rate, nil, &workers)

	if deps.Runtime.Config.AppID != "x" {
		t.Errorf("Config.AppID: got %q want x", deps.Runtime.Config.AppID)
	}
	if deps.Runtime.Logger != logger {
		t.Error("Logger pointer mismatch")
	}
	if deps.Runtime.ShutdownHooks != &hooks {
		t.Error("ShutdownHooks pointer mismatch")
	}
	if deps.Runtime.RateLimit != rate {
		t.Error("RateLimit pointer mismatch")
	}
	if deps.Runtime.HeartbeatDebugStatus == nil {
		t.Error("HeartbeatDebugStatus should be non-nil even when scheduler is nil")
	}
}

func TestBuildAppDeps_HeartbeatNil_StatusReturnsDisabled(t *testing.T) {
	var hooks []func(context.Context) error
	var workers []Worker
	deps := buildAppDeps(CommonConfig{}, testLogger(), &hooks, InfraClients{}, nil, nil, &workers)

	status, ok := deps.Runtime.HeartbeatDebugStatus().(HeartbeatDebugStatus)
	if !ok {
		t.Fatal("expected HeartbeatDebugStatus type")
	}
	if status.Enabled {
		t.Error("expected Enabled=false when heartbeat scheduler is nil")
	}
}

func TestBuildAppDeps_RegisterWorker_AppendsToCallerSlice(t *testing.T) {
	var hooks []func(context.Context) error
	var workers []Worker
	deps := buildAppDeps(CommonConfig{}, testLogger(), &hooks, InfraClients{}, nil, nil, &workers)

	deps.Runtime.RegisterWorker(Worker{Name: "a"})
	deps.Runtime.RegisterWorker(Worker{Name: "b"})

	if len(workers) != 2 {
		t.Fatalf("workers len: got %d want 2", len(workers))
	}
	if workers[0].Name != "a" || workers[1].Name != "b" {
		t.Errorf("workers order mismatch: %+v", workers)
	}
}

// --- runShutdownHooks ------------------------------------------------------

func TestRunShutdownHooks_LIFOOrder(t *testing.T) {
	var order []string
	hooks := []func(context.Context) error{
		func(context.Context) error { order = append(order, "first"); return nil },
		func(context.Context) error { order = append(order, "second"); return nil },
		func(context.Context) error { order = append(order, "third"); return nil },
	}
	runShutdownHooks(context.Background(), hooks, testLogger())
	want := []string{"third", "second", "first"}
	for i, name := range want {
		if i >= len(order) || order[i] != name {
			t.Fatalf("LIFO order broken: got %v want %v", order, want)
		}
	}
}

func TestRunShutdownHooks_ContinuesPastErrors(t *testing.T) {
	var called int
	hooks := []func(context.Context) error{
		func(context.Context) error { called++; return nil },
		func(context.Context) error { called++; return errors.New("boom") },
		func(context.Context) error { called++; return nil },
	}
	runShutdownHooks(context.Background(), hooks, testLogger())
	if called != 3 {
		t.Fatalf("error in one hook should not abort the rest: called=%d want 3", called)
	}
}

// --- App.startWorkers ------------------------------------------------------

func newTestApp(workers ...Worker) *App {
	return &App{
		cfg:     CommonConfig{AppID: "test"},
		fiber:   fiber.New(),
		logger:  testLogger(),
		workers: workers,
	}
}

func TestApp_StartWorkers_NoWorkers(t *testing.T) {
	a := newTestApp()
	wg := a.startWorkers(context.Background())
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("wg.Wait blocked on empty worker list")
	}
}

func TestApp_StartWorkers_WorkerCompletes(t *testing.T) {
	var ran atomic.Bool
	a := newTestApp(Worker{
		Name: "completes",
		Run: func(ctx context.Context) error {
			ran.Store(true)
			return nil
		},
	})
	wg := a.startWorkers(context.Background())
	wg.Wait()
	if !ran.Load() {
		t.Fatal("worker Run was not invoked")
	}
}

func TestApp_StartWorkers_WorkerRespectsCancellation(t *testing.T) {
	started := make(chan struct{})
	a := newTestApp(Worker{
		Name: "blocks",
		Run: func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			return ctx.Err()
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	wg := a.startWorkers(ctx)
	<-started
	cancel()

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not honour ctx.Done")
	}
}

// --- App.waitWorkers -------------------------------------------------------

func TestApp_WaitWorkers_ReturnsWhenAllDone(t *testing.T) {
	a := newTestApp()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		wg.Done()
	}()
	start := time.Now()
	a.waitWorkers(&wg)
	if elapsed := time.Since(start); elapsed > shutdownTimeout {
		t.Fatalf("waitWorkers took %v, expected <%v", elapsed, shutdownTimeout)
	}
}

// --- App.gracefulShutdown --------------------------------------------------

func TestApp_GracefulShutdown_FreshFiberAppSucceeds(t *testing.T) {
	a := newTestApp()
	// Fresh fiber app with no listener — shutdown should still complete
	// without error and run any pending hooks.
	if err := a.gracefulShutdown(); err != nil {
		t.Fatalf("gracefulShutdown: %v", err)
	}
}

func TestApp_GracefulShutdown_RunsHooks(t *testing.T) {
	var ran atomic.Bool
	a := newTestApp()
	a.shutdownHooks = []func(context.Context) error{
		func(context.Context) error { ran.Store(true); return nil },
	}
	if err := a.gracefulShutdown(); err != nil {
		t.Fatalf("gracefulShutdown: %v", err)
	}
	if !ran.Load() {
		t.Fatal("shutdown hook was not invoked")
	}
}

// --- setupObservability ----------------------------------------------------

func TestSetupObservability_DisabledByDefault(t *testing.T) {
	// SENTRY_ENABLED defaults false; OTEL exporter empty -> both no-op.
	t.Setenv("SENTRY_ENABLED", "false")
	t.Setenv("SENTRY_DSN", "")

	logger := testLogger()
	var hooks []func(context.Context) error
	if err := setupObservability(logger, &hooks); err != nil {
		t.Fatalf("setupObservability with everything disabled should not error: %v", err)
	}
	if len(hooks) < 2 {
		t.Fatalf("expected at least 2 hooks (Sentry + OTel noop), got %d", len(hooks))
	}
	for i, h := range hooks {
		if err := h(context.Background()); err != nil {
			t.Errorf("noop hook %d returned error: %v", i, err)
		}
	}
}
