package testutil

import (
	"path/filepath"
	"sync"
	"testing"

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
