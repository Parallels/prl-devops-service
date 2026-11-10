package database

// Lifecycle escape-hatch semantics (plan D-7) — these are NOT legacy
// delegates: SQL mode has no JSON save queue. Connect/Disconnect/IsConnected
// live in gorm_database.go.

import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

// Load is a no-op: schema ownership is AutoMigrate + #424, no JSON load.
func (g *GormDatabase) Load(ctx basecontext.ApiContext) error { return nil }

// Filename returns the connection identity (":memory:" verbatim for in-memory
// sqlite) — D-7 escape-hatch table.
func (g *GormDatabase) Filename() string { return g.dbFilename() }

// SaveAs keeps delegating to the legacy engine until the flip-day SQL
// snapshot lands: TODO(#421 task 18): VACUUM INTO.
func (g *GormDatabase) SaveAs(ctx basecontext.ApiContext, filename string) error {
	return g.legacy.SaveAs(ctx, filename)
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
