package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver "pgx" for the lock connection
)

// Leader holds a session-scoped database lock so a multi-instance deploy stays
// single-writer: only the holder runs the work its Run guards. The lock lives
// on a *dedicated connection* — the moment that connection drops (process
// death, network blip, server restart, failed keepalive) the server releases
// it and a standby claims it on its next attempt. That is the whole point of a
// session lock over a row-based lease: no stale-leader window, no clock-skew
// worry, no cleanup job.
//
// Both dialects are supported and behave the same from the caller's side, but
// the primitives differ in ways worth knowing rather than discovering:
//
//   - MySQL GET_LOCK takes a name and blocks up to a timeout. PostgreSQL
//     advisory locks take a 64-bit integer and have no timeout variant, so the
//     Postgres path polls pg_try_advisory_lock until LockTimeout elapses.
//     Failover latency stays ~(LockTimeout + RetryInterval) either way.
//   - LockName is hashed to that integer for Postgres. Two names could in
//     principle collide, and the key space is shared with anything else using
//     advisory locks on the database — so the name still wants to be specific.
//   - GET_LOCK is scoped to the MySQL *server*; a PostgreSQL advisory lock is
//     scoped to the *database*. Instances pointed at different databases on one
//     Postgres server do not contend. That is a loosening, not a tightening.
//
// The lock pool is built separately from the app's primary pool on purpose: a
// pooled connection returned to the pool still carries the session lock, and
// any later reuse or pool recycle would silently release it.
type Leader struct {
	cfg      LeaderConfig
	log      Scoped
	instance string
	driver   string
	key      int64
	disabled bool

	db *sql.DB

	flag      atomic.Bool
	gained    chan struct{}
	gainedOne sync.Once

	mu          sync.RWMutex
	sinceAt     time.Time
	lastErr     string
	lastAttempt time.Time
}

// LeaderConfig tunes the lock. Zero durations fall back to the defaults
// LoadLeaderConfig uses, so a caller can set only the fields it cares about.
type LeaderConfig struct {
	// Enabled false makes the Leader report IsLeader() == true forever and
	// Run a no-op — the correct setting for a one-instance deploy, where the
	// lock adds a dependency and no safety.
	Enabled bool

	// LockName must be unique per workload. Everything sharing a name
	// contends for one leadership, which is the point — but two unrelated
	// services sharing one by accident means one of them never runs.
	LockName string

	// LockTimeout bounds a single acquire attempt.
	LockTimeout time.Duration

	// RetryInterval is how long a standby waits before re-attempting after a
	// failed or timed-out acquire.
	RetryInterval time.Duration

	// KeepaliveInterval is how often the holder pings its lock connection to
	// notice a silently-dropped socket. It doubles as the ping deadline, so a
	// momentarily slow server does not drop leadership unnecessarily.
	KeepaliveInterval time.Duration

	// Driver and DSN default to the primary database (LoadDatabaseConfig) when
	// empty. Override only when the lock must live somewhere other than the
	// app's own database.
	Driver string
	DSN    string
}

const (
	defaultLeaderLockTimeout       = 5 * time.Second
	defaultLeaderRetryInterval     = 3 * time.Second
	defaultLeaderKeepaliveInterval = 30 * time.Second

	// leaderPollInterval is how often the Postgres path retries
	// pg_try_advisory_lock inside one acquire attempt. MySQL needs no
	// equivalent — GET_LOCK blocks server-side for the whole timeout.
	leaderPollInterval = time.Second
)

// LoadLeaderConfig reads LEADER_* from the environment. LEADER_LOCK_NAME
// defaults to "<APP_ID>:leader:v1" so two different services do not collide
// out of the box; set it explicitly when several deploys of one service must
// share a leadership.
func LoadLeaderConfig() LeaderConfig {
	name := GetEnv("LEADER_LOCK_NAME", "")
	if name == "" {
		app := GetEnv("APP_ID", "app")
		name = app + ":leader:v1"
	}
	return LeaderConfig{
		Enabled:           GetEnvBool("LEADER_ENABLED", true),
		LockName:          name,
		LockTimeout:       time.Duration(GetEnvInt("LEADER_LOCK_TIMEOUT_SECS", 5)) * time.Second,
		RetryInterval:     time.Duration(GetEnvInt("LEADER_RETRY_INTERVAL_SECS", 3)) * time.Second,
		KeepaliveInterval: time.Duration(GetEnvInt("LEADER_KEEPALIVE_INTERVAL_SECS", 30)) * time.Second,
	}
}

// LeaderState is the JSON view of lock status, for a service's status endpoint
// so an operator can see who holds the lock without querying the database.
type LeaderState struct {
	Enabled       bool      `json:"enabled"`
	IsLeader      bool      `json:"is_leader"`
	InstanceID    string    `json:"instance_id"`
	LockName      string    `json:"lock_name"`
	SinceAt       time.Time `json:"since_at,omitempty"`
	LastAttemptAt time.Time `json:"last_attempt_at,omitempty"`
	LastError     string    `json:"last_error,omitempty"`
}

// NewLeader builds the dedicated single-connection pool and returns a Leader
// that is safe to call IsLeader on immediately. The pool stays empty until Run
// acquires a connection.
func NewLeader(cfg LeaderConfig, logger *Logger) (*Leader, error) {
	cfg = cfg.normalized()
	log := logger.Scoped(ComponentLeader)
	instance := LeaderInstanceID()

	if !cfg.Enabled {
		l := &Leader{cfg: cfg, log: log, instance: instance, disabled: true, gained: make(chan struct{})}
		l.flag.Store(true)
		close(l.gained)
		return l, nil
	}
	if cfg.LockName == "" {
		return nil, errors.New("leader: LockName is empty")
	}

	driver, dsn, err := cfg.resolveDSN()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("leader: open lock pool: %w", err)
	}
	// One connection, never recycled: the lock is connection-scoped, so a pool
	// that swaps connections underneath would drop leadership silently. The
	// keepalive ping is what catches a genuinely dead connection.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	db.SetConnMaxIdleTime(0)

	return &Leader{
		cfg:      cfg,
		log:      log,
		instance: instance,
		driver:   driver,
		key:      leaderLockKey(cfg.LockName),
		db:       db,
		gained:   make(chan struct{}),
	}, nil
}

func (c LeaderConfig) normalized() LeaderConfig {
	if c.LockTimeout <= 0 {
		c.LockTimeout = defaultLeaderLockTimeout
	}
	if c.RetryInterval <= 0 {
		c.RetryInterval = defaultLeaderRetryInterval
	}
	if c.KeepaliveInterval <= 0 {
		c.KeepaliveInterval = defaultLeaderKeepaliveInterval
	}
	return c
}

// resolveDSN maps the configured driver to a database/sql driver name and a
// DSN, falling back to the primary database when the caller supplied neither.
func (c LeaderConfig) resolveDSN() (driver, dsn string, err error) {
	dbCfg := LoadDatabaseConfig()
	kind := c.Driver
	if kind == "" {
		kind = dbCfg.Driver
	}
	switch kind {
	case DBDriverPostgres:
		if c.DSN != "" {
			return "pgx", c.DSN, nil
		}
		return "pgx", BuildPostgresDSN(dbCfg), nil
	case DBDriverMySQL, "":
		if c.DSN != "" {
			return DBDriverMySQL, c.DSN, nil
		}
		return DBDriverMySQL, BuildMySQLDSN(dbCfg), nil
	default:
		return "", "", fmt.Errorf("leader: unsupported driver %q", kind)
	}
}

// Gained is closed exactly once, when this instance first acquires the lock.
// Work that was skipped while standing by can select on it to start
// immediately instead of waiting for the next tick of its own loop.
func (l *Leader) Gained() <-chan struct{} { return l.gained }

func (l *Leader) IsLeader() bool { return l.flag.Load() }

func (l *Leader) State() LeaderState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return LeaderState{
		Enabled:       !l.disabled,
		IsLeader:      l.flag.Load(),
		InstanceID:    l.instance,
		LockName:      l.cfg.LockName,
		SinceAt:       l.sinceAt,
		LastAttemptAt: l.lastAttempt,
		LastError:     l.lastErr,
	}
}

// Run is the Worker body. With the lock disabled it blocks on ctx so the
// harness still has something to wait on; otherwise it loops acquiring the
// lock, holding it with keepalives, and releasing on shutdown.
func (l *Leader) Run(ctx context.Context) error {
	if l.disabled {
		<-ctx.Done()
		return nil
	}
	l.log.Lifecycle("LEADER_ARMED", map[string]any{
		"instance":    l.instance,
		"lock_name":   l.cfg.LockName,
		"driver":      l.driver,
		"timeout":     l.cfg.LockTimeout.String(),
		"keepalive":   l.cfg.KeepaliveInterval.String(),
		"retry_after": l.cfg.RetryInterval.String(),
	}, WithOperation("arm"))

	defer func() {
		l.flag.Store(false)
		l.mu.Lock()
		l.sinceAt = time.Time{}
		l.mu.Unlock()
		if l.db != nil {
			_ = l.db.Close()
		}
	}()

	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := l.acquireAndHold(ctx); err != nil && !errors.Is(err, context.Canceled) {
			l.recordError(err)
			l.log.Warn(ctx, "LEADER_CYCLE_ENDED", map[string]any{
				"instance": l.instance,
				"reason":   err.Error(),
			}, WithOperation("acquire"), WithLogKind(LogKindLifecycle))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(l.cfg.RetryInterval):
		}
	}
}

// acquireAndHold blocks until either leadership is lost (keepalive fails, ctx
// done) or the acquire times out without winning. A timeout returns nil so Run
// simply sleeps and re-attempts; losing a held lock is what gets logged.
func (l *Leader) acquireAndHold(ctx context.Context) error {
	conn, err := l.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("checkout: %w", err)
	}
	defer conn.Close()

	l.mu.Lock()
	l.lastAttempt = time.Now()
	l.mu.Unlock()

	got, err := l.acquire(ctx, conn)
	if err != nil {
		return err
	}
	if !got {
		return nil
	}

	l.flag.Store(true)
	l.gainedOne.Do(func() { close(l.gained) })
	l.mu.Lock()
	l.sinceAt = time.Now()
	l.lastErr = ""
	l.mu.Unlock()
	l.log.Lifecycle("LEADER_ACQUIRED", map[string]any{
		"instance":  l.instance,
		"lock_name": l.cfg.LockName,
	}, WithOperation("acquire"))

	defer l.release(conn)

	tick := time.NewTicker(l.cfg.KeepaliveInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			pingCtx, cancel := context.WithTimeout(ctx, l.cfg.KeepaliveInterval)
			err := conn.PingContext(pingCtx)
			cancel()
			if err != nil {
				return fmt.Errorf("keepalive ping: %w", err)
			}
		}
	}
}

// acquire returns (true, nil) on success, (false, nil) when the lock is held
// elsewhere for the whole LockTimeout, and an error only when the query itself
// failed.
func (l *Leader) acquire(ctx context.Context, conn *sql.Conn) (bool, error) {
	if l.driver == DBDriverMySQL {
		secs := int(l.cfg.LockTimeout / time.Second)
		if secs < 1 {
			secs = 1
		}
		var got sql.NullInt64
		if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", l.cfg.LockName, secs).Scan(&got); err != nil {
			return false, fmt.Errorf("get_lock: %w", err)
		}
		// NULL means the wait was aborted (killed / deadlock), which is rare
		// but not an error worth escalating: another instance holds it.
		return got.Valid && got.Int64 == 1, nil
	}

	deadline := time.Now().Add(l.cfg.LockTimeout)
	for {
		var got bool
		if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", l.key).Scan(&got); err != nil {
			return false, fmt.Errorf("pg_try_advisory_lock: %w", err)
		}
		if got {
			return true, nil
		}
		if time.Now().After(deadline) {
			return false, nil
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(leaderPollInterval):
		}
	}
}

// release is best-effort. If the connection is already broken — usually the
// reason we got here — the statement fails and the server releases the session
// lock on close anyway.
func (l *Leader) release(conn *sql.Conn) {
	releaseCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if l.driver == DBDriverMySQL {
		_, _ = conn.ExecContext(releaseCtx, "SELECT RELEASE_LOCK(?)", l.cfg.LockName)
	} else {
		_, _ = conn.ExecContext(releaseCtx, "SELECT pg_advisory_unlock($1)", l.key)
	}
	cancel()

	l.flag.Store(false)
	l.mu.Lock()
	held := l.sinceAt
	l.sinceAt = time.Time{}
	l.mu.Unlock()
	if held.IsZero() {
		return
	}
	l.log.Lifecycle("LEADER_RELEASED", map[string]any{
		"instance":  l.instance,
		"lock_name": l.cfg.LockName,
		"held_ms":   time.Since(held).Milliseconds(),
	}, WithOperation("release"))
}

func (l *Leader) recordError(err error) {
	l.mu.Lock()
	l.lastErr = err.Error()
	l.mu.Unlock()
}

// LeaderInstanceID identifies this process in LeaderState so an operator can
// tell which pod holds the lock. Hostname (k8s pod name, container ID, machine
// name) plus PID is enough — nobody runs so many instances that this needs a
// UUID.
func LeaderInstanceID() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		h = "unknown"
	}
	return fmt.Sprintf("%s.%d", h, os.Getpid())
}

// leaderLockKey maps a lock name onto the bigint pg_advisory_lock takes.
// FNV-1a is chosen for being dependency-free and stable across releases — the
// key must not change between deploys or two versions would each think they
// hold the lock.
func leaderLockKey(name string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int64(h.Sum64())
}
