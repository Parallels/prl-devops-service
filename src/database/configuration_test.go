package database

import (
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"gorm.io/gorm"
)

func configCount(t *testing.T, g *GormDatabase) int64 {
	t.Helper()
	var n int64
	if err := g.gormDB.Table("configurations").Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func schemaCount(t *testing.T, g *GormDatabase) int64 {
	t.Helper()
	var n int64
	if err := g.gormDB.Table("schema_metadata").Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestSQLConfig_SeedOnReadSingleton(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	if n := configCount(t, g); n != 0 {
		t.Fatalf("fresh engine must start with no configuration row, got %d", n)
	}
	cfg, err := g.GetConfiguration(ctx)
	if err != nil || cfg == nil || cfg.ID == "" {
		t.Fatalf("seed-on-read: %v %+v", err, cfg)
	}
	if n := configCount(t, g); n != 1 {
		t.Errorf("seed-on-read must create exactly one row, got %d", n)
	}
	again, _ := g.GetConfiguration(ctx)
	if again.ID != cfg.ID {
		t.Errorf("idempotence broken: %q vs %q", cfg.ID, again.ID)
	}

	// Corruption defence: a hand-inserted second row must not break reads.
	if err := g.gormDB.Exec("INSERT INTO configurations (id) VALUES ('second-row')").Error; err != nil {
		t.Fatal(err)
	}
	first, err := g.GetConfiguration(ctx)
	if err != nil || first.ID != cfg.ID {
		t.Errorf("multi-row defence must take first by rowid: %v %+v", err, first)
	}
}

func TestSQLConfig_InstallIdFlow(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	// Fresh install: SetId is THE first-boot write (bootstrappers do this).
	got, err := g.SetId(ctx, "boot-installed-id")
	if err != nil || got != "boot-installed-id" {
		t.Fatalf("fresh SetId: %v %q", err, got)
	}
	if n := configCount(t, g); n != 1 {
		t.Errorf("want 1 row, got %d", n)
	}
	if id, _ := g.GetId(ctx); id != "boot-installed-id" {
		t.Errorf("GetId drift: %q", id)
	}
	// SeedId on an existing id returns the stored value, changes nothing.
	for i := 0; i < 3; i++ {
		if id, err := g.SeedId(ctx); err != nil || id != "boot-installed-id" {
			t.Fatalf("SeedId #%d: %v %q", i, err, id)
		}
	}
	if n := configCount(t, g); n != 1 {
		t.Errorf("SeedId x3 must keep ONE row, got %d", n)
	}
	// Second SetId: legacy plain error, message-pinned.
	if _, err := g.SetId(ctx, "second-write"); err == nil || err.Error() != "ID already exists" {
		t.Errorf("want plain 'ID already exists', got %v", err)
	}

	// Empty-id row (migrated shape): SeedId fills the SAME row, never adds.
	g2 := newSQLDB(t)
	if err := g2.gormDB.Exec("INSERT INTO configurations (id) VALUES ('')").Error; err != nil {
		t.Fatal(err)
	}
	filled, err := g2.SeedId(ctx)
	if err != nil || filled == "" {
		t.Fatalf("SeedId must fill empty id: %v %q", err, filled)
	}
	if n := configCount(t, g2); n != 1 {
		t.Errorf("fill must not add rows, count %d", n)
	}
	if id, _ := g2.GetId(ctx); id != filled {
		t.Errorf("GetId after fill: %q vs %q", id, filled)
	}

	// SetId("") on a fresh engine creates the empty singleton (legacy quirk).
	g3 := newSQLDB(t)
	if got, err := g3.SetId(ctx, ""); err != nil || got != "" {
		t.Errorf("SetId empty on fresh: %v %q", err, got)
	}
	if n := configCount(t, g3); n != 1 {
		t.Errorf("want the empty singleton row, got %d", n)
	}
}

// TestSQLConfig_ConcurrentSeedId pins the spec's race-closure: racing first
// boots must converge on ONE stored id with ONE row (run under -race).
func TestSQLConfig_ConcurrentSeedId(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	const n = 8
	var wg sync.WaitGroup
	ids := make([]string, n)
	errs := make([]error, n)
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			ids[i], errs[i] = g.SeedId(ctx)
		}(i)
	}
	close(start)
	wg.Wait()
	for i := range errs {
		if errs[i] != nil {
			t.Fatalf("concurrent SeedId[%d]: %v", i, errs[i])
		}
	}
	for i := range ids {
		if ids[i] == "" || ids[i] != ids[0] {
			t.Fatalf("race divergence: %v", ids)
		}
	}
	if c := configCount(t, g); c != 1 {
		t.Errorf("want one row after %d racing seeds, got %d", n, c)
	}
}

func TestSQLSchema_VersionLifecycle(t *testing.T) {
	g := newSQLDB(t)
	ctx := basecontext.NewBaseContext()

	// First-boot path: empty table ⇒ "" with NO error (not NotFound drift).
	v, err := g.GetSchemaVersion(ctx)
	if err != nil || v != "" {
		t.Fatalf("fresh GetSchemaVersion: %v %q", err, v)
	}
	// No validation invented: any string round-trips, empties included.
	for _, want := range []string{"0", "1", "2024-01-01", "custom—anything", "", "again", ""} {
		if err := g.UpdateSchemaVersion(ctx, want); err != nil {
			t.Fatalf("update %q: %v", want, err)
		}
		got, err := g.GetSchemaVersion(ctx)
		if err != nil || got != want {
			t.Fatalf("round-trip: %q -> %q (%v)", want, got, err)
		}
		if n := schemaCount(t, g); n != 1 {
			t.Fatalf("schema_metadata must hold ONE row, got %d", n)
		}
	}
}

func TestSQLConfigSchema_NotConnected(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	cold := NewGormDatabase(ctx, nil, nil)

	if _, err := cold.GetConfiguration(ctx); err != datainterfaces.ErrDatabaseNotConnected {
		t.Errorf("GetConfiguration guard: %v", err)
	}
	if _, err := cold.GetId(ctx); err != datainterfaces.ErrDatabaseNotConnected {
		t.Errorf("GetId guard: %v", err)
	}
	if _, err := cold.SeedId(ctx); err != datainterfaces.ErrDatabaseNotConnected {
		t.Errorf("SeedId guard: %v", err)
	}
	if _, err := cold.SetId(ctx, "x"); err != datainterfaces.ErrDatabaseNotConnected {
		t.Errorf("SetId guard: %v", err)
	}
	if _, err := cold.GetSchemaVersion(ctx); err != datainterfaces.ErrDatabaseNotConnected {
		t.Errorf("GetSchemaVersion guard: %v", err)
	}
	if err := cold.UpdateSchemaVersion(ctx, "x"); err != datainterfaces.ErrDatabaseNotConnected {
		t.Errorf("UpdateSchemaVersion guard: %v", err)
	}

	// The #419 OQ-3 pinned table name must exist and start empty.
	g := newSQLDB(t)
	var row models.DatabaseSchema
	err := g.gormDB.Table("schema_metadata").Take(&row).Error
	if err != gorm.ErrRecordNotFound {
		t.Errorf("schema_metadata table must exist and be empty: %v", err)
	}
}
