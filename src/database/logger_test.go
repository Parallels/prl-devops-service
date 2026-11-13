package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
	gormlogger "gorm.io/gorm/logger"
)

func sqlFC(sql string, rows int64) func() (string, int64) {
	return func() (string, int64) { return sql, rows }
}

func TestGormLogger_SilentSuppressesEverything(t *testing.T) {
	resetConfig(t)
	rec := &recordingContext{}
	l := NewGormLoggerSilent(rec)

	l.Trace(context.Background(), time.Now(), sqlFC("SELECT 1", 1), errors.New("boom"))
	l.Info(context.Background(), "info %d", 1)
	l.Warn(context.Background(), "warn")
	l.Error(context.Background(), "error")
	l.LogMode(gormlogger.Info).Trace(context.Background(), time.Now().Add(-time.Second), sqlFC("slow", 0), nil)

	assert.Empty(t, rec.errs)
	assert.Empty(t, rec.infos)
	assert.Empty(t, rec.warns)
	assert.Empty(t, rec.debugs)
}

func TestGormLogger_DisabledRealContextIsSilent(t *testing.T) {
	resetConfig(t)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	// No assertion on internal state possible for a disabled BaseContext —
	// the observable contract is "no panic and no output", output being
	// unobservable is the point; Verbose() mirrors the same flag we key off.
	l := NewGormLogger(ctx)
	assert.NotPanics(t, func() {
		l.Trace(context.Background(), time.Now(), sqlFC("SELECT 1", 1), errors.New("boom"))
	})
}

func TestGormLogger_ErrorModeOnlyErrors(t *testing.T) {
	resetConfig(t)
	rec := &recordingContext{}
	l := NewGormLogger(rec).LogMode(gormlogger.Error)

	l.Error(context.Background(), "plain error")
	l.Warn(context.Background(), "plain warn")
	l.Info(context.Background(), "plain info")
	l.Trace(context.Background(), time.Now().Add(-time.Second), sqlFC("SELECT slow", 7), nil) // slow
	l.Trace(context.Background(), time.Now(), sqlFC("SELECT bad", 0), errors.New("record not found"))

	assert.Equal(t, "plain error", rec.errs[0])
	assert.Contains(t, rec.errs[1], "record not found")
	assert.Empty(t, rec.warns)
	assert.Empty(t, rec.infos)
	assert.Empty(t, rec.debugs)
}

func TestGormLogger_WarnModeSlowQuery(t *testing.T) {
	resetConfig(t)
	rec := &recordingContext{}
	l := NewGormLogger(rec) // default mode Warn

	l.Trace(context.Background(), time.Now().Add(-300*time.Millisecond), sqlFC("SELECT slow", 3), nil)
	l.Trace(context.Background(), time.Now(), sqlFC("SELECT fast", 3), nil)

	assert.Len(t, rec.warns, 1)
	assert.Contains(t, rec.warns[0], "slow query")
	assert.Contains(t, rec.warns[0], "SELECT slow")
	assert.Empty(t, rec.debugs) // LOG_LEVEL not debug
}

func TestGormLogger_InfoModeVerboseDebug(t *testing.T) {
	t.Setenv(constants.LOG_LEVEL_ENV_VAR, "DEBUG")
	resetConfig(t)
	rec := &recordingContext{}
	l := NewGormLogger(rec).LogMode(gormlogger.Info)

	l.Trace(context.Background(), time.Now(), sqlFC("SELECT every", 5), nil)
	assert.Len(t, rec.debugs, 1)
	assert.Contains(t, rec.debugs[0], "SELECT every")
	assert.Contains(t, rec.debugs[0], "rows=5")
}

func TestGormLogger_InfoLevelNoPerQueryDebug(t *testing.T) {
	t.Setenv(constants.LOG_LEVEL_ENV_VAR, "INFO")
	resetConfig(t)
	rec := &recordingContext{}
	l := NewGormLogger(rec).LogMode(gormlogger.Info)

	l.Trace(context.Background(), time.Now(), sqlFC("SELECT quiet", 5), nil)
	assert.Empty(t, rec.debugs) // LOG_LEVEL=info: no per-query statements even at gorm Info mode
}
