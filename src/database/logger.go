package database

import (
	"context"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	gormlogger "gorm.io/gorm/logger"
)

const defaultSlowQueryThreshold = 200 * time.Millisecond

// baseContextLogger bridges GORM's logger chatter onto ApiContext (D-6).
// Mode mapping: Silent → nothing; Error → errors only; Warn → warn+errors
// (incl. slow queries); Info → everything, and per-query SQL additionally
// when config LOG_LEVEL is debug/trace.
type baseContextLogger struct {
	ctx           basecontext.ApiContext
	verbose       bool // LOG_LEVEL debug/trace → per-query statements via LogDebugf
	silent        bool // hard suppression: Silent mode or a DisableLog()'d context
	mode          gormlogger.LogLevel
	slowThreshold time.Duration
}

// NewGormLogger reads LOG_LEVEL at construction (config.Get() is lazy-safe).
func NewGormLogger(ctx basecontext.ApiContext) gormlogger.Interface {
	level := strings.ToLower(strings.TrimSpace(config.Get().GetKey(constants.LOG_LEVEL_ENV_VAR)))
	l := &baseContextLogger{
		ctx:           ctx,
		verbose:       level == "debug" || level == "trace",
		mode:          gormlogger.Warn, // parity with gorm's own default
		slowThreshold: defaultSlowQueryThreshold,
	}
	// A DisableLog()'d context self-discards its own output; mirror that as
	// Silent so nothing we log through it can be surprising in tests.
	if v, ok := ctx.(interface{ Verbose() bool }); ok && !v.Verbose() {
		l.silent = true
	}
	return l
}

// NewGormLoggerSilent is the deterministic opt-out for tests that assert on
// zero output regardless of context internals.
func NewGormLoggerSilent(ctx basecontext.ApiContext) gormlogger.Interface {
	l := NewGormLogger(ctx).(*baseContextLogger)
	l.silent = true
	return l
}

func (l *baseContextLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.mode = level
	if level == gormlogger.Silent {
		clone.silent = true
	}
	return &clone
}

func (l *baseContextLogger) Info(_ context.Context, format string, a ...interface{}) {
	if l.silent || l.mode < gormlogger.Info {
		return
	}
	l.ctx.LogInfof(format, a...)
}

func (l *baseContextLogger) Warn(_ context.Context, format string, a ...interface{}) {
	if l.silent || l.mode < gormlogger.Warn {
		return
	}
	l.ctx.LogWarnf(format, a...)
}

func (l *baseContextLogger) Error(_ context.Context, format string, a ...interface{}) {
	if l.silent {
		return
	}
	l.ctx.LogErrorf(format, a...)
}

// Trace never routes through GORM's context.Context (it is not our
// ApiContext) — output goes to the struct's captured ctx only.
func (l *baseContextLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.silent {
		return
	}

	if err != nil {
		// D-6 chose "err always errors" — including gorm's ErrRecordNotFound
		// passthrough; revisit the filter when #421 shows real noise volume.
		sql, rows := fc()
		l.ctx.LogErrorf("query failed (%d rows): %v — %s", rows, err, sql)
		return
	}

	elapsed := time.Since(begin)
	if elapsed > l.slowThreshold && l.mode >= gormlogger.Warn {
		sql, _ := fc()
		l.ctx.LogWarnf("slow query (%s): %s", elapsed, sql)
		return
	}

	if l.verbose && l.mode >= gormlogger.Info {
		sql, rows := fc()
		l.ctx.LogDebugf("%s rows=%d elapsed=%s", sql, rows, elapsed)
	}
}
