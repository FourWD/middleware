package infra

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/FourWD/middleware/kit"
	"github.com/gofiber/fiber/v3"
)

const (
	BootstrapTimeout = 5 * time.Second
	cleanupTimeout   = 5 * time.Second
	shutdownTimeout  = 10 * time.Second
	HealthTimeout    = 2 * time.Second
)

// MigrationConfig holds database migration settings.
type MigrationConfig struct {
	Enabled bool
	Path    string
}

// FirebaseConfig holds optional Firebase/Firestore/FCM settings.
type FirebaseConfig struct {
	CredentialsFile             string
	NotificationCredentialsFile string
}

type PubSubConfig struct {
	Enabled         bool
	ProjectID       string
	CredentialsFile string
}

type StorageConfig struct {
	Enabled         bool
	Bucket          string
	CredentialsFile string
}

type MailConfig struct {
	Enabled bool
	Domain  string
	APIKey  string
}

// AuthConfig holds JWT and authentication settings.
type AuthConfig struct {
	JWTSecret              string
	JWTRefreshSecret       string
	JWTIssuer              string
	AccessTokenTTLMinutes  int
	RefreshTokenTTLMinutes int
	BlacklistEnabled       bool
}

// CommonConfig holds infrastructure configuration loaded from environment variables.
type CommonConfig struct {
	AppID           string
	AppVersion      string
	AppEnv          string
	LogLevel        string
	Timezone        string
	HTTPAddress     string
	ProxyHeader     string
	DebugAuthToken  string
	PublicPaths     []string
	HTTPBodyLimitMB int

	// Rate limit (3-tier: strict / default / skip)
	RateLimitEnabled          bool
	RateLimitStrictPerMinute  int
	RateLimitDefaultPerSecond int

	Database               DatabaseConfig
	SecondaryDatabase      DatabaseConfig
	Redis                  RedisConfig
	RedisEnabled           bool
	Mongo                  MongoConfig
	MongoEnabled           bool
	MongoMiddleware        MongoConfig
	MongoMiddlewareEnabled bool
	Firebase               FirebaseConfig
	PubSub                 PubSubConfig
	Storage                StorageConfig
	Mail                   MailConfig
	Migration              MigrationConfig
	Auth                   AuthConfig
}

// LoadCommonConfig reads all infrastructure configuration from environment variables.
func LoadCommonConfig() CommonConfig {
	appID := GetEnv("APP_ID", "")
	if gaeService := os.Getenv("GAE_SERVICE"); gaeService != "" {
		appID = gaeService
	}

	return CommonConfig{
		AppID:           appID,
		AppVersion:      strings.TrimSpace(GetEnv("APP_VERSION", "")),
		AppEnv:          strings.ToLower(strings.TrimSpace(GetEnv("APP_ENV", "local"))),
		LogLevel:        strings.ToLower(strings.TrimSpace(GetEnv("LOG_LEVEL", "info"))),
		Timezone:        strings.TrimSpace(GetEnv("APP_TIMEZONE", "Asia/Bangkok")),
		HTTPAddress:     resolveHTTPAddress(),
		ProxyHeader:     resolveProxyHeader(),
		DebugAuthToken:  strings.TrimSpace(GetEnv("HTTP_DEBUG_AUTH_TOKEN", "")),
		PublicPaths:     SplitCSV(GetEnv("HTTP_PUBLIC_PATHS", "")),
		HTTPBodyLimitMB: GetEnvInt("HTTP_BODY_LIMIT_MB", 0),

		RateLimitEnabled:          GetEnvBool("RATE_LIMIT_ENABLED", false),
		RateLimitStrictPerMinute:  GetEnvInt("RATE_LIMIT_STRICT_PER_MINUTE", 10),
		RateLimitDefaultPerSecond: GetEnvInt("RATE_LIMIT_DEFAULT_PER_SECOND", 100),

		Database:               LoadDatabaseConfig(),
		SecondaryDatabase:      LoadSecondaryDatabaseConfig(),
		Redis:                  LoadRedisConfig(),
		RedisEnabled:           GetEnvBool("REDIS_ENABLED", false),
		Mongo:                  LoadMongoConfig(),
		MongoEnabled:           GetEnvBool("MONGO_ENABLED", false),
		MongoMiddleware:        LoadMongoMiddlewareConfig(),
		MongoMiddlewareEnabled: GetEnvBool("MONGO_MIDDLEWARE_ENABLED", false),
		Firebase: FirebaseConfig{
			CredentialsFile: GetEnv("FIREBASE_CREDENTIALS", ""),
			// FCM runs on the same service account as Firestore.
			NotificationCredentialsFile: GetEnv("FIREBASE_CREDENTIALS", ""),
		},
		PubSub: PubSubConfig{
			Enabled: GetEnvBool("PUBSUB_ENABLED", false),
			// Topics live in the service's own project.
			ProjectID:       GetEnv("GOOGLE_CLOUD_PROJECT", ""),
			CredentialsFile: GetEnv("PUBSUB_CREDENTIALS_FILE", ""),
		},
		Storage: StorageConfig{
			Enabled:         GetEnvBool("STORAGE_ENABLED", false),
			Bucket:          GetEnv("STORAGE_BUCKET", ""),
			CredentialsFile: GetEnv("STORAGE_CREDENTIALS_FILE", ""),
		},
		Mail: MailConfig{
			Enabled: GetEnvBool("MAIL_ENABLED", false),
			Domain:  GetEnv("MAILGUN_DOMAIN", ""),
			APIKey:  GetEnv("MAILGUN_API_KEY", ""),
		},
		Migration: MigrationConfig{
			Enabled: GetEnvBool("MIGRATIONS_ENABLED", false),
			Path:    GetEnv("MIGRATIONS_PATH", "migrations"),
		},
		Auth: AuthConfig{
			JWTSecret:              GetEnv("JWT_SECRET", ""),
			JWTRefreshSecret:       GetEnv("JWT_REFRESH_SECRET", ""),
			JWTIssuer:              appID,
			AccessTokenTTLMinutes:  60,
			RefreshTokenTTLMinutes: 10080,
			BlacklistEnabled:       GetEnvBool("JWT_BLACKLIST_ENABLED", false),
		},
	}
}

func resolveProxyHeader() string {
	return strings.TrimSpace(GetEnv("HTTP_PROXY_HEADER", ""))
}

func resolveHTTPAddress() string {
	if httpAddress := strings.TrimSpace(os.Getenv("HTTP_ADDRESS")); httpAddress != "" {
		return normalizeHTTPAddress(httpAddress)
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return ":" + port
	}
	return ":8080"
}

func normalizeHTTPAddress(addr string) string {
	if strings.Contains(addr, ":") {
		return addr
	}
	return ":" + addr
}

type AppRuntimeDeps struct {
	Config               CommonConfig
	Logger               *Logger
	ShutdownHooks        *[]func(context.Context) error
	HeartbeatDebugStatus func() any
	RegisterWorker       func(w Worker)
	RateLimit            *RateLimiter
}

type AppDataDeps struct {
	Databases       Databases
	Redis           *RedisClient
	Mongo           *MongoClient
	MongoMiddleware *MongoClient
}

type AppSecurityDeps struct {
	BlacklistStore BlacklistStore
	// RefreshTokens is auto-wired when MongoMiddleware or Redis is enabled.
	// Plug it into TokenManager via NewTokenManager(cfg, store) for one-shot
	// refresh-token rotation. nil when no compatible backend is configured.
	RefreshTokens RefreshTokenStore
}

type AppCloudDeps struct {
	Firebase *FirebaseClient
	PubSub   *PubSubClient
	Storage  *StorageClient
}

type AppIntegrationDeps struct {
	Mail *MailClient
}

// AppDeps is passed to RouteRegistrar with all initialized infrastructure dependencies.
type AppDeps struct {
	Runtime      AppRuntimeDeps
	Data         AppDataDeps
	Security     AppSecurityDeps
	Cloud        AppCloudDeps
	Integrations AppIntegrationDeps
}

// RouteRegistrar registers middleware and routes on the fiber app.
type RouteRegistrar func(web *fiber.App, deps AppDeps) error

// App is the running application.
type App struct {
	cfg           CommonConfig
	fiber         *fiber.App
	logger        *Logger
	shutdownHooks []func(context.Context) error
	workers       []Worker
}

// NewApp initializes all infrastructure and calls registrar to wire
// project-specific routes. Registers the default middleware stack
// (RequestID → CORS → Sentry → Recover → OTel → Metrics → RequestLog,
// plus Envelope when HTTP_ENVELOPE_ENABLED=true) before invoking registrar.
// Do NOT call RegisterStack again — it is already wired.
//
// AuthenticationMiddleware and MigrateInfra are NOT registered automatically.
// For WebSocket/SSE projects that need realtime routes between base and HTTP
// stack, construct fiber manually using RegisterBaseStack + RegisterHTTPStack
// instead of NewApp.
func NewApp(registrar RouteRegistrar) (*App, error) {
	if err := LoadEnvFiles(); err != nil {
		return nil, err
	}

	// Must run before any config is read: "sm://" values are references, and a
	// service booting with one unresolved would dial MySQL with the reference
	// string as its password.
	if err := kit.ResolveSecretEnv(context.Background()); err != nil {
		return nil, err
	}

	cfg := LoadCommonConfig()
	if err := validateCommonConfig(cfg); err != nil {
		return nil, err
	}

	var shutdownHooks []func(context.Context) error
	appLogger := NewLoggerWith(cfg)
	redirectStdlibLog(appLogger)

	initGlobals(cfg, appLogger)
	setupTimezone(cfg.Timezone, appLogger)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()
		runShutdownHooks(ctx, shutdownHooks, appLogger)
	}

	if err := setupObservability(appLogger, &shutdownHooks); err != nil {
		cleanup()
		return nil, err
	}

	clients, err := initInfrastructure(cfg, appLogger, &shutdownHooks)
	if err != nil {
		appLogger.LifecycleError(err, "APP_INIT_INFRASTRUCTURE_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation("init_infrastructure"))
		cleanup()
		return nil, err
	}

	if err := bindAllGlobals(clients, appLogger); err != nil {
		cleanup()
		return nil, err
	}

	// Best-effort: ensure Mongo auth-store indexes exist. Runs in a goroutine
	// with its own timeout so a slow Mongo (or missing createIndex permission)
	// never blocks boot. Idempotent.
	ensureAuthStoreIndexes(clients, appLogger)

	rateLimiter := buildRateLimiter(cfg, clients.Redis, &shutdownHooks)
	registerDBMetrics(clients.Databases, appLogger)
	registerGAEVersionCheck(appLogger, &shutdownHooks)

	heartbeatScheduler, err := NewHeartbeatScheduler(LoadHeartbeatConfig(), appLogger)
	if err != nil {
		appLogger.LifecycleError(err, "APP_INIT_HEARTBEAT_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation("init_heartbeat"))
		cleanup()
		return nil, err
	}

	web := buildFiberApp(cfg, appLogger)

	var workers []Worker
	deps := buildAppDeps(cfg, appLogger, &shutdownHooks, clients, rateLimiter, heartbeatScheduler, &workers)

	if err := registrar(web, deps); err != nil {
		appLogger.LifecycleError(err, "APP_REGISTER_ROUTES_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation("register_routes"))
		cleanup()
		return nil, err
	}

	if heartbeatScheduler != nil {
		heartbeatScheduler.Start()
		shutdownHooks = append(shutdownHooks, heartbeatScheduler.Shutdown)
	}

	return &App{
		cfg:           cfg,
		fiber:         web,
		logger:        appLogger,
		shutdownHooks: shutdownHooks,
		workers:       workers,
	}, nil
}

// setupObservability initializes Sentry + OTel and appends their shutdown
// functions to hooks. Either failure aborts boot.
func setupObservability(logger *Logger, hooks *[]func(context.Context) error) error {
	sentryShutdown, err := SetupSentry(LoadSentryConfig())
	if err != nil {
		logger.LifecycleError(err, "APP_SETUP_SENTRY_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation("setup_sentry"))
		return err
	}
	*hooks = append(*hooks, sentryShutdown)

	otelShutdown, err := SetupOTel(context.Background(), LoadOTelConfig())
	if err != nil {
		logger.LifecycleError(err, "APP_SETUP_OTEL_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation("setup_otel"))
		return err
	}
	*hooks = append(*hooks, otelShutdown)

	return nil
}

// bindAllGlobals wires the initialized clients into the package-level globals
// that legacy callers read from (common.Database, infra.FirestoreClient, etc.).
// Only the primary Database binding can fail; the rest are best-effort.
func bindAllGlobals(clients InfraClients, logger *Logger) error {
	if err := bindDatabaseGlobal(clients.Databases.Primary); err != nil {
		logger.LifecycleError(err, "APP_BIND_DATABASE_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation("bind_database"))
		return err
	}
	bindFirebaseGlobals(clients.Firebase)
	bindMongoGlobal(clients.Mongo, clients.MongoMiddleware)
	bindPubSubGlobal(clients.PubSub)
	return nil
}

// buildFiberApp constructs the fiber app, registers the default middleware
// stack, mounts request-meta middleware, and registers default routes.
func buildFiberApp(cfg CommonConfig, logger *Logger) *fiber.App {
	web := NewFiberApp(FiberConfig{
		AppID:       cfg.AppID,
		ProxyHeader: cfg.ProxyHeader,
		BodyLimit:   cfg.HTTPBodyLimitMB * 1024 * 1024,
	})

	stackCfg := LoadStackConfig()
	stackCfg.Logger = logger
	RegisterStack(web, stackCfg)

	// Stamp client IP + User-Agent onto ctx so use-case layers can read
	// network identity without threading fiber.Ctx through every signature.
	// Installed after RegisterStack so it sits at the user-middleware scope.
	web.Use(NewRequestMetaMiddleware())

	registerDefaultRoutes(web, cfg)
	return web
}

// buildAppDeps assembles the AppDeps struct passed to the registrar.
// workers is a pointer so RegisterWorker can append after this returns.
func buildAppDeps(
	cfg CommonConfig,
	logger *Logger,
	hooks *[]func(context.Context) error,
	clients InfraClients,
	rateLimiter *RateLimiter,
	heartbeat *HeartbeatScheduler,
	workers *[]Worker,
) AppDeps {
	return AppDeps{
		Runtime: AppRuntimeDeps{
			Config:        cfg,
			Logger:        logger,
			ShutdownHooks: hooks,
			RateLimit:     rateLimiter,
			HeartbeatDebugStatus: func() any {
				if heartbeat == nil {
					return HeartbeatDebugStatus{Enabled: false}
				}
				return heartbeat.DebugStatus()
			},
			RegisterWorker: func(w Worker) {
				*workers = append(*workers, w)
			},
		},
		Data: AppDataDeps{
			Databases:       clients.Databases,
			Redis:           clients.Redis,
			Mongo:           clients.Mongo,
			MongoMiddleware: clients.MongoMiddleware,
		},
		Security: AppSecurityDeps{
			BlacklistStore: clients.Blacklist,
			RefreshTokens:  clients.RefreshTokens,
		},
		Cloud: AppCloudDeps{
			Firebase: clients.Firebase,
			PubSub:   clients.PubSub,
			Storage:  clients.Storage,
		},
		Integrations: AppIntegrationDeps{
			Mail: clients.Mail,
		},
	}
}

// Run starts workers + the HTTP server and blocks until a shutdown signal
// is received. On shutdown: cancel workers → wait bounded → shutdown fiber
// → run cleanup hooks.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	engineCtx, engineCancel := context.WithCancel(context.Background())
	defer engineCancel()

	wg := a.startWorkers(engineCtx)
	serverErr := a.startHTTPServer()

	select {
	case err := <-serverErr:
		a.logger.LifecycleError(err, "APP_HTTP_SERVER_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation("http_server_run"))
		engineCancel()
		a.waitWorkers(wg)
		return err
	case <-ctx.Done():
		a.logger.LifecycleEvent("APP_SHUTDOWN_SIGNAL", nil,
			WithComponent(ComponentApp), WithOperation("shutdown_signal"))
	}

	engineCancel()
	a.waitWorkers(wg)
	return a.gracefulShutdown()
}

func (a *App) startWorkers(ctx context.Context) *sync.WaitGroup {
	var wg sync.WaitGroup
	for _, w := range a.workers {
		wg.Add(1)
		go func(worker Worker) {
			defer wg.Done()
			a.logger.LifecycleEvent("APP_WORKER_STARTED", nil,
				WithField("worker", worker.Name),
				WithComponent(ComponentApp), WithOperation("worker_start"))
			if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.LifecycleError(err, "APP_WORKER_STOPPED_WITH_ERROR", nil,
					WithField("worker", worker.Name),
					WithComponent(ComponentApp), WithOperation("worker_run"))
				return
			}
			a.logger.LifecycleEvent("APP_WORKER_STOPPED", nil,
				WithField("worker", worker.Name),
				WithComponent(ComponentApp), WithOperation("worker_stop"))
		}(w)
	}
	return &wg
}

func (a *App) startHTTPServer() <-chan error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.LifecycleEvent("APP_HTTP_SERVER_START", map[string]any{
			"address": a.cfg.HTTPAddress,
		}, WithComponent(ComponentApp), WithOperation("http_server_start"))
		errCh <- a.fiber.Listen(a.cfg.HTTPAddress, fiber.ListenConfig{
			DisableStartupMessage: a.cfg.AppEnv != "local",
		})
	}()
	return errCh
}

func (a *App) gracefulShutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.fiber.ShutdownWithContext(shutdownCtx); err != nil {
		a.logger.LifecycleError(err, "APP_HTTP_SHUTDOWN_FAILURE", nil,
			WithComponent(ComponentApp), WithOperation("http_server_shutdown"))
		return fmt.Errorf("shutdown fiber app: %w", err)
	}

	runShutdownHooks(shutdownCtx, a.shutdownHooks, a.logger)
	return nil
}

func (a *App) waitWorkers(wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(shutdownTimeout):
		a.logger.LifecycleWarn("APP_WORKERS_SHUTDOWN_TIMEOUT", nil,
			WithComponent(ComponentApp), WithOperation("workers_shutdown"))
	}
}

func runShutdownHooks(ctx context.Context, hooks []func(context.Context) error, appLogger *Logger) {
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := hooks[i](ctx); err != nil {
			appLogger.LifecycleError(err, "APP_CLEANUP_HOOK_FAILURE", nil,
				WithComponent(ComponentApp), WithOperation("cleanup_hook"))
		}
	}
}

// ensureAuthStoreIndexes asks the Mongo-backed auth stores to create their
// lookup + TTL indexes. Wrapped in a goroutine with its own deadline so
// startup never blocks on Mongo; permission errors are logged and the
// service still boots. Idempotent — safe to re-run.
func ensureAuthStoreIndexes(clients InfraClients, logger *Logger) {
	mbs, mbsOK := clients.Blacklist.(*MongoBlacklistStore)
	mrs, mrsOK := clients.RefreshTokens.(*MongoRefreshTokenStore)
	if !mbsOK && !mrsOK {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if mbsOK {
			if err := mbs.EnsureIndexes(ctx); err != nil {
				logger.LifecycleError(err, "AUTH_BLACKLIST_ENSURE_INDEXES_FAILURE", nil,
					WithComponent(ComponentAuth), WithOperation("ensure_indexes"),
					WithField("store", "blacklist"))
			}
		}
		if mrsOK {
			if err := mrs.EnsureIndexes(ctx); err != nil {
				logger.LifecycleError(err, "AUTH_REFRESH_ENSURE_INDEXES_FAILURE", nil,
					WithComponent(ComponentAuth), WithOperation("ensure_indexes"),
					WithField("store", "refresh_tokens"))
			}
		}
	}()
}

var allowedAppEnvs = map[string]struct{}{
	"local": {},
	"dev":   {},
	"prod":  {},
	"test":  {},
}

func validateCommonConfig(cfg CommonConfig) error {
	if _, ok := allowedAppEnvs[cfg.AppEnv]; !ok {
		return fmt.Errorf("invalid APP_ENV: %q (allowed: local, dev, prod, test)", cfg.AppEnv)
	}
	if strings.TrimSpace(cfg.AppID) == "" {
		return fmt.Errorf("invalid APP_ID")
	}
	if strings.TrimSpace(cfg.HTTPAddress) == "" {
		return fmt.Errorf("invalid HTTP_ADDRESS")
	}
	if cfg.DebugAuthToken != "" && len(cfg.DebugAuthToken) < 16 {
		return fmt.Errorf("invalid HTTP_DEBUG_AUTH_TOKEN: must be at least 16 characters")
	}
	// Primary database is opt-out via empty DB_NAME — services that don't
	// need a database (gateways, webhook relays) can leave DB_NAME blank.
	if strings.TrimSpace(cfg.Database.Name) != "" {
		if err := validateDatabaseConfig("DB", cfg.Database); err != nil {
			return err
		}
	}
	if strings.TrimSpace(cfg.SecondaryDatabase.Name) != "" {
		if err := validateDatabaseConfig("DB2", cfg.SecondaryDatabase); err != nil {
			return err
		}
	}
	if cfg.RedisEnabled {
		if strings.TrimSpace(cfg.Redis.Addr) == "" {
			return fmt.Errorf("invalid REDIS_ADDR")
		}
		if cfg.Redis.DB < 0 {
			return fmt.Errorf("invalid REDIS_DB")
		}
	}
	if cfg.Migration.Enabled && strings.TrimSpace(cfg.Migration.Path) == "" {
		return fmt.Errorf("invalid MIGRATIONS_PATH")
	}
	if strings.TrimSpace(cfg.Auth.JWTSecret) == "" {
		return fmt.Errorf("invalid JWT_SECRET")
	}
	if cfg.Auth.BlacklistEnabled && !cfg.MongoMiddlewareEnabled && !cfg.RedisEnabled {
		return fmt.Errorf("JWT_BLACKLIST_ENABLED requires MONGO_MIDDLEWARE_ENABLED or REDIS_ENABLED")
	}
	if cfg.PubSub.Enabled && strings.TrimSpace(cfg.PubSub.ProjectID) == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT is required when PUBSUB_ENABLED=true")
	}
	if cfg.Storage.Enabled && strings.TrimSpace(cfg.Storage.Bucket) == "" {
		return fmt.Errorf("STORAGE_BUCKET is required when STORAGE_ENABLED=true")
	}
	if cfg.Mail.Enabled {
		if strings.TrimSpace(cfg.Mail.Domain) == "" {
			return fmt.Errorf("MAILGUN_DOMAIN is required when MAIL_ENABLED=true")
		}
		if strings.TrimSpace(cfg.Mail.APIKey) == "" {
			return fmt.Errorf("MAILGUN_API_KEY is required when MAIL_ENABLED=true")
		}
	}
	return nil
}

func validateDatabaseConfig(prefix string, cfg DatabaseConfig) error {
	switch cfg.Driver {
	case DBDriverMySQL, DBDriverPostgres:
	default:
		return fmt.Errorf("invalid %s_DRIVER", prefix)
	}

	// MySQL via Cloud SQL Unix socket skips host/port validation.
	if cfg.Driver == DBDriverMySQL && strings.TrimSpace(cfg.Instance) != "" {
		return validateDatabaseFields(prefix, cfg, false)
	}
	return validateDatabaseFields(prefix, cfg, true)
}

// validateDatabaseFields validates a DatabaseConfig. requireHostPort=false
// skips host/port (used by Cloud SQL Unix-socket connections).
func validateDatabaseFields(prefix string, cfg DatabaseConfig, requireHostPort bool) error {
	if requireHostPort {
		if strings.TrimSpace(cfg.Host) == "" {
			return fmt.Errorf("invalid %s_HOST", prefix)
		}
		if cfg.Port <= 0 {
			return fmt.Errorf("invalid %s_PORT", prefix)
		}
	}
	if strings.TrimSpace(cfg.User) == "" {
		return fmt.Errorf("invalid %s_USER", prefix)
	}
	if strings.TrimSpace(cfg.Name) == "" {
		return fmt.Errorf("invalid %s_NAME", prefix)
	}
	if cfg.MaxIdleConns <= 0 {
		return fmt.Errorf("invalid %s_MAX_IDLE_CONNS", prefix)
	}
	if cfg.MaxOpenConns <= 0 {
		return fmt.Errorf("invalid %s_MAX_OPEN_CONNS", prefix)
	}
	if cfg.MaxLifetime <= 0 {
		return fmt.Errorf("invalid %s_MAX_LIFETIME_MINUTES", prefix)
	}
	return nil
}
