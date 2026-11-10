package database

// Unit tests for the GormDatabase hybrid bootstrap (task 04).
//
// IMPORTANT — no legacy engine in this binary: data.JsonDatabase keeps a
// background save goroutine that logs through the shared logger/config
// globals, which would data-race with the config-mutating logger/config
// tests in this package. Anything that needs a live legacy handle lives in
// database/parity instead (see parity/hybrid_delegates_test.go), the package
// whose sole job is hosting legacy engines under -race.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"gorm.io/gorm"
)

// R-1: the SQL side must satisfy the frozen 158-method surface.
var _ datainterfaces.DatabaseService = (*GormDatabase)(nil)

func newMemoryGorm(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := NewDatabaseConnection(basecontext.NewBaseContext(), DatabaseConfig{Type: "sqlite", Path: ":memory:"})
	if err != nil {
		t.Fatalf("sqlite :memory: connection: %v", err)
	}
	return conn
}

func TestNewGormDatabase_ImplementsInterface(t *testing.T) {
	conn := newMemoryGorm(t)
	g := NewGormDatabase(basecontext.NewBaseContext(), conn, nil)
	if err := g.Connect(basecontext.NewBaseContext()); err != nil {
		t.Fatalf("Connect on :memory:: %v", err)
	}
	if !g.IsConnected() {
		t.Fatal("IsConnected false after Connect")
	}
	if g.Filename() != ":memory:" {
		t.Errorf("Filename: want :memory:, got %q", g.Filename())
	}
}

func TestLifecycleSemantics(t *testing.T) {
	g := NewGormDatabase(basecontext.NewBaseContext(), newMemoryGorm(t), nil)
	ctx := basecontext.NewBaseContext()

	if err := g.SaveNow(ctx); err != nil {
		t.Errorf("SaveNow: want nil, got %v", err)
	}
	if err := g.SaveAsync(ctx); err != nil {
		t.Errorf("SaveAsync: want nil, got %v", err)
	}
	if err := g.SaveNowForced(ctx, "test"); err != nil {
		t.Errorf("SaveNowForced: want nil, got %v", err)
	}
	if err := g.Load(ctx); err != nil {
		t.Errorf("Load: want nil, got %v", err)
	}
	g.ProcessSaveQueue(ctx) // must not panic
	empty, err := g.IsDataFileEmpty(ctx)
	if empty || err != nil {
		t.Errorf("IsDataFileEmpty: want false/nil, got %v/%v", empty, err)
	}
	if err := g.Backup(ctx); err != nil {
		t.Errorf("Backup: want nil, got %v", err)
	}
	g.RunBackup(ctx)
	if !g.warnedBackup.Load() {
		t.Error("warn-once flag not set after Backup/RunBackup")
	}
}

func TestClose_Idempotent(t *testing.T) {
	g := NewGormDatabase(basecontext.NewBaseContext(), newMemoryGorm(t), nil)
	ctx := basecontext.NewBaseContext()
	if err := g.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	if err := g.Disconnect(ctx); err != nil {
		t.Fatalf("first Disconnect: %v", err)
	}
	if err := g.Disconnect(ctx); err != nil {
		t.Fatalf("second Disconnect (closeOnce): %v", err)
	}
	if g.IsConnected() {
		t.Error("IsConnected true after Disconnect")
	}
}
