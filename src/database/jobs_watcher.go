package database

// Stale-job watcher BODY (PBI #421 task 15, plan D-8). Mirrors the legacy
// ghost-job goroutine body (data/main.go:117-129): ticker on the shared
// constants.GhostJobCheckIntervalSeconds cadence, cancel channel of the
// same `chan bool` shape, DetectStaleJobs per tick. ZERO goroutines are
// started here — this is the blocking body plus the Stop hook; task 18
// owns `go g.runStaleJobWatcher(ctx)` wiring (and retires the legacy
// goroutine outside legacy mode).

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
)

func (g *GormDatabase) runStaleJobWatcher(ctx basecontext.ApiContext) bool {
	if g.staleCancel == nil {
		g.staleCancel = make(chan bool)
	}
	ticker := time.NewTicker(time.Duration(constants.GhostJobCheckIntervalSeconds) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-g.staleCancel:
			ctx.LogInfof("[Database] Ghost job cleanup stopped")
			return true
		case <-ticker.C:
			g.DetectStaleJobs(ctx)
		}
	}
}

// StopStaleJobWatcher signals the running watcher body to exit; blocking
// send mirrors the legacy cancel idiom. Nil-safe before start.
func (g *GormDatabase) StopStaleJobWatcher() {
	if g.staleCancel != nil {
		g.staleCancel <- true
	}
}
