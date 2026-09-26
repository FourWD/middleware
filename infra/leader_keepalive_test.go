package infra

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// leaderFakeDriver answers GET_LOCK with 1 and IS_USED_LOCK with the value in
// owned, enough to drive the MySQL keepalive path without a server.
type leaderFakeDriver struct{ owned *atomic.Int64 }

func (d leaderFakeDriver) Open(string) (driver.Conn, error) { return leaderFakeConn(d), nil }

type leaderFakeConn struct{ owned *atomic.Int64 }

func (c leaderFakeConn) Prepare(query string) (driver.Stmt, error) {
	return leaderFakeStmt{query: query, owned: c.owned}, nil
}
func (leaderFakeConn) Close() error              { return nil }
func (leaderFakeConn) Begin() (driver.Tx, error) { return nil, driver.ErrSkip }

type leaderFakeStmt struct {
	query string
	owned *atomic.Int64
}

func (leaderFakeStmt) Close() error  { return nil }
func (leaderFakeStmt) NumInput() int { return -1 }
func (leaderFakeStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}
func (s leaderFakeStmt) Query([]driver.Value) (driver.Rows, error) {
	v := int64(1)
	if strings.Contains(s.query, "IS_USED_LOCK") {
		v = s.owned.Load()
	}
	return &leaderFakeRows{v: v}, nil
}

type leaderFakeRows struct {
	v    int64
	done bool
}

func (*leaderFakeRows) Columns() []string { return []string{"v"} }
func (*leaderFakeRows) Close() error      { return nil }
func (r *leaderFakeRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.v
	return nil
}

var leaderFakeOwned atomic.Int64

func init() {
	sql.Register("leaderfake", leaderFakeDriver{owned: &leaderFakeOwned})
}

func TestLeaderKeepaliveDropsFlagWhenLockLost(t *testing.T) {
	l, err := NewLeader(LeaderConfig{
		Enabled:           true,
		LockName:          "test:leader",
		Driver:            DBDriverMySQL,
		DSN:               "u:p@tcp(127.0.0.1:1)/x",
		KeepaliveInterval: 20 * time.Millisecond,
	}, NewSlogAPILogger(io.Discard, SlogAPILoggerConfig{LogLevel: "debug"}))
	if err != nil {
		t.Fatal(err)
	}
	_ = l.db.Close()
	l.db, err = sql.Open("leaderfake", "")
	if err != nil {
		t.Fatal(err)
	}
	defer l.db.Close()

	leaderFakeOwned.Store(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- l.acquireAndHold(ctx) }()

	require.Eventually(t, l.IsLeader, time.Second, time.Millisecond, "leadership never acquired")
	time.Sleep(60 * time.Millisecond)
	if !l.IsLeader() {
		t.Fatal("leadership dropped while the lock was still owned")
	}

	leaderFakeOwned.Store(0)
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "keepalive verify") {
			t.Fatalf("want a keepalive verify error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("acquireAndHold did not return after the lock was lost")
	}
	if l.IsLeader() {
		t.Fatal("IsLeader() still true after the lock was lost")
	}
}

// A holder must notice a lost lock before a polling standby can take it over
// with the default settings.
func TestLeaderDefaultKeepaliveShorterThanFailover(t *testing.T) {
	stale := defaultLeaderKeepaliveInterval + leaderVerifyTimeout(defaultLeaderKeepaliveInterval)
	if failover := defaultLeaderLockTimeout + defaultLeaderRetryInterval; stale >= failover {
		t.Fatalf("stale window %v not below standby failover %v", stale, failover)
	}
}
