package parity

// Differential parity harness (PBI #421 task 03, plan D-5).

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
)

// DomainSuite is the contract every domain task 05-16 plugs into: Seed uses
// ONLY exported DatabaseService API surface; Checks receive both engines and
// assert equivalent observable behaviour.
type DomainSuite struct {
	Name   string
	Seed   func(t testing.TB, db datainterfaces.DatabaseService)
	Checks []func(t testing.TB, legacy datainterfaces.DatabaseService, sql datainterfaces.DatabaseService)
}

// legacyMu serialises the legacy singleton factory (see doc.go).
var legacyMu sync.Mutex

// sqlFactory is registered by task 04 (GormDatabase bootstrap). Until then
// NewSQLEngine falls back to the legacy engine so the harness itself stays
// testable (self-proof runs legacy-vs-legacy + a drift-catch assertion).
var sqlFactory func(t testing.TB) datainterfaces.DatabaseService

// SetSQLEngineFactory is called from task 04 onward to plug the real SQL
// engine into every DomainSuite.
func SetSQLEngineFactory(f func(t testing.TB) datainterfaces.DatabaseService) {
	sqlFactory = f
}

// NewLegacyEngine spins up an isolated JsonDatabase in a t.TempDir() and
// serialises it against other legacy engines for the duration of the test.
func NewLegacyEngine(t testing.TB) *data.JsonDatabase {
	t.Helper()
	legacyMu.Lock()
	t.Cleanup(legacyMu.Unlock)
	data.ResetForTesting()
	t.Cleanup(data.ResetForTesting)
	db := data.NewJsonDatabase(basecontext.NewBaseContext(), filepath.Join(t.TempDir(), "data.json"))
	t.Cleanup(func() {
		// Stop the background save process before TempDir removal, or the
		// saver's late writes race the cleanup (directory not empty).
		db.Disconnect(basecontext.NewBaseContext()) //nolint:errcheck // best-effort shutdown
	})
	return db
}

// NewSQLEngine returns the SQL-side engine for a DomainSuite. Task 03 lands
// before GormDatabase implements DatabaseService (task 04), so the SQL side
// is typed as the interface; when no factory is registered the legacy engine
// acts as fallback and the self-proof labels itself accordingly.
func NewSQLEngine(t testing.TB, legacyFallback *data.JsonDatabase) datainterfaces.DatabaseService {
	t.Helper()
	if sqlFactory != nil {
		return sqlFactory(t)
	}
	t.Log("parity: no SQL engine factory registered (task 04 pending) — using legacy fallback")
	return legacyFallback
}

// SeedBoth runs one seed function against both engines. While the SQL side
// falls back to the legacy instance (pre-task-04), a double seed would trip
// uniqueness errors, so the identity pair seeds once.
func SeedBoth(t testing.TB, seed func(db datainterfaces.DatabaseService), legacy, sql datainterfaces.DatabaseService) {
	t.Helper()
	if seed == nil {
		return
	}
	seed(legacy)
	if sql != legacy {
		seed(sql)
	}
}

// AssertSameList fails the test with a readable diff when two lists differ.
func AssertSameList[E any](t testing.TB, what string, want, got []E) {
	t.Helper()
	if reflect.DeepEqual(want, got) {
		return
	}
	t.Errorf("parity mismatch in %s: len(want)=%d len(got)=%d\nwant=%#v\ngot =%#v", what, len(want), len(got), want, got)
	if len(want) == len(got) {
		for i := range want {
			if !reflect.DeepEqual(want[i], got[i]) {
				t.Errorf("  first differing element [%d]:\n    want=%#v\n    got =%#v", i, want[i], got[i])
				break
			}
		}
	}
}

// RunMatrix drives suites: seed both engines, run every check. When t is a
// *testing.T each suite/check becomes a subtest; a bare testing.TB (e.g. from
// a benchmark harness) runs them inline instead.
func RunMatrix(t testing.TB, suites ...DomainSuite) {
	t.Helper()
	for _, suite := range suites {
		suite := suite
		subRunner(t)(suite.Name, func(t testing.TB) {
			legacy := NewLegacyEngine(t)
			sql := NewSQLEngine(t, legacy)
			SeedBoth(t, func(db datainterfaces.DatabaseService) {
				if suite.Seed != nil {
					suite.Seed(t, db)
				}
			}, legacy, sql)
			for i, check := range suite.Checks {
				check := check
				subRunner(t)(fmt.Sprintf("check-%02d", i), func(t testing.TB) {
					check(t, legacy, sql)
				})
			}
		})
	}
}

// subRunner returns a subtest runner bound to t; with a *testing.T it creates
// real subtests, with a bare testing.TB it invokes the body inline.
func subRunner(t testing.TB) func(name string, fn func(t testing.TB)) {
	tt, isT := t.(*testing.T)
	return func(name string, fn func(t testing.TB)) {
		if isT {
			tt.Run(name, func(st *testing.T) { fn(st) })
			return
		}
		fn(t)
	}
}
