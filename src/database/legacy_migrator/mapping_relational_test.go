package legacy_migrator

// Relational units (task 04): the #419 join/edge semantics on the enriched
// golden fixture — canonical-vs-orphan entity dedupe, forward-only edges,
// OnConflict-Ignore idempotency, snapshot-map flattening, and the pinned
// plan order.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func edgeCount(t *testing.T, db *gorm.DB, table string, where string, arg interface{}) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Table(table).Where(where, arg).Count(&n).Error)
	return n
}

func TestRelationalGoldenE2E(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)

	_, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{SourcePath: path})
	require.NoError(t, err)

	// Per-table COUNT(*) vs the #419 validation formulas on the enriched fixture:
	//   claims: 4 top-level + c-12 orphan = 5      roles: r-1,r-2 + r-9 orphan = 3
	//   role_claims: Σ|role claims| deduped = 4     users: 3
	//   user_roles: (u-2:{r-1,r-2,r-9},u-3:{r-2}) = 4   user_claims: 3
	//   manifests: 1 (contents serialised in-row)   virtual_machines: 2 (one host_id derived)
	//   vm_snapshots: 3 (host map, 2 keys) + 1 (wrapper) = 4
	want := map[string]int64{
		"Claims": 5, "Roles": 3, "RoleClaims": 4, "Users": 3,
		"UserRoles": 4, "UserClaims": 3, "ManifestsCatalog": 1,
		"VirtualMachines": 2, "VMSnapshots": 4,
	}
	got := countAllScalars(t, db)
	for name, n := range want {
		assert.Equal(t, n, got[name], name)
	}

	var r9 models.Role
	require.NoError(t, db.First(&r9, "id = ?", "r-9").Error) // orphan embedded role preserved
	assert.Equal(t, "AdHoc Embedded", r9.Name)

	var vm101 models.VirtualMachine
	require.NoError(t, db.First(&vm101, "id = ?", "vm-101").Error)
	assert.Equal(t, "oh-1", vm101.HostId) // derived from parent host (key preserved)
}

func TestSharedRoleAcrossUsers(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	_, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{SourcePath: path})
	require.NoError(t, err)

	assert.Equal(t, int64(1), edgeCount(t, db, "roles", "id = ?", "r-2"), "ONE entity row")
	assert.Equal(t, int64(2), edgeCount(t, db, "user_roles", "role_id = ?", "r-2"), "TWO edge rows")
}

func TestEdgeRerunNoDuplicates(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	ctx := basecontext.NewRootBaseContext()
	_, err := RunMigration(ctx, db, Options{SourcePath: path})
	require.NoError(t, err)
	before := countAllScalars(t, db)

	// Direct spec replay (marker gate bypassed) — the OnConflict-DoNothing path:
	legacy, err := data.LoadLegacyData(ctx, path)
	require.NoError(t, err)
	for _, s := range collectionPlan() {
		_, err := s.Migrate(ctx, db, &legacy, defaultChunkSize)
		require.NoError(t, err, s.Name)
	}
	assert.Equal(t, before, countAllScalars(t, db), "re-run changed no count")
}

func TestOrphanEmbeddedRolePreserved(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	_, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{SourcePath: path})
	require.NoError(t, err)

	var claims []models.Claim
	require.NoError(t, db.Find(&claims).Error)
	var c12 bool
	for _, c := range claims {
		if c.ID == "c-12" {
			c12 = true
		}
	}
	assert.True(t, c12, "orphan claim c-12 (only ever embedded) became an entity row")
	assert.Equal(t, int64(1), edgeCount(t, db, "role_claims", "role_id = ?", "r-9"))
}

func TestBackrefsProduceNoPhantomEdges(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	require.NotPanics(t, func() {
		_, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{SourcePath: path})
		require.NoError(t, err)
	})
	// Role.Users/Claim.Users are json:"-": never in the file, never in the DB.
	// Edges derive from forward direction ONLY ⇒ exact forward counts hold:
	assert.Equal(t, int64(3), countModel(t, db, &models.User{}))
	assert.Equal(t, int64(4), edgeCount(t, db, "user_roles", "1 = ?", 1))
	assert.Equal(t, int64(3), edgeCount(t, db, "user_claims", "1 = ?", 1))
}

func TestSnapshotMapFlattenDeterministic(t *testing.T) {
	dump := func(t *testing.T) []models.VMSnapshot {
		db := migDB(t)
		path := goldenAt(t)
		_, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{SourcePath: path})
		require.NoError(t, err)
		var rows []models.VMSnapshot
		require.NoError(t, db.Order("id").Find(&rows).Error)
		return rows
	}
	first, second := dump(t), dump(t)
	require.Len(t, first, 4)
	assert.Equal(t, first, second, "byte-deterministic flatten across runs")

	byID := map[string]models.VMSnapshot{}
	for _, r := range first {
		byID[r.ID] = r
	}
	assert.Equal(t, "oh-1", byID["os-s-1"].HostID) // record.HostId → host_id
	assert.Equal(t, "vm-uuid-a", byID["os-s-1"].VMID)
	assert.Equal(t, "vm-uuid-b", byID["os-s-2"].VMID) // map keys preserved verbatim (they are data)
	assert.Equal(t, "vm-1", byID["wvs-1"].VMID)       // wrapper collection: VMId carried, host unknown
	assert.Empty(t, byID["wvs-1"].HostID)
}

// TestPlanOrderPinned replaces the spec's "permute ⇒ FK violation" guard:
// #419 declared constraint:- on every relationship, so no dialect here can
// raise an FK error — the order is pinned as an explicit contract instead
// (deviation, ledger entry 38).
func TestPlanOrderPinned(t *testing.T) {
	names := make([]string, 0, 32)
	for _, s := range collectionPlan() {
		names = append(names, s.Name)
	}
	assert.Equal(t, []string{
		"Claims", "Roles", "RoleClaims", "Users", "UserRoles", "UserClaims",
		"ApiKeys", "Configuration", "PackerTemplates", "ManifestsCatalog",
		"OrchestratorHosts", "VirtualMachines", "VMSnapshots",
		"ReverseProxy", "ReverseProxyHosts", "CatalogManagers", "Jobs",
		"EnrollmentTokens", "UserConfigs",
	}, names)
}
