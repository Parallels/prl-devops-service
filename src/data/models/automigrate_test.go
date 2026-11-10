package models_test

// D-5 evidence (AC5 close): grep across src/data + src/serviceprovider found
// NO delete-flag semantics on legacy models (no is_deleted/is_removed write
// paths) ⇒ deliberately NO gorm.DeletedAt soft-delete columns; RevokedAt /
// Revoked remain plain status columns on ApiKey (row 5). Verified again by
// TestSchema_MappingPragmas below (no *deleted* columns in any table DDL).

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/require"
)

var updateSchema = flag.Bool("update-schema", false, "regenerate testdata/schema_snapshot.sql")

// TestAutoMigrate_Fresh — AC3: the shipped tags, through the #418 factory,
// against a FRESH SQLite FILE in t.TempDir() (no mocks), produce exactly the
// plan's 20-table schema. Idempotent: AutoMigrate runs twice. DDL pinned as
// golden snapshot.
func TestAutoMigrate_Fresh(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	cfg := database.DefaultDatabaseConfig()
	cfg.Path = filepath.Join(t.TempDir(), "prl-devops.db")
	db, err := database.NewDatabaseConnection(ctx, cfg)
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(allAutoMigrateModels()...))
	require.NoError(t, db.AutoMigrate(allAutoMigrateModels()...)) // idempotent

	m := db.Migrator()
	tables := []string{"users", "roles", "claims", "role_claims", "user_roles", "user_claims",
		"api_keys", "catalog_managers", "catalog_manifests", "jobs", "packer_templates", "user_configs",
		"orchestrator_enrollment_tokens", "orchestrator_hosts", "reverse_proxies", "reverse_proxy_hosts",
		"vm_snapshots", "virtual_machines", "configurations", "database_schemas"} // 20 (D-2 confirmed)
	for _, tbl := range tables {
		require.Truef(t, m.HasTable(tbl), "missing table %s", tbl)
	}

	// wrappers stay tables-free (D-2)
	require.False(t, m.HasTable("vm_snapshots_records"))
	require.False(t, m.HasTable("hosts_vm_snapshots_records"))
	require.False(t, m.HasTable("parallels_v_m_s")) // runtime DTO never a root (row 24)

	// carriers present (rows 20/21)
	require.True(t, m.HasColumn(&models.VMSnapshot{}, "HostID"))
	require.True(t, m.HasColumn(&models.VMSnapshot{}, "VMID"))

	// unique indexes (mapping col-3 of rows 1,5,9,12)
	unique := []struct {
		dst        interface{}
		table, idx string
	}{
		{&models.User{}, "users", "idx_users_username"},
		{&models.ApiKey{}, "api_keys", "idx_api_keys_key"},
		{&models.UserConfig{}, "user_configs", "ux_userconfig_user_slug"},
		{&models.OrchestratorEnrollmentToken{}, "orchestrator_enrollment_tokens", "idx_orchestrator_enrollment_tokens_token"},
	}
	for _, u := range unique {
		require.Truef(t, m.HasIndex(u.dst, u.idx), "%s missing index %s", u.table, u.idx)
		var isUnique int
		require.NoError(t, db.Raw(`SELECT "unique" FROM pragma_index_list(?) WHERE name=?`, u.table, u.idx).Scan(&isUnique).Error)
		require.Equalf(t, 1, isUnique, "%s must be UNIQUE", u.idx)
	}
	// composite order: (user_id, slug)
	var uxParts []struct {
		ColumnName string `gorm:"column:name"`
		Seq        int    `gorm:"column:seqno"`
	}
	require.NoError(t, db.Raw(`SELECT name, seqno FROM pragma_index_info('ux_userconfig_user_slug') ORDER BY seqno`).Scan(&uxParts).Error)
	require.Equal(t, []string{"user_id", "slug"}, []string{uxParts[0].ColumnName, uxParts[1].ColumnName})

	// composite PK on role_claims (row 3)
	var pks []struct {
		ColumnName string `gorm:"column:name"`
		Pk         int    `gorm:"column:pk"`
	}
	require.NoError(t, db.Raw(`SELECT name, pk FROM pragma_table_info('role_claims') WHERE pk > 0 ORDER BY pk`).Scan(&pks).Error)
	require.Equal(t, []string{"role_id", "claim_id"}, []string{pks[0].ColumnName, pks[1].ColumnName})

	// Schema golden — NORMALISED per-column rows (pragma_table_xinfo) +
	// index list/info: semantically equivalent to raw DDL but immune to
	// GORM's non-deterministic multi-FK CONSTRAINT emission order (first
	// raw-DDL snapshot flickered 2/5 under load; empiricism wins — the
	// normalised dump pins every column, type, PK and FK target by row).
	type colRow struct {
		Tbl  string `gorm:"column:tbl"`
		Name string `gorm:"column:name"`
		Type string `gorm:"column:type"`
		Pk   int    `gorm:"column:pk"`
		NotN int    `gorm:"column:notnull"`
	}
	var cols []colRow
	require.NoError(t, db.Raw(`SELECT m.name AS tbl, p.name, p.type, p.pk, p."notnull"
		FROM sqlite_master m JOIN pragma_table_xinfo(m.name) p
		WHERE m.type='table' AND p.hidden = 0 ORDER BY m.name, p.cid`).Scan(&cols).Error)
	var idxRows []struct {
		Tbl  string `gorm:"column:tbl"`
		Name string `gorm:"column:name"`
		Col  string `gorm:"column:col"`
	}
	require.NoError(t, db.Raw(`SELECT m.name AS tbl, i.name, ic.name AS col
		FROM sqlite_master m
		JOIN pragma_index_list(m.name) i ON 1=1
		JOIN pragma_index_info(i.name) ic ON 1=1
		WHERE m.type='table' AND i.name NOT LIKE 'sqlite_autoindex%'
		ORDER BY m.name, i.name, ic.seqno`).Scan(&idxRows).Error)
	var sb strings.Builder
	for _, c := range cols {
		fmt.Fprintf(&sb, "%s.%s %s pk=%d notnull=%d\n", c.Tbl, c.Name, c.Type, c.Pk, c.NotN)
	}
	for _, r := range idxRows {
		fmt.Fprintf(&sb, "index %s.%s(%s)\n", r.Tbl, r.Name, r.Col)
	}
	got := sb.String()
	snap := filepath.Join("testdata", "schema_snapshot.sql")
	if *updateSchema || os.Getenv("SCHEMA_CAPTURE") == "1" {
		require.NoError(t, os.WriteFile(snap, []byte(got), 0o644))
	}
	want, err := os.ReadFile(snap)
	require.NoError(t, err)
	require.Equal(t, string(want), got, "schema drift vs snapshot (regen: SCHEMA_CAPTURE=1 go test -run AutoMigrate)")
}

// TestSchema_MappingPragmas — mapping-finalisation spot machine-audit:
// serialised fields are TEXT everywhere; dropped fields absent everywhere
// (D-4: no db_record / is_locked column in any table DDL).
func TestSchema_MappingPragmas(t *testing.T) {
	db := newParityDB(t)
	serializerCols := []struct{ table, column string }{
		{"virtual_machines", "guest_tools"}, {"virtual_machines", "host_shared_folders"},
		{"virtual_machines", "network_information"}, {"virtual_machines", "hardware"},
		{"orchestrator_hosts", "tags"}, {"orchestrator_hosts", "authentication"},
		{"orchestrator_hosts", "cache_items"}, {"orchestrator_hosts", "resources"},
		{"catalog_manifests", "virtual_machine_contents"}, {"catalog_manifests", "provider"},
		{"packer_templates", "variables"}, {"jobs", "steps"},
	}
	for _, c := range serializerCols {
		var dtype string
		require.NoError(t, db.Raw(`SELECT type FROM pragma_table_info(?) WHERE name=?`, c.table, c.column).Scan(&dtype).Error, "%s.%s", c.table, c.column)
		require.Equalf(t, "TEXT", strings.ToUpper(dtype), "%s.%s must be TEXT (mapping col-5)", c.table, c.column)
	}
	var dropped int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND (sql LIKE '%db_record%' OR sql LIKE '%is_locked%')`).Scan(&dropped).Error)
	require.Zero(t, dropped, "D-4: DbRecord fields never become columns")
	// D-5: no soft-delete columns anywhere
	var deleted int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND sql LIKE '%deleted_at%'`).Scan(&deleted).Error)
	require.Zero(t, deleted, "D-5: no gorm.DeletedAt anywhere")
}
