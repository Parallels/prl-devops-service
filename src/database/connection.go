package database

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDatabaseConnection opens the backend selected by cfg, applies engine settings,
// and returns a pool-safe handle. Caller owns Close. ctx is for logging.
//
// #418 scope: both engines implemented; unknown types are rejected outright —
// never a silent fallback to sqlite (task 03's interim sentinel was removed
// when the postgres case landed in task 04).
func NewDatabaseConnection(ctx basecontext.ApiContext, cfg DatabaseConfig) (*gorm.DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	switch strings.ToLower(cfg.Type) {
	case DBTypeSQLite:
		return newSQLiteConnection(ctx, cfg)
	case DBTypePostgres:
		return newPostgresConnection(ctx, cfg)
	default:
		return nil, ErrUnsupportedDatabaseType
	}
}

// Close releases the underlying pool; nil-safe; logs (not errors) on failure —
// there is nothing actionable a caller can do with a close failure at the
// three teardown sites we will have.
func Close(ctx basecontext.ApiContext, db *gorm.DB) {
	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		ctx.LogErrorf("database: failed to acquire sql.DB for close: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		ctx.LogErrorf("database: failed to close database connection: %v", err)
	}
}

// gormConfig is the single construction point for gorm.Config so task 05's
// logger bridge replaces exactly one call site.
// TODO(#418 task 05): replace Silent logger with the ctx-bridged gorm logger.
func gormConfig() *gorm.Config {
	return &gorm.Config{
		Logger:         gormlogger.Default.LogMode(gormlogger.Silent),
		TranslateError: false,
		// NamingStrategy stays default: legacy json tags already define the
		// column names #419 will tag explicitly.
	}
}
