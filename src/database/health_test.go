package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/stretchr/testify/assert"
)

// recordingContext captures ApiContext log output for assertions. Embedding
// the interface keeps the fake small — anything beyond the Log* surface is
// never called on these paths (a stray call panics loudly, which is a feature).
type recordingContext struct {
	basecontext.ApiContext
	infos  []string
	warns  []string
	errs   []string
	debugs []string
}

func (r *recordingContext) LogInfof(f string, a ...interface{}) {
	r.infos = append(r.infos, fmt.Sprintf(f, a...))
}
func (r *recordingContext) LogWarnf(f string, a ...interface{}) {
	r.warns = append(r.warns, fmt.Sprintf(f, a...))
}
func (r *recordingContext) LogErrorf(f string, a ...interface{}) {
	r.errs = append(r.errs, fmt.Sprintf(f, a...))
}
func (r *recordingContext) LogDebugf(f string, a ...interface{}) {
	r.debugs = append(r.debugs, fmt.Sprintf(f, a...))
}
func (r *recordingContext) DisableLog() {}
func (r *recordingContext) EnableLog()  {}

// Verbose is part of ApiContext (NewGormLogger keys log-suppression off it);
// the fake reports "logging on" and asserts on captured lines instead.
func (r *recordingContext) Verbose() bool { return true }

type fakePinger struct {
	failUntil int
	calls     int
}

func (f *fakePinger) Ping() error {
	f.calls++
	if f.calls <= f.failUntil {
		return fmt.Errorf("dial refused (attempt %d)", f.calls)
	}
	return nil
}

func TestPingWithRetry_SucceedsAfterKFailures(t *testing.T) {
	rec := &recordingContext{}
	p := &fakePinger{failUntil: 1}

	start := time.Now()
	err := PingWithRetry(rec, p, 3, 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, 2, p.calls)
	assert.Len(t, rec.warns, 1)
	assert.Contains(t, rec.warns[0], "database ping attempt 1/3 failed")
	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond)
}

func TestPingWithRetry_ExhaustsAndWrapsSentinel(t *testing.T) {
	rec := &recordingContext{}
	p := &fakePinger{failUntil: 1 << 30}

	delay := 40 * time.Millisecond
	start := time.Now()
	err := PingWithRetry(rec, p, 3, delay)
	assert.ErrorIs(t, err, ErrDatabaseUnreachable)
	assert.Contains(t, err.Error(), "3 attempts")
	assert.Equal(t, 3, p.calls, "exactly `attempts` pings, no more")
	assert.Len(t, rec.warns, 2, "every failure but the last is warned")
	assert.GreaterOrEqual(t, time.Since(start), 2*delay, "sleeps between attempts")
}

func TestPingWithRetry_ZeroValuesUseDefaults(t *testing.T) {
	rec := &recordingContext{}
	p := &fakePinger{}
	assert.NoError(t, PingWithRetry(rec, p, 0, 0)) // defaults: 1 attempt needed, 1 call
	assert.Equal(t, 1, p.calls)
}
