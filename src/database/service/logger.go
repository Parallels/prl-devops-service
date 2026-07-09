package service

import (
	"context"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"gorm.io/gorm/logger"
)

// BaseContextLogger is a GORM logger that bridges to basecontext.ApiContext
type BaseContextLogger struct {
	ctx                  basecontext.ApiContext
	SlowThreshold        time.Duration
	IgnoreRecordNotFound bool
	LogLevel             logger.LogLevel
}

// NewBaseContextLogger creates a new GORM logger that writes to basecontext
func NewBaseContextLogger(ctx basecontext.ApiContext, level logger.LogLevel) *BaseContextLogger {
	return &BaseContextLogger{
		ctx:                  ctx,
		SlowThreshold:        200 * time.Millisecond,
		IgnoreRecordNotFound: true,
		LogLevel:             level,
	}
}

// LogMode sets the log level
func (l *BaseContextLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info logs info messages
func (l *BaseContextLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.ctx.LogInfof(msg, data...)
	}
}

// Warn logs warning messages
func (l *BaseContextLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.ctx.LogWarnf(msg, data...)
	}
}

// Error logs error messages
func (l *BaseContextLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.ctx.LogErrorf(msg, data...)
	}
}

// Trace logs SQL queries and execution time
func (l *BaseContextLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.LogLevel >= logger.Error:
		l.ctx.LogErrorf("[%.3fms] [rows:%d] %s | Error: %v", float64(elapsed.Nanoseconds())/1e6, rows, sql, err)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		l.ctx.LogWarnf("[SLOW SQL %.3fms] [rows:%d] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	case l.LogLevel >= logger.Info:
		l.ctx.LogDebugf("[%.3fms] [rows:%d] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}

// ConvertLogLevel converts project log level to GORM log level
func ConvertLogLevel(isDebug bool) logger.LogLevel {
	if isDebug {
		return logger.Info // Show SQL queries in debug mode
	}
	return logger.Silent // Silent in production
}
