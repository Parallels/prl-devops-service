package database

// Lifecycle escape-hatch semantics (plan D-7) — these are NOT legacy
// delegates: SQL mode has no JSON save queue. Connect/Disconnect/IsConnected
// live in gorm_database.go. Task 18 removed the last true reach-through:
// SaveAs is now the SQL-native panic snapshot (VACUUM INTO).

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
)

// Load is a no-op: schema ownership is AutoMigrate + #424, no JSON load.
func (g *GormDatabase) Load(ctx basecontext.ApiContext) error { return nil }

// Filename returns the connection identity (":memory:" verbatim for in-memory
// sqlite) — D-7 escape-hatch table.
func (g *GormDatabase) Filename() string { return g.dbFilename() }

// SaveAs is the panic-snapshot surface (main.go recover path calls it through
// the interface). sqlite: VACUUM INTO writes a consistent snapshot of the
// live database — WAL-safe and transactionally atomic, strictly better than
// copying a live file. Driver constraints mirrored exactly: VACUUM INTO takes
// a literal string, NOT a bind parameter, so the path is single-quoted with
// SQL-quote doubling after filepath.Abs (VACUUM INTO requires absolute
// paths — D-7 row; the relative "data.json.<ts>.panic" from main.go resolves
// against CWD, same semantics the legacy snapshot had).
// postgres: online file snapshots have no analogue — log-once warn + no-op
// (documented gap → #427 backup story, OQ-421-3).
func (g *GormDatabase) SaveAs(ctx basecontext.ApiContext, filename string) error {
	if g.gormDB == nil {
		return errors.NewWithCode("database operation failed: no gorm connection", 500)
	}
	switch g.dialect() {
	case "postgres":
		if !g.warnedSnapshot.Swap(true) {
			ctx.LogWarnf("[Database] SQL mode: SaveAs panic snapshots are not supported on PostgreSQL; use native pg tooling (#427)")
		}
		return nil
	case "sqlite":
		abs, err := filepath.Abs(filename)
		if err != nil {
			return errors.NewWithCode(fmt.Sprintf("database operation failed: %v", err), 500)
		}
		quoted := "'" + strings.ReplaceAll(abs, "'", "''") + "'"
		return g.gormDB.Exec("VACUUM INTO " + quoted).Error
	default:
		return errors.NewWithCode("database operation failed: SaveAs unsupported dialect", 500)
	}
}

// SaveAsync/SaveNow/SaveNowForced: SQL writes are the transaction itself;
// the JSON save queue does not exist in SQL mode (D-7). SaveNowForced joined
// the surface after the D-7 table was drafted (upstream PR #557) — same
// no-op class as SaveNow.
func (g *GormDatabase) SaveAsync(ctx basecontext.ApiContext) error { return nil }
func (g *GormDatabase) SaveNow(ctx basecontext.ApiContext) error   { return nil }
func (g *GormDatabase) SaveNowForced(ctx basecontext.ApiContext, reason string) error {
	return nil
}

// ProcessSaveQueue: no save queue, nothing to pump.
func (g *GormDatabase) ProcessSaveQueue(ctx basecontext.ApiContext) {}

// IsDataFileEmpty speaks about a JSON data file that does not exist in SQL
// mode — always "not empty", no error (D-7).
func (g *GormDatabase) IsDataFileEmpty(ctx basecontext.ApiContext) (bool, error) {
	return false, nil
}

// Backup/RunBackup: file backups are replaced by SQL backup stories (#427).
// Warn ONCE so long-lived processes do not spam the log.
func (g *GormDatabase) RunBackup(ctx basecontext.ApiContext) {
	if !g.warnedBackup.Swap(true) {
		ctx.LogWarnf("[Database] SQL mode: file backups are disabled; use database-native backup tooling")
	}
}

func (g *GormDatabase) Backup(ctx basecontext.ApiContext) error {
	if !g.warnedBackup.Swap(true) {
		ctx.LogWarnf("[Database] SQL mode: file backups are disabled; use database-native backup tooling")
	}
	return nil
}
