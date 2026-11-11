package parity

// #425 task 02 — auth-wave seed helpers + engine variants (E-4: these two
// factories mirror testutil.LegacyEngine/testutil.GormEngine 1:1 and ONLY
// add a seeding hook, because three converted scenarios (old-API-key
// normalisation) must simulate PRE-TYPE data: rows with Type=="" that no
// public CreateApiKey can produce (both engines default "external" at
// create — verified data/api_key.go:169, database/api_key.go:189). The
// legacy original tests reached into the private db.data.ApiKeys slice;
// engine-faithful equivalents: legacy = pre-Write data.json (the loader
// ingests it verbatim, legacy_load.go:80), gorm = raw INSERT after
// MigrateSchema. All other scenarios keep the vanilla task-01 engines.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/database/testutil"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Type aliases keeping the converted scenario bodies inside the file-size
// budget while preserving the exact Scenario.Run signature.
type (
	ctxT = basecontext.ApiContext
	dbT  = datainterfaces.DatabaseService
)

// authCtxPlain == basecontext.NewBaseContext() posture (nil
// AuthorizationContext — IsAuthorized FALSE for wrapped records, matching
// the legacy ctx the source tests used; authz_test.go semantics).
func authCtxPlain() basecontext.ApiContext { return &parityCtx{} }

// authCtxClaims = user ctx with injected claims (catalog-manager/forwarded
// posture, authorize_layer.go hasInjected branch).
func authCtxClaims(claims ...string) basecontext.ApiContext {
	return &parityCtx{auth: &basecontext.AuthorizationContext{
		User: &apimodels.ApiUser{ID: "ext-user"}, InjectedClaims: claims}}
}

// seedDefaultRolesClaims replays the `_, _ =` default seeds the legacy
// fixtures use (errors intentionally ignored, fidelity of arrangement).
func seedDefaultRolesClaims(ctx basecontext.ApiContext, db datainterfaces.DatabaseService) {
	for _, r := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: r, ID: r})
	}
	for _, c := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: c, ID: c})
	}
}

var authLegacyMu sync.Mutex // same serialise discipline as testutil.LegacyEngine

// authLegacyEngine = testutil.LegacyEngine + optional pre-Write hook (runs
// before Connect so the JSON loader ingests the seeded file verbatim).
func authLegacyEngine(preWrite func(t *testing.T, dbPath string)) testutil.Engine {
	return testutil.Engine{
		Name: "legacy",
		New: func(t *testing.T) (datainterfaces.DatabaseService, func()) {
			ctx := testutil.WithCleanConfig(t)
			authLegacyMu.Lock()
			t.Cleanup(authLegacyMu.Unlock)
			data.ResetForTesting()
			t.Cleanup(data.ResetForTesting)
			dbPath := filepath.Join(t.TempDir(), "test_db.json")
			if preWrite != nil {
				preWrite(t, dbPath)
			}
			db := data.NewJsonDatabase(ctx, dbPath)
			require.NotNil(t, db)
			require.NoError(t, db.Connect(ctx))
			return db, func() { _ = db.Disconnect(ctx) }
		},
		Capabilities: testutil.EngineCapabilities{HasNotConnectedSemantics: true, IsFileEngine: true},
	}
}

// authGormEngine = testutil.GormEngine + optional post-Migrate hook (raw
// handle for pre-Type row injection).
func authGormEngine(postMigrate func(t *testing.T, gdb *gorm.DB)) testutil.Engine {
	return testutil.Engine{
		Name: "gorm",
		New: func(t *testing.T) (datainterfaces.DatabaseService, func()) {
			ctx := testutil.WithCleanConfig(t)
			gdb := testutil.OpenMemoryDB(t)
			testutil.MigrateSchema(t, gdb)
			if postMigrate != nil {
				postMigrate(t, gdb)
			}
			svc := database.NewGormDatabase(ctx, gdb)
			require.NoError(t, svc.Connect(ctx))
			return svc, func() { _ = svc.Disconnect(ctx) }
		},
		Capabilities: testutil.EngineCapabilities{HasNotConnectedSemantics: true, IsFileEngine: false},
	}
}

// preWriteLegacyKeys seeds a data.json holding Type=="" keys BEFORE
// Connect — the legacy posture of keys created before the Type field
// existed (secret is stored as-is; these keys are never auth-checked).
func preWriteLegacyKeys(ids []string) func(t *testing.T, dbPath string) {
	return func(t *testing.T, dbPath string) {
		store := data.Data{}
		for _, id := range ids {
			store.ApiKeys = append(store.ApiKeys, models.ApiKey{
				ID: id, Name: id, Key: id, Secret: "hashed_" + id,
				CreatedAt: "2020-01-01T00:00:00Z", UpdatedAt: "2020-01-01T00:00:00Z",
			})
		}
		b, err := json.Marshal(store)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dbPath, b, 0o640))
	}
}

// insertRawLegacyKeys is the SQL twin of preWriteLegacyKeys: raw INSERT
// bypassing CreateApiKey so Type stays "" exactly as legacy_migrator would
// land rows from a pre-Type JSON store.
func insertRawLegacyKeys(ids []string) func(t *testing.T, gdb *gorm.DB) {
	return func(t *testing.T, gdb *gorm.DB) {
		for _, id := range ids {
			require.NoError(t, gdb.Table("api_keys").Create(&models.ApiKey{
				ID: id, Name: id, Key: id, Secret: "hashed_" + id,
				CreatedAt: "2020-01-01T00:00:00Z", UpdatedAt: "2020-01-01T00:00:00Z",
			}).Error)
		}
	}
}
