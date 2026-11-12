package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/require"
)

// EngineCapabilities feeds the exclusion ledger in parity/doc.go: scenarios
// t.Skip with the ledger entry id when the engine lacks what they assert.
type EngineCapabilities struct {
	HasNotConnectedSemantics bool // both engines return ErrDatabaseNotConnected cold (#421-15 proof)
	IsFileEngine             bool // legacy persists a JSON file; gorm leg runs on memory
}

// Engine pairs a name with a constructor returning a READY DatabaseService
// (connected) plus its cleanup func.
type Engine struct {
	Name         string
	New          func(t *testing.T) (datainterfaces.DatabaseService, func())
	Capabilities EngineCapabilities
}

// Scenario is one behaviour asserted identically on every engine.
type Scenario struct {
	Name string
	Run  func(t *testing.T, ctx basecontext.ApiContext, db datainterfaces.DatabaseService)
}

// LegacyEngine mirrors the #421 parity discipline (harness.go
// NewLegacyEngine): process-singleton reset via data.ResetForTesting
// (//go:build !release), serialised by mutex, fresh t.TempDir store per
// engine, Disconnect BEFORE TempDir removal (the save goroutine's late
// writes would race the dir delete). So legacy legs ARE isolated per test —
// no shared-store posture; the self-cleaning rule is belt-and-braces.
var legacyMu sync.Mutex

func LegacyEngine() Engine {
	return Engine{
		Name: "legacy",
		New: func(t *testing.T) (datainterfaces.DatabaseService, func()) {
			ctx := WithCleanConfig(t)
			legacyMu.Lock()
			t.Cleanup(legacyMu.Unlock)
			data.ResetForTesting()
			t.Cleanup(data.ResetForTesting)
			db := data.NewJsonDatabase(ctx, filepath.Join(t.TempDir(), "test_db.json"))
			require.NotNil(t, db)
			require.NoError(t, db.Connect(ctx))
			return db, func() { _ = db.Disconnect(ctx) } // best-effort shutdown, #421 order
		},
		Capabilities: EngineCapabilities{HasNotConnectedSemantics: true, IsFileEngine: true},
	}
}

// GormEngine: isolated memory DB through the #418 factory + #419 schema +
// the #421 GormDatabase service.
func GormEngine() Engine {
	return Engine{
		Name: "gorm",
		New: func(t *testing.T) (datainterfaces.DatabaseService, func()) {
			ctx := WithCleanConfig(t)
			gdb := OpenMemoryDB(t)
			MigrateSchema(t, gdb)
			svc := database.NewGormDatabase(ctx, gdb)
			require.NoError(t, svc.Connect(ctx))
			return svc, func() { _ = svc.Disconnect(ctx) }
		},
		Capabilities: EngineCapabilities{HasNotConnectedSemantics: true, IsFileEngine: false},
	}
}

// PostgresEngine — the #425 task-05 integration leg (plan D-6 leg (b)).
// Lives untagged so the integration-tagged parity file can name it; New
// t.Skip-fires when POSTGRES_TEST_DSN is unset, so default runs are
// unaffected. Isolation: each engine instance gets a unique schema
// (search_path=prl_t425_<nanos>) created fresh and DROPped CASCADE on
// cleanup — concurrent CI runs cannot collide. The schema DROP runs
// inside the returned cleanup BEFORE Disconnect so it executes while the
// pool is still open.
func PostgresEngine() Engine {
	return Engine{
		Name: "postgres",
		New: func(t *testing.T) (datainterfaces.DatabaseService, func()) {
			ctx := WithCleanConfig(t)
			dsn := os.Getenv("POSTGRES_TEST_DSN")
			if dsn == "" {
				t.Skip("POSTGRES_TEST_DSN unset — Postgres legs skipped (#425 D-6: CI service container / make test-integration only)")
			}
			schema := fmt.Sprintf("prl_t425_%d", time.Now().UnixNano())
			sep := "?"
			if strings.Contains(dsn, "?") {
				sep = "&"
			}
			gdb, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
				Type: database.DBTypePostgres,
				DSN:  dsn + sep + "search_path=" + schema,
			})
			require.NoErrorf(t, err, "postgres boot failed on %s", schema)
			t.Cleanup(func() { database.Close(ctx, gdb) })
			require.NoError(t, gdb.Exec("CREATE SCHEMA IF NOT EXISTS "+schema).Error)
			require.NoError(t, database.RunAutoMigrate(ctx, gdb))
			svc := database.NewGormDatabase(ctx, gdb)
			require.NoError(t, svc.Connect(ctx))
			return svc, func() {
				require.NoError(t, gdb.Exec("DROP SCHEMA IF EXISTS "+schema+" CASCADE").Error)
				_ = svc.Disconnect(ctx)
			}
		},
		Capabilities: EngineCapabilities{HasNotConnectedSemantics: true, IsFileEngine: false},
	}
}

// RunOnEngines executes every scenario against every engine as
// <scenario>/<engine> subtests.
func RunOnEngines(t *testing.T, scenarios []Scenario, engines ...Engine) {
	t.Helper()
	for _, sc := range scenarios {
		for _, eng := range engines {
			eng := eng
			sc := sc
			t.Run(sc.Name+"/"+eng.Name, func(t *testing.T) {
				_, ctx := RecordingContext()
				svc, cleanup := eng.New(t)
				t.Cleanup(cleanup)
				sc.Run(t, ctx, svc)
			})
		}
	}
}
