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
	AppID              string
	AppVersion         string
	AppEnv             string
	LogLevel           string
	Timezone           string
	GCPProjectID       string
	HTTPAddress        string
	ProxyHeader        string
	DebugAuthToken     string
	PublicPaths        []string

	// Rate limit (3-tier: strict / default / skip)
	RateLimitEnabled          bool
	RateLimitStrictPerMinute  int
	RateLimitDefaultPerSecond int

	Database           DatabaseConfig
	SecondaryDatabase  DatabaseConfig
	SecondaryDBEnabled bool
	Redis              RedisConfig
	RedisEnabled       bool
	Mongo              MongoConfig
	MongoEnabled       bool
	Firebase           FirebaseConfig
	PubSub             PubSubConfig
	Storage            StorageConfig
	Mail               MailConfig
	Migration          MigrationConfig
	Auth               AuthConfig
}

// LoadCommonConfig reads all infrastructure configuration from environment variables.
func LoadCommonConfig() CommonConfig {
	appID := GetEnv("APP_ID", "")
	if gaeService := os.Getenv("GAE_SERVICE"); gaeService != "" {
		appID = gaeService
	}

	return CommonConfig{
		AppID:              appID,
		AppVersion:         strings.TrimSpace(GetEnv("APP_VERSION", "")),
		AppEnv:             strings.ToLower(strings.TrimSpace(GetEnv("APP_ENV", "local"))),
		LogLevel:           strings.ToLower(strings.TrimSpace(GetEnv("LOG_LEVEL", "info"))),
		Timezone:           strings.TrimSpace(GetEnv("APP_TIMEZONE", "Asia/Bangkok")),
		GCPProjectID:       strings.TrimSpace(GetEnv("GCP_PROJECT_ID", "")),
		HTTPAddress:        resolveHTTPAddress(),
		ProxyHeader:        resolveProxyHeader(),
		DebugAuthToken:     strings.TrimSpace(GetEnv("HTTP_DEBUG_AUTH_TOKEN", "")),
		PublicPaths:        splitCSV(GetEnv("HTTP_PUBLIC_PATHS", "")),

		RateLimitEnabled:          GetEnvBool("RATE_LIMIT_ENABLED", false),
		RateLimitStrictPerMinute:  GetEnvInt("RATE_LIMIT_STRICT_PER_MINUTE", 10),
		RateLimitDefaultPerSecond: GetEnvInt("RATE_LIMIT_DEFAULT_PER_SECOND", 100),

		Database:           LoadDatabaseConfig(),
		SecondaryDBEnabled: GetEnvBool("DB2_ENABLED", false),
		SecondaryDatabase:  LoadSecondaryDatabaseConfig(),
		Redis:              LoadRedisConfig(),
		RedisEnabled:       GetEnvBool("REDIS_ENABLED", false),
		Mongo:              LoadMongoConfig(),
		MongoEnabled:       GetEnvBool("MONGO_ENABLED", false),
		Firebase: FirebaseConfig{
			CredentialsFile:             GetEnv("FIREBASE_CREDENTIALS", ""),
			NotificationCredentialsFile: GetEnv("FIREBASE_NOTIFICATION_CREDENTIALS", ""),
		},
		PubSub: PubSubConfig{
			Enabled:         GetEnvBool("PUBSUB_ENABLED", false),
			ProjectID:       GetEnv("PUBSUB_PROJECT_ID", ""),
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
	Databases Databases
	Redis     *RedisClient
	Mongo     *MongoClient
}

type AppSecurityDeps struct {
	BlacklistStore BlacklistStore
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
// It receives all initialized infrastructure dependencies.
// Implement this in your project's internal/app package.
type RouteRegistrar func(web *fiber.App, deps AppDeps) error

// App is the running application.
type App struct {
	cfg           CommonConfig
	fiber         *fiber.App
	logger        *Logger
	shutdownHooks []func(context.Context) error
	workers       []Worker
}

// NewApp initializes all infrastructure and calls registrar to wire project-specific routes.
// It loads configuration from environment variables automatically and registers the
// default middleware stack (RequestID → CORS → Sentry → Recover → OTel → Metrics →
// RequestLog, plus Envelope when HTTP_ENVELOPE_ENABLED=true) before invoking registrar.
// Do NOT call RegisterStack again inside your registrar — it is already wired.
//
// AuthenticationMiddleware and MigrateInfra are NOT registered automatically. Call
// them yourself from registrar when needed:
//
//	func Register(app *fiber.App, deps infra.AppDeps) error {
//	    if err := infra.MigrateInfra(deps); err != nil {
//	        return err
//	    }
//	    app.Use(infra.AuthenticationMiddleware)
//	    router.SetupRoutes(app)
//	    return nil
//	}
//
// For WebSocket/SSE projects that need realtime routes registered between base and
// HTTP stack, construct the fiber app manually using RegisterBaseStack +
// RegisterHTTPStack instead of using NewApp.
func NewApp(registrar RouteRegistrar) (*App, error) {
	if err := LoadEnvFiles(); err != nil {
		return nil, err
	}

	cfg := LoadCommonConfig()
	if err := validateCommonConfig(cfg); err != nil {
		return nil, err
	}

	var shutdownHooks []func(context.Context) error
	appLogger := NewLoggerWith(cfg)

	initGlobals(cfg, appLogger)
	setupTimezone(cfg.Timezone, appLogger)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()
		runShutdownHooks(ctx, shutdownHooks, appLogger)
	}

	sentryShutdown, err := SetupSentry(LoadSentryConfig())
	if err != nil {
		appLogger.Error(err, M("setup sentry failed"), WithComponent("app"), WithOperation("setup_sentry"), WithLogKind("startup"))
		cleanup()
		return nil, err
	}
	shutdownHooks = append(shutdownHooks, sentryShutdown)

	otelShutdown, err := SetupOTel(context.Background(), LoadOTelConfig())
	if err != nil {
		appLogger.Error(err, M("setup otel failed"), WithComponent("app"), WithOperation("setup_otel"), WithLogKind("startup"))
		cleanup()
		return nil, err
	}
	shutdownHooks = append(shutdownHooks, otelShutdown)

	dbs, redisClient, mongoClient, firebaseClient, blacklistStore, pubSubClient, storageClient, mailClient, err := initInfrastructure(cfg, appLogger, &shutdownHooks)
	if err != nil {
		appLogger.Error(err, M("init infrastructure failed"), WithComponent("app"), WithOperation("init_infrastructure"), WithLogKind("startup"))
		cleanup()
		return nil, err
	}

	bindFirebaseGlobals(firebaseClient)
	bindMongoGlobal(mongoClient)

	rateLimiter := buildRateLimiter(cfg, redisClient, &shutdownHooks)

	registerDBMetrics(dbs, appLogger)
	registerGAEVersionCheck(cfg, appLogger, &shutdownHooks)

	heartbeatScheduler, err := NewHeartbeatScheduler(LoadHeartbeatConfig(), appLogger)
	if err != nil {
		appLogger.Error(err, M("init heartbeat failed"), WithComponent("app"), WithOperation("init_heartbeat"), WithLogKind("startup"))
		cleanup()
		return nil, err
	}

	web := NewFiberApp(FiberConfig{
		AppID:       cfg.AppID,
		ProxyHeader: cfg.ProxyHeader,
	})

	stackCfg := LoadStackConfig()
	stackCfg.Logger = appLogger
	RegisterStack(web, stackCfg)

	registerDefaultRoutes(web, cfg)

	var workers []Worker

	deps := AppDeps{
		Runtime: AppRuntimeDeps{
			Config:        cfg,
			Logger:        appLogger,
			ShutdownHooks: &shutdownHooks,
			RateLimit: rateLimiter,
			HeartbeatDebugStatus: func() any {
				if heartbeatScheduler == nil {
					return HeartbeatDebugStatus{Enabled: false}
				}
				return heartbeatScheduler.DebugStatus()
			},
			RegisterWorker: func(w Worker) {
				workers = append(workers, w)
			},
		},
		Data: AppDataDeps{
			Databases: dbs,
			Redis:     redisClient,
			Mongo:     mongoClient,
		},
		Security: AppSecurityDeps{
			BlacklistStore: blacklistStore,
		},
		Cloud: AppCloudDeps{
			Firebase: firebaseClient,
			PubSub:   pubSubClient,
			Storage:  storageClient,
		},
		Integrations: AppIntegrationDeps{
			Mail: mailClient,
		},
	}

	if err := registrar(web, deps); err != nil {
		appLogger.Error(err, M("register routes failed"), WithComponent("app"), WithOperation("register_routes"), WithLogKind("startup"))
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

// Run starts workers + the HTTP server and blocks until a shutdown signal is received.
// On shutdown: cancel workers → wait bounded → shutdown fiber → run cleanup hooks.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	engineCtx, engineCancel := context.WithCancel(context.Background())
	defer engineCancel()

	var wg sync.WaitGroup
	for _, w := range a.workers {
		wg.Add(1)
		go func(worker Worker) {
			defer wg.Done()
			a.logger.Info(M("worker started"),
				WithField("worker", worker.Name),
				WithComponent("app"), WithOperation("worker_start"), WithLogKind("lifecycle"))
			if err := worker.Run(engineCtx); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error(err, M("worker stopped with error"),
					WithField("worker", worker.Name),
					WithComponent("app"), WithOperation("worker_run"), WithLogKind("lifecycle"))
				return
			}
			a.logger.Info(M("worker stopped"),
				WithField("worker", worker.Name),
				WithComponent("app"), WithOperation("worker_stop"), WithLogKind("lifecycle"))
		}(w)
	}

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info(M("starting http server"), WithField("address", a.cfg.HTTPAddress), WithComponent("app"), WithOperation("http_server_start"), WithLogKind("lifecycle"))
		errCh <- a.fiber.Listen(a.cfg.HTTPAddress, fiber.ListenConfig{
			DisableStartupMessage: a.cfg.AppEnv != "local",
		})
	}()

	select {
	case err := <-errCh:
		a.logger.Error(err, M("http server failed"), WithComponent("app"), WithOperation("http_server_run"), WithLogKind("lifecycle"))
		engineCancel()
		a.waitWorkers(&wg)
		return err
	case <-ctx.Done():
		a.logger.Info(M("shutdown signal received"), WithComponent("app"), WithOperation("shutdown_signal"), WithLogKind("lifecycle"))
	}

	engineCancel()
	a.waitWorkers(&wg)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.fiber.ShutdownWithContext(shutdownCtx); err != nil {
		a.logger.Error(err, M("shutdown fiber app failed"), WithComponent("app"), WithOperation("http_server_shutdown"), WithLogKind("lifecycle"))
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
		a.logger.Warn(M("workers shutdown timeout"),
			WithComponent("app"), WithOperation("workers_shutdown"), WithLogKind("lifecycle"))
	}
}

func runShutdownHooks(ctx context.Context, hooks []func(context.Context) error, appLogger *Logger) {
	for i := len(hooks) - 1; i >= 0; i-- {
		if err := hooks[i](ctx); err != nil {
			appLogger.Error(err, M("cleanup hook failed"), WithComponent("app"), WithOperation("cleanup_hook"), WithLogKind("lifecycle"))
		}
	}
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
	if err := validateDatabaseConfig("DB", cfg.Database); err != nil {
		return err
	}
	if cfg.SecondaryDBEnabled {
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
	if cfg.Auth.BlacklistEnabled && !cfg.MongoEnabled && !cfg.RedisEnabled {
		return fmt.Errorf("JWT_BLACKLIST_ENABLED requires MONGO_ENABLED=true or REDIS_ENABLED=true")
	}
	if cfg.PubSub.Enabled && strings.TrimSpace(cfg.PubSub.ProjectID) == "" {
		return fmt.Errorf("PUBSUB_PROJECT_ID is required when PUBSUB_ENABLED=true")
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

	if cfg.Driver == DBDriverMySQL && strings.TrimSpace(cfg.Instance) != "" {
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

	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("invalid %s_HOST", prefix)
	}
	if cfg.Port <= 0 {
		return fmt.Errorf("invalid %s_PORT", prefix)
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
