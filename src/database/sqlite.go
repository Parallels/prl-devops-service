package database

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// No pool caps for SQLite by decision (D-4): the pure-Go driver serialises
// writes below and WAL + busy_timeout(10s) absorb read concurrency; capping
// database/sql would only add queueing on top of the engine's own lock.

func newSQLiteConnection(ctx basecontext.ApiContext, cfg DatabaseConfig) (*gorm.DB, error) {
	path := cfg.Path

	if path != MemoryPath {
		abs, err := filepath.Abs(path) // documented: relative paths resolve against CWD
		if err != nil {
			return nil, errors.NewWithCode(fmt.Sprintf("failed to resolve sqlite path %s: %v", redactPath(path), err), 500)
		}
		path = abs

		if err := helpers.CreateDirIfNotExist(filepath.Dir(path)); err != nil {
			return nil, errors.NewWithCode(fmt.Sprintf("failed to create sqlite directory %s: %v", redactPath(filepath.Dir(path)), err), 500)
		}
	}

	db, err := gorm.Open(sqlite.Open(buildSQLiteDSN(path)), gormConfig(ctx))
	if err != nil {
		return nil, errors.NewWithCode(fmt.Sprintf("failed to open sqlite database at %s: %v", redactPath(path), err), 500)
	}

	// D-5 boot verification — uniform with postgres (task 05; sqlite previously
	// had no explicit ping, the pragma pass below is the deeper proof).
	sqlDB, err := db.DB()
	if err != nil {
		Close(ctx, db)
		return nil, errors.NewFromError(err)
	}
	// Single connection, always (task 08 finding): database/sql pooling on a
	// :memory: DSN FORKS THE DATABASE — every pooled connection is its own
	// private DB, silently losing the singleton tables (configurations,
	// schema_metadata) and any in-flight transaction pairing. On file DBs it
	// additionally matches the legacy engine's single-writer posture and
	// removes SQLITE_BUSY contention. This is a hard correctness pin, not a
	// tunable; postgres keeps its real pool.
	sqlDB.SetMaxOpenConns(1)
	if err := PingWithRetry(ctx, sqlDB, DefaultPingAttempts, DefaultPingDelay); err != nil {
		Close(ctx, db)
		return nil, err
	}

	// Private in-memory databases support neither WAL nor meaningful persistence
	// checks; file DBs must boot production-real or not at all.
	if path != MemoryPath {
		if err := verifySQLitePragmas(ctx, db); err != nil {
			Close(ctx, db) // no leaked -wal/-shm holders
			return nil, err
		}
	}

	return db, nil
}

// buildSQLiteDSN quotes the path via url.URL Opaque so spaces (macOS reality:
// /Users/Carlos Lapao/... class paths happen) survive, then appends pragmas.
func buildSQLiteDSN(path string) string {
	if path == MemoryPath {
		return ":memory:" // glebarez expects exactly this for private memory DBs
	}

	u := url.URL{Scheme: "file", Opaque: path}
	return u.String() + "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"
}

// verifySQLitePragmas fails fast at boot (rather than at first write) when the
// filesystem cannot honour WAL — e.g. some network mounts — naming exactly
// which pragma mismatched.
func verifySQLitePragmas(ctx basecontext.ApiContext, db *gorm.DB) error {
	var journalMode string
	if err := db.Raw("PRAGMA journal_mode").Scan(&journalMode).Error; err != nil {
		return errors.NewWithCode(fmt.Sprintf("failed to read sqlite journal_mode pragma: %v", err), 500)
	}
	if !strings.EqualFold(journalMode, "wal") {
		return errors.NewWithCode(fmt.Sprintf("sqlite journal_mode = %q, expected \"wal\" (filesystem may not support WAL)", journalMode), 500)
	}

	var foreignKeys int64
	if err := db.Raw("PRAGMA foreign_keys").Scan(&foreignKeys).Error; err != nil {
		return errors.NewWithCode(fmt.Sprintf("failed to read sqlite foreign_keys pragma: %v", err), 500)
	}
	if foreignKeys != 1 {
		return errors.NewWithCode(fmt.Sprintf("sqlite foreign_keys = %d, expected 1", foreignKeys), 500)
	}

	return nil
}

// redactPath is a placeholder for D-8 redaction symmetry. SQLite paths are not
// credential-bearing (task 04's redactDSN covers the DSN-bearing postgres
// side), so this is a plain passthrough today — kept so the error-format call
// sites above never need to change when that assumption is revisited.
func redactPath(path string) string {
	return path
}
