package models_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Parity suite (PBI #419 task 05): each fixture is observed twice — once
// through the LEGACY engine's read shape (rehydration included, mirroring
// data/users.go GetUsers), once seeded through GORM and re-read with the
// D-7 preload chain — and the marshalled results must be BYTE-identical.
// Order normalisation: both sides sorted by ID (legacy store order == seed
// insertion order == sqlite rowid order; #421 adds explicit ORDER BY for
// production reads). Join rows are seeded ONLY via associations or the
// explicit RoleClaims model — never hand-written SQL.

func allAutoMigrateModels() []interface{} {
	return []interface{}{
		&models.User{}, &models.Role{}, &models.Claim{}, &models.RoleClaims{}, &models.ApiKey{},
		&models.DatabaseSchema{},
		&models.CatalogManager{}, &models.CatalogManifest{}, &models.Job{}, &models.PackerTemplate{},
		&models.UserConfig{}, &models.ReverseProxy{}, &models.ReverseProxyHost{},
		&models.OrchestratorEnrollmentToken{}, &models.Configuration{},
		&models.OrchestratorHost{}, &models.VMSnapshot{}, &models.VirtualMachine{},
	}
}

func newParityDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := database.DefaultDatabaseConfig()
	cfg.Path = database.MemoryPath
	db, err := database.NewDatabaseConnection(basecontext.NewBaseContext(), cfg)
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(allAutoMigrateModels()...))
	return db
}

func readParityStore(t *testing.T, name string, v interface{}) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "parity", name))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, v))
}

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	out, err := json.Marshal(v)
	require.NoError(t, err)
	return out
}

type usersStore struct {
	Users  []models.User  `json:"users"`
	Roles  []models.Role  `json:"roles"`
	Claims []models.Claim `json:"claims"`
}

// legacyRehydrated mirrors data/users.go:44-56 exactly (in-place refill of
// every embedded role's claims from the live roles table).
func legacyRehydrated(store usersStore) []models.User {
	out := make([]models.User, len(store.Users))
	copy(out, store.Users)
	for u := range out {
		for r := range out[u].Roles {
			for _, dbRole := range store.Roles {
				if strings.EqualFold(dbRole.ID, out[u].Roles[r].ID) {
					out[u].Roles[r].Claims = dbRole.Claims
					break
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func seedUsersRoles(t *testing.T, db *gorm.DB, store usersStore) {
	t.Helper()
	for _, c := range store.Claims {
		require.NoError(t, db.Create(&c).Error) // claims first: pure rows
	}
	for _, r := range store.Roles {
		require.NoError(t, db.Create(&r).Error) // association: role_claims rows
	}
	for _, u := range store.Users {
		// reference-only roles (PK set, claims cleared) — join rows via tags
		refs := make([]models.Role, len(u.Roles))
		for i, rl := range u.Roles {
			refs[i] = models.Role{ID: rl.ID}
		}
		u.Roles = refs
		require.NoError(t, db.Create(&u).Error) // user_roles + user_claims rows
	}
}

func sortedUsers(us []models.User) []models.User {
	out := append([]models.User(nil), us...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func TestUserRead_ParityWithLegacyRehydration(t *testing.T) {
	var store usersStore
	readParityStore(t, "users_store.json", &store)
	want := legacyRehydrated(store)

	db := newParityDB(t)
	seedUsersRoles(t, db, store)

	var got []models.User
	require.NoError(t, db.Preload("Claims").Preload("Roles.Claims").Find(&got).Error)
	require.Equal(t, string(mustJSON(t, want)), string(mustJSON(t, sortedUsers(got))))
}

func TestUserRoleClaims_LiveRehydrationParity(t *testing.T) {
	var store usersStore
	readParityStore(t, "users_store.json", &store)

	db := newParityDB(t)
	seedUsersRoles(t, db, store)

	// role gains a claim AFTER users exist — legacy view changes without
	// touching user rows (users.go:44-56); GORM must agree via role_claims.
	extra := models.Claim{ID: "c-9", Name: "vm.delete", Internal: false}
	require.NoError(t, db.Create(&extra).Error)
	require.NoError(t, db.Create(&models.RoleClaims{RoleID: "r-1", ClaimID: "c-9"}).Error)

	var fresh usersStore
	readParityStore(t, "users_store.json", &fresh)
	for i := range fresh.Roles {
		if fresh.Roles[i].ID == "r-1" {
			fresh.Roles[i].Claims = append(fresh.Roles[i].Claims, extra)
		}
	}
	want := legacyRehydrated(fresh)

	var got []models.User
	require.NoError(t, db.Preload("Claims").Preload("Roles.Claims").Find(&got).Error)
	require.Equal(t, string(mustJSON(t, want)), string(mustJSON(t, sortedUsers(got))))
}

type hostsStore struct {
	Hosts []models.OrchestratorHost `json:"hosts"`
}

func sortedHosts(hs []models.OrchestratorHost) []models.OrchestratorHost {
	out := append([]models.OrchestratorHost(nil), hs...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	for i := range out {
		sort.Slice(out[i].VirtualMachines, func(a, b int) bool { return out[i].VirtualMachines[a].ID < out[i].VirtualMachines[b].ID })
		sort.Slice(out[i].ReverseProxyHosts, func(a, b int) bool { return out[i].ReverseProxyHosts[a].ID < out[i].ReverseProxyHosts[b].ID })
	}
	return out
}

func TestHostRead_ParityWithNPlusOneGuard(t *testing.T) {
	var store hostsStore
	readParityStore(t, "hosts_store.json", &store)
	want := sortedHosts(store.Hosts)

	db := newParityDB(t)
	for _, h := range store.Hosts {
		require.NoError(t, db.Create(&h).Error) // cascade: VMs, proxy, proxy hosts
	}

	// D-7 N+1 guard: the DEFAULT host read must NOT drag associations along.
	var plain []models.OrchestratorHost
	require.NoError(t, db.Find(&plain).Error)
	for _, h := range plain {
		require.Empty(t, h.VirtualMachines, "default Find must not load VMs")
		require.Empty(t, h.ReverseProxyHosts, "default Find must not load proxy hosts")
		require.Nil(t, h.ReverseProxy, "default Find must not load has-one proxy")
	}

	var got []models.OrchestratorHost
	require.NoError(t, db.Preload("ReverseProxy").Preload("ReverseProxyHosts").Find(&got).Error)
	for i := range got { // detail view: explicit VM query per host (D-7)
		var vms []models.VirtualMachine
		require.NoError(t, db.Where("host_id = ?", got[i].ID).Find(&vms).Error)
		got[i].VirtualMachines = vms
	}
	require.Equal(t, string(mustJSON(t, want)), string(mustJSON(t, sortedHosts(got))))
}

type snapshotsStore struct {
	Records []models.HostsVMSnapshotsRecord `json:"records"`
}

func TestSnapshot_RegroupParity(t *testing.T) {
	var store snapshotsStore
	readParityStore(t, "snapshots_store.json", &store)

	db := newParityDB(t)
	var vmids []string
	for _, rec := range store.Records {
		for vmid := range rec.VMSnapshots {
			vmids = append(vmids, vmid)
		}
	}
	sort.Strings(vmids)
	for _, rec := range store.Records {
		for _, vmid := range vmids {
			for _, s := range rec.VMSnapshots[vmid] {
				s.HostID = rec.HostId // carriers from wrapper keys
				s.VMID = vmid
				require.NoError(t, db.Create(&s).Error)
			}
		}
	}

	// regroup query: rows → map[vm_id][]VMSnapshot per host
	for _, rec := range store.Records {
		var rows []models.VMSnapshot
		require.NoError(t, db.Where("host_id = ?", rec.HostId).Find(&rows).Error)
		got := map[string][]models.VMSnapshot{}
		for _, r := range rows {
			got[r.VMID] = append(got[r.VMID], r)
		}
		for k := range got {
			sort.Slice(got[k], func(i, j int) bool { return got[k][i].ID < got[k][j].ID })
		}
		want := map[string][]models.VMSnapshot{}
		for vmid, ss := range rec.VMSnapshots {
			cp := append([]models.VMSnapshot(nil), ss...)
			sort.Slice(cp, func(i, j int) bool { return cp[i].ID < cp[j].ID })
			want[vmid] = cp
		}
		require.Equal(t, string(mustJSON(t, want)), string(mustJSON(t, got)), "host %s", rec.HostId)
	}

	// composite index present for the regroup query
	db2 := db
	var idxCount int64
	require.NoError(t, db2.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_vmsnapshots_host_vm'").Scan(&idxCount).Error)
	require.Equal(t, int64(1), idxCount)
}

func TestSerializer_NilVsEmptyRoundTrip(t *testing.T) {
	db := newParityDB(t)
	rows := []models.PackerTemplate{
		{ID: "pt-nil", Name: "nil-everything"}, // nil map/slices
		{ID: "pt-empty", Name: "empty-collections", Variables: map[string]string{}, Addons: []string{}, Specs: map[string]string{}, Defaults: map[string]string{}},
		{ID: "pt-full", Name: "populated", Variables: map[string]string{"a": "1"}, Addons: []string{"drm"}, Specs: map[string]string{"cpu": "2"}, Defaults: map[string]string{}},
	}
	for i := range rows {
		require.NoError(t, db.Create(&rows[i]).Error)
	}

	// STORAGE fidelity: nil is written as SQL NULL (not "null" text, not "{}").
	var typeofText string
	require.NoError(t, db.Raw("SELECT typeof(variables) FROM packer_templates WHERE id='pt-nil'").Scan(&typeofText).Error)
	require.Equal(t, "null", typeofText, "nil map must be stored as SQL NULL")

	var got []models.PackerTemplate
	require.NoError(t, db.Find(&got).Error)
	byID := map[string]models.PackerTemplate{}
	for _, g := range got {
		byID[g.ID] = g
	}

	// READ-BACK fidelity: three distinct states survive Create→Find and
	// marshal distinctly (null vs {} vs values) — guards the wire contract
	// for serializer fields WITHOUT omitempty (PackerTemplate x4,
	// CatalogManifest x6 are the only such fields in the model set).
	require.Nil(t, byID["pt-nil"].Variables, "nil map must round-trip as nil")
	require.Nil(t, byID["pt-nil"].Addons)
	require.NotNil(t, byID["pt-empty"].Variables, "empty map must stay non-nil")
	require.Len(t, byID["pt-empty"].Variables, 0)
	require.NotNil(t, byID["pt-empty"].Addons)
	require.Len(t, byID["pt-empty"].Addons, 0)
	require.Equal(t, map[string]string{"a": "1"}, byID["pt-full"].Variables)
	for _, want := range rows {
		require.Equal(t, string(mustJSON(t, want)), string(mustJSON(t, byID[want.ID])), want.ID)
	}
}
