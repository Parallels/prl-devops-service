package parity

// #425 task 03 — infra wave: catalog + catalog_managers (legacy fixtures
// from src/data/catalog_test.go + catalog_managers_test.go; bodies
// transcribed 1:1 — arrangements and sentinels preserved, never weakened).
// Plus the D-5 ordering differential: identical pre-seeded out-of-order
// UpdatedAt stamps on both engines (E-4 pattern), assert engine-equal
// sequences (cross-engine equality, not "correctness" — D-5).

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestParityCatalogInfraWave(t *testing.T) {
	testutil.RunOnEngines(t, catalogScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
	testutil.RunOnEngines(t, catalogManagerScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
}

func catalogScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "catalog/create_read_update_roundtrip", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateCatalogManifest(ctx, models.CatalogManifest{
				ID: "infra-manifest-1", Name: "test-manifest", Version: "1.0.0", Architecture: "amd64",
			})
			require.NoError(t, err)
			assert.NotEmpty(t, created.CreatedAt)
			assert.NotEmpty(t, created.UpdatedAt)
			assert.Equal(t, "test-manifest", created.Name)
			loaded, err := db.GetCatalogManifestByName(ctx, created.ID) // ID-OR-name lookup (catalog.go:55)
			require.NoError(t, err)
			assert.Equal(t, "test-manifest", loaded.Name)
			// update leg (catalog_test.go:40): stamps must move
			time.Sleep(10 * time.Millisecond) // arrangement fidelity
			updatedManifest := *created
			updatedManifest.Name = "updated-manifest"
			updated, err := db.UpdateCatalogManifest(ctx, updatedManifest)
			require.NoError(t, err)
			assert.Equal(t, "updated-manifest", updated.Name)
			assert.NotEqual(t, created.UpdatedAt, updated.UpdatedAt)
		}},
		{Name: "catalog/delete_manifest", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateCatalogManifest(ctx, models.CatalogManifest{
				ID: "infra-manifest-del", Name: "del-manifest", Version: "1.0.0", Architecture: "amd64",
			})
			require.NoError(t, err)
			require.NoError(t, db.DeleteCatalogManifest(ctx, created.ID))
			_, err = db.GetCatalogManifestByName(ctx, created.ID)
			require.Error(t, err)
			assert.Equal(t, data.ErrCatalogManifestNotFound, err)
		}},
		{Name: "catalog/delete_by_catalog_id_cascades", Run: func(t *testing.T, _ ctxT, db dbT) {
			// DeleteCatalogManifest matches ID OR CatalogId OR Name (catalog.go:236-238).
			// PINNED parity behaviour: CreateCatalogManifest regenerates the ID
			// (both engines ⇒ content-hash hex), so survivors are keyed by
			// CatalogId, never by the caller-supplied ID literal.
			ctx := tplCtxRoot()
			for _, m := range []models.CatalogManifest{
				{ID: "pair-a", CatalogId: "cat-shared", Version: "1.0.0", Architecture: "amd64"},
				{ID: "pair-b", CatalogId: "cat-shared", Version: "2.0.0", Architecture: "arm64"},
				{ID: "lone-c", CatalogId: "cat-other", Version: "1.0.0", Architecture: "amd64"},
			} {
				_, err := db.CreateCatalogManifest(ctx, m)
				require.NoError(t, err)
			}
			require.NoError(t, db.DeleteCatalogManifest(ctx, "cat-shared"))
			all, err := db.GetCatalogManifests(ctx, "")
			require.NoError(t, err)
			require.Len(t, all, 1, "both CatalogId-mates gone, lone survivor")
			// PINNED: create normalises CatalogId to UPPER CASE (both engines).
			assert.Equal(t, "CAT-OTHER", all[0].CatalogId)
		}},
		{Name: "catalog/serialised_array_columns_fidelity", Run: func(t *testing.T, _ ctxT, db dbT) {
			// Spec invariants: Tags/RequiredRoles/RequiredClaims must survive
			// create + read-modify-write WITHOUT reordering elements.
			ctx := tplCtxRoot()
			created, err := db.CreateCatalogManifest(ctx, models.CatalogManifest{
				ID: "infra-manifest-arr", Name: "arr-manifest", Version: "1.0.0", Architecture: "amd64",
				Tags: []string{"zeta", "alpha", "mid"}, RequiredRoles: []string{"operator", "viewer"},
				RequiredClaims: []string{"c-second", "c-first"},
			})
			require.NoError(t, err)
			loaded, err := db.GetCatalogManifestByName(ctx, created.ID)
			require.NoError(t, err)
			// PINNED parity behaviour: create auto-appends the "latest" tag on
			// BOTH engines (append position observed: after supplied tags).
			assert.Equal(t, []string{"zeta", "alpha", "mid", "latest"}, loaded.Tags, "tags order preserved + latest appended")
			assert.Equal(t, []string{"operator", "viewer"}, loaded.RequiredRoles)
			assert.Equal(t, []string{"c-second", "c-first"}, loaded.RequiredClaims)
			// read-modify-write (catalog_test.go:94 leg): append keeps position
			loaded.Tags = append(loaded.Tags, "tag2")
			require.NoError(t, db.UpdateCatalogManifestTags(ctx, *loaded))
			again, err := db.GetCatalogManifestByName(ctx, created.ID)
			require.NoError(t, err)
			assert.Equal(t, []string{"zeta", "alpha", "mid", "latest", "tag2"}, again.Tags)
			// untouched columns survive a tags-only update
			assert.Equal(t, []string{"operator", "viewer"}, again.RequiredRoles)
		}},
		{Name: "catalog/taint_untaint_revoke_version", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			_, err := db.CreateCatalogManifest(ctx, models.CatalogManifest{
				ID: "infra-manifest-taint", CatalogId: "taint-catalog", Version: "1.0.0", Architecture: "amd64",
			})
			require.NoError(t, err)
			tainted, err := db.TaintCatalogManifestVersion(ctx, "taint-catalog", "1.0.0")
			require.NoError(t, err)
			assert.True(t, tainted.Tainted)
			untainted, err := db.UnTaintCatalogManifestVersion(ctx, "taint-catalog", "1.0.0")
			require.NoError(t, err)
			assert.False(t, untainted.Tainted)
			revoked, err := db.RevokeCatalogManifestVersion(ctx, "taint-catalog", "1.0.0")
			require.NoError(t, err)
			assert.True(t, revoked.Revoked)
		}},
	}
}

func catalogManagerScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "managers/add_update_get", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			mgr := models.CatalogManager{
				ID: "infra-mgr-1", Name: "Infra Manager", URL: "https://example.com/catalog",
				AuthenticationMethod: "None",
			}
			require.NoError(t, db.AddCatalogManager(ctx, mgr))
			loaded, err := db.GetCatalogManager("infra-mgr-1")
			require.NoError(t, err)
			assert.Equal(t, "Infra Manager", loaded.Name)
			mgr.Name = "Updated Manager"
			require.NoError(t, db.UpdateCatalogManager(ctx, mgr))
			updated, err := db.GetCatalogManager("infra-mgr-1")
			require.NoError(t, err)
			assert.Equal(t, "Updated Manager", updated.Name)
		}},
		{Name: "managers/delete", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			require.NoError(t, db.AddCatalogManager(ctx, models.CatalogManager{
				ID: "infra-mgr-del", Name: "Del Manager", URL: "https://example.com",
				AuthenticationMethod: "None",
			}))
			require.NoError(t, db.DeleteCatalogManager(ctx, "infra-mgr-del"))
			_, err := db.GetCatalogManager("infra-mgr-del")
			require.Error(t, err)
			assert.Equal(t, data.ErrCatalogManagerNotFound, err)
		}},
		{Name: "managers/internal_guards_without_root", Run: func(t *testing.T, _ ctxT, db dbT) {
			// managers_test.go:90/114: plain ctx may ADD internal, never update/delete it.
			ctx := authCtxPlain()
			require.NoError(t, db.AddCatalogManager(ctx, models.CatalogManager{
				ID: "infra-mgr-internal", Name: "Internal Manager", Internal: true,
				AuthenticationMethod: "None",
			}))
			err := db.UpdateCatalogManager(ctx, models.CatalogManager{
				ID: "infra-mgr-internal", Name: "Attempt Update", Internal: true,
				AuthenticationMethod: "None",
			})
			require.Error(t, err)
			assert.Equal(t, data.ErrUpdateInternalCatalogManager, err)
			err = db.DeleteCatalogManager(ctx, "infra-mgr-internal")
			require.Error(t, err)
			assert.Equal(t, data.ErrRemoveInternalCatalogManager, err)
		}},
	}
}

// ---------- ordering differential (D-5) ----------

func TestParityCatalogOrderDifferential(t *testing.T) {
	// Same three stamps pre-seeded out of insertion order on both engines;
	// GetCatalogManifests orders UpdatedAt desc on both (data/catalog.go:36,
	// database/catalog.go:85). Distinct same-width stamps ⇒ lexicographic ==
	// chronological by construction (no T14 exposure). Assert ENGINE EQUALITY.
	seedRows := []models.CatalogManifest{ // insertion order a, c, b
		{ID: "ord-a", Name: "ord-a", Version: "1", Architecture: "amd64", CreatedAt: "2020-01-01T00:00:00Z", UpdatedAt: "2020-01-01T00:00:00Z"},
		{ID: "ord-c", Name: "ord-c", Version: "1", Architecture: "amd64", CreatedAt: "2020-03-01T00:00:00Z", UpdatedAt: "2020-03-01T00:00:00Z"},
		{ID: "ord-b", Name: "ord-b", Version: "1", Architecture: "amd64", CreatedAt: "2020-02-01T00:00:00Z", UpdatedAt: "2020-02-01T00:00:00Z"},
	}
	preWrite := func(t *testing.T, dbPath string) {
		store := data.Data{ManifestsCatalog: seedRows}
		b, err := json.Marshal(store)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dbPath, b, 0o640))
	}
	rawInsert := func(t *testing.T, gdb *gorm.DB) {
		for i := range seedRows {
			require.NoError(t, gdb.Table("catalog_manifests").Create(&seedRows[i]).Error)
		}
	}
	legacySvc, legacyCloser := authLegacyEngine(preWrite).New(t)
	t.Cleanup(legacyCloser)
	gormSvc, gormCloser := authGormEngine(rawInsert).New(t)
	t.Cleanup(gormCloser)

	ctx := tplCtxRoot()
	seqL := idsOf(allManifests(t, ctx, legacySvc))
	seqG := idsOf(allManifests(t, ctx, gormSvc))
	assert.Equal(t, []string{"ord-c", "ord-b", "ord-a"}, seqL, "pinned: UpdatedAt desc")
	assert.Equal(t, seqL, seqG, "cross-engine order equality (D-5)")
}

func allManifests(t *testing.T, ctx ctxT, db dbT) []models.CatalogManifest {
	all, err := db.GetCatalogManifests(ctx, "")
	require.NoError(t, err)
	require.NotEmpty(t, all, "non-vacuity: seeds listed")
	return all
}

func idsOf(ms []models.CatalogManifest) []string {
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		out = append(out, m.ID)
	}
	return out
}
