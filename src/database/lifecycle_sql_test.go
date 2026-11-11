package database

// SQL lifecycle semantics (PBI #421 task 18): the panic-snapshot surface
// (SaveAs) and the D-8 watcher launcher, plus the production AutoMigrate
// registry gate. These assert the SQL-native behaviour that REPLACED the
// legacy delegates — legacy comparisons retired with the handle (the parity
// harness keeps the zero-reach-through pin).

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	postgresdrv "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSaveAsVacuumsSnapshot(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	conn, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: "sqlite", Path: MemoryPath})
	if err != nil {
		t.Fatal(err)
	}
	if err := RunAutoMigrate(ctx, conn); err != nil {
		t.Fatal(err)
	}
	g := NewGormDatabase(ctx, conn)

	// Row that must survive the vacuum (string-date columns per D-1 models).
	if err := conn.Exec("INSERT INTO users (id, username, name, created_at) VALUES ('snap-1', 'snapuser', 'Snap', '2026-06-01T12:00:00Z')").Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	snap := filepath.Join(t.TempDir(), "panic-snapshot.db")
	if err := g.SaveAs(ctx, snap); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	if _, err := os.Stat(snap); err != nil {
		t.Fatalf("snapshot file: %v", err)
	}

	// The snapshot is a STANDALONE database: open it fresh and see the
	// migrated schema AND the row (VACUUM carries data atomically).
	snapDB, openErr := NewDatabaseConnection(ctx, DatabaseConfig{Type: "sqlite", Path: snap})
	if openErr != nil {
		t.Fatalf("open snapshot: %v", openErr)
	}
	var n int
	if err := snapDB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'users'").Scan(&n).Error; err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if n != 1 {
		t.Fatalf("users table missing from snapshot (count %d)", n)
	}
	var rows int
	if err := snapDB.Raw("SELECT COUNT(*) FROM users WHERE username = 'snapuser'").Scan(&rows).Error; err != nil {
		t.Fatalf("snapshot row read: %v", err)
	}
	if rows != 1 {
		t.Fatalf("seeded row missing from snapshot (count %d)", rows)
	}
}

func TestSaveAsRelativePathResolvesToCwd(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()
	dir := t.TempDir()
	t.Chdir(dir) // VACUUM INTO needs absolute; main.go passes a RELATIVE panic name

	if err := g.SaveAs(ctx, "data.json.20260601.panic"); err != nil {
		t.Fatalf("SaveAs relative: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "data.json.20260601.panic")); err != nil {
		t.Fatalf("relative snapshot must land in cwd: %v", err)
	}
}

func TestSaveAsQuotesOddPaths(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	// VACUUM INTO takes a literal, not a bind param — the quote doubling in
	// SaveAs is what keeps this path injectable-safe and working.
	snap := filepath.Join(t.TempDir(), "we''ird; drop.db")
	if err := g.SaveAs(ctx, snap); err != nil {
		t.Fatalf("SaveAs quoted path: %v", err)
	}
	if _, err := os.Stat(snap); err != nil {
		t.Fatalf("quoted-name snapshot: %v", err)
	}
}

func TestSaveAsPostgresWarnsNoOp(t *testing.T) {
	// No connection needed: the pg branch returns before touching the pool
	// (OQ-421-3: pg snapshots = native tooling, documented gap in #427).
	g := NewGormDatabase(basecontext.NewBaseContext(), &gorm.DB{Config: &gorm.Config{Dialector: &postgresdrv.Dialector{}}})
	ctx := basecontext.NewBaseContext()

	for i := 0; i < 3; i++ { // log-once must stay crash-free across calls
		if err := g.SaveAs(ctx, "/tmp/whatever.db"); err != nil {
			t.Fatalf("pg SaveAs call %d: %v", i, err)
		}
	}
}

func TestSaveAsWithoutConnection(t *testing.T) {
	g := NewGormDatabase(basecontext.NewBaseContext(), nil)
	if err := g.SaveAs(basecontext.NewBaseContext(), "x.db"); err == nil {
		t.Fatal("cold engine must refuse SaveAs, not fake success")
	}
}

func TestSaveAsUnknownDialectRefuses(t *testing.T) {
	g := NewGormDatabase(basecontext.NewBaseContext(), &gorm.DB{Config: &gorm.Config{}})
	if err := g.SaveAs(basecontext.NewBaseContext(), "x.db"); err == nil {
		t.Fatal("unknown dialect must error, not silently no-op")
	}
}

func TestStartStaleJobWatcherRendezvous(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	g.StartStaleJobWatcher(ctx) // channel initialised SYNCHRONOUSLY (task-15 race lesson)
	done := make(chan bool)
	go func() { g.StopStaleJobWatcher(); close(done) }()
	select {
	case <-done: // the watcher body consumed the cancel
	case <-time.After(5 * time.Second):
		t.Fatal("Stop never reached the started watcher (D-8 launcher broken)")
	}
}

func TestRunAutoMigrateRegistry(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	conn, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: "sqlite", Path: MemoryPath})
	if err != nil {
		t.Fatal(err)
	}
	if err := RunAutoMigrate(ctx, conn); err != nil {
		t.Fatalf("RunAutoMigrate: %v", err)
	}
	var n int
	if err := conn.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table'").Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if got := len(AllGormModels()); n < got {
		t.Fatalf("registry says %d models, sqlite_master has %d tables", got, n)
	}
	if err := RunAutoMigrate(ctx, nil); err != ErrDatabaseUnreachable {
		t.Fatalf("nil conn: %v", err)
	}
	// Second run must be idempotent (boot restarts re-enter this).
	if err := RunAutoMigrate(ctx, conn); err != nil {
		t.Fatalf("second RunAutoMigrate: %v", err)
	}
}
