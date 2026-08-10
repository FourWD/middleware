package infra

import (
	"strings"
	"testing"
	"time"
)

// The key must not drift between releases: two versions of one service that
// hashed a name differently would each believe they hold the lock.
func TestLeaderLockKeyIsStable(t *testing.T) {
	const name = "gps_db_gateway:leader:v1"
	if got, want := leaderLockKey(name), int64(-728858297408799779); got != want {
		t.Fatalf("leaderLockKey(%q) = %d, want %d — changing the hash splits leadership across a rolling deploy", name, got, want)
	}
	if leaderLockKey("a") == leaderLockKey("b") {
		t.Fatal("distinct names hashed to the same key")
	}
}

func TestLeaderConfigNormalized(t *testing.T) {
	got := LeaderConfig{}.normalized()
	if got.LockTimeout != defaultLeaderLockTimeout {
		t.Errorf("LockTimeout = %v, want %v", got.LockTimeout, defaultLeaderLockTimeout)
	}
	if got.RetryInterval != defaultLeaderRetryInterval {
		t.Errorf("RetryInterval = %v, want %v", got.RetryInterval, defaultLeaderRetryInterval)
	}
	if got.KeepaliveInterval != defaultLeaderKeepaliveInterval {
		t.Errorf("KeepaliveInterval = %v, want %v", got.KeepaliveInterval, defaultLeaderKeepaliveInterval)
	}

	explicit := LeaderConfig{LockTimeout: time.Minute}.normalized()
	if explicit.LockTimeout != time.Minute {
		t.Errorf("explicit LockTimeout overwritten: %v", explicit.LockTimeout)
	}
}

func TestLeaderResolveDSN(t *testing.T) {
	t.Run("falls back to the primary database driver", func(t *testing.T) {
		t.Setenv("DB_DRIVER", DBDriverPostgres)
		t.Setenv("DB_NAME", "gps_gateway")
		driver, dsn, err := LeaderConfig{}.resolveDSN()
		if err != nil {
			t.Fatal(err)
		}
		if driver != "pgx" {
			t.Errorf("driver = %q, want pgx", driver)
		}
		if !strings.Contains(dsn, "dbname=gps_gateway") {
			t.Errorf("dsn %q does not carry the primary database name", dsn)
		}
	})

	t.Run("mysql maps to its own driver name", func(t *testing.T) {
		t.Setenv("DB_DRIVER", DBDriverMySQL)
		driver, _, err := LeaderConfig{}.resolveDSN()
		if err != nil {
			t.Fatal(err)
		}
		if driver != DBDriverMySQL {
			t.Errorf("driver = %q, want %q", driver, DBDriverMySQL)
		}
	})

	t.Run("explicit DSN wins over the primary database", func(t *testing.T) {
		t.Setenv("DB_DRIVER", DBDriverMySQL)
		driver, dsn, err := LeaderConfig{Driver: DBDriverPostgres, DSN: "host=lock-host dbname=locks"}.resolveDSN()
		if err != nil {
			t.Fatal(err)
		}
		if driver != "pgx" || dsn != "host=lock-host dbname=locks" {
			t.Errorf("got (%q, %q), want (pgx, the explicit DSN)", driver, dsn)
		}
	})

	t.Run("unknown driver is rejected", func(t *testing.T) {
		if _, _, err := (LeaderConfig{Driver: "oracle"}).resolveDSN(); err == nil {
			t.Fatal("want an error for an unsupported driver")
		}
	})
}

// A disabled Leader must be usable without a database — that is the whole
// point of the one-instance path.
func TestNewLeaderDisabled(t *testing.T) {
	l, err := NewLeader(LeaderConfig{Enabled: false}, AppLog)
	if err != nil {
		t.Fatal(err)
	}
	if !l.IsLeader() {
		t.Error("a disabled Leader must report IsLeader() == true")
	}
	select {
	case <-l.Gained():
	default:
		t.Error("a disabled Leader must have Gained() already closed")
	}
	if st := l.State(); st.Enabled {
		t.Error("State().Enabled must be false when the lock is disabled")
	}
}

func TestNewLeaderRejectsEmptyLockName(t *testing.T) {
	if _, err := NewLeader(LeaderConfig{Enabled: true}, AppLog); err == nil {
		t.Fatal("want an error: an empty lock name would let every instance win")
	}
}

func TestLoadLeaderConfigDefaultsLockNameToAppID(t *testing.T) {
	t.Setenv("APP_ID", "gps-db-gateway")
	t.Setenv("LEADER_LOCK_NAME", "")
	if got, want := LoadLeaderConfig().LockName, "gps-db-gateway:leader:v1"; got != want {
		t.Errorf("LockName = %q, want %q", got, want)
	}

	t.Setenv("LEADER_LOCK_NAME", "explicit")
	if got := LoadLeaderConfig().LockName; got != "explicit" {
		t.Errorf("LockName = %q, want the explicit value", got)
	}
}
