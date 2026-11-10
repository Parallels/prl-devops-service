package database

import (
	"fmt"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
)

// Pinger is the reachability surface we depend on — satisfied by *sql.DB.
type Pinger interface {
	Ping() error
}

const (
	DefaultPingAttempts = 3
	DefaultPingDelay    = 500 * time.Millisecond // flat; slow-start DBs are a #420/#421 tuning topic
)

// PingWithRetry verifies a freshly opened handle. Every failed attempt except
// the last is warned (the last failure becomes the returned error, so it is
// never only a log line); final failure wraps ErrDatabaseUnreachable —
// #421 and #420 branch on that sentinel.
func PingWithRetry(ctx basecontext.ApiContext, p Pinger, attempts int, delay time.Duration) error {
	if attempts <= 0 {
		attempts = DefaultPingAttempts
	}
	if delay <= 0 {
		delay = DefaultPingDelay
	}

	var err error
	for i := 1; i <= attempts; i++ {
		if err = p.Ping(); err == nil {
			return nil
		}
		if i < attempts {
			ctx.LogWarnf("database ping attempt %d/%d failed: %v", i, attempts, err)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("%w: %d attempts", ErrDatabaseUnreachable, attempts)
}
