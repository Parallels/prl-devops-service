package models_test

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/require"
)

// TestWaveB_SmokeAutoMigrate — smoke only; the full AutoMigrate matrix lands
// in task 06. Proves wave-B tag shapes are schema-buildable in one pass:
// PKs, indexes, the composite unique index, serializer:json columns as TEXT,
// and DbRecord fully excluded (D-4).
func TestWaveB_SmokeAutoMigrate(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	cfg := database.DefaultDatabaseConfig()
	cfg.Path = database.MemoryPath // in-memory sqlite, no filesystem side effects

	db, err := database.NewDatabaseConnection(ctx, cfg)
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(
		&models.CatalogManager{}, &models.CatalogManifest{}, &models.Job{},
		&models.PackerTemplate{}, &models.UserConfig{}, &models.ReverseProxy{},
		&models.ReverseProxyHost{}, &models.OrchestratorEnrollmentToken{}, &models.Configuration{},
	))

	// composite unique index exists under the declared symbol
	require.True(t, db.Migrator().HasIndex(&models.UserConfig{}, "ux_userconfig_user_slug"),
		"composite unique (user_id, slug) missing")

	// serializer columns materialise as TEXT (D-3 storage contract), asserted
	// per-column via pragma_table_info — robust against identifier quoting.
	for _, col := range []string{"tags", "provider", "virtual_machine_contents", "pack_contents",
		"required_roles", "required_claims", "minimum_spec_requirements"} {
		var ctype string
		require.NoError(t, db.Raw(
			"SELECT LOWER(type) FROM pragma_table_info('catalog_manifests') WHERE name = ?", col).
			Scan(&ctype).Error)
		require.Equal(t, "text", ctype, "serializer column %s must be TEXT", col)
	}

	// DbRecord columns must NOT exist anywhere (gorm:"-:all", D-4)
	for _, tbl := range []string{"jobs", "user_configs", "orchestrator_enrollment_tokens"} {
		var n int64
		require.NoError(t, db.Raw(
			"SELECT COUNT(*) FROM pragma_table_info(?) WHERE name IN ('is_locked','locked_by','locked_at')", tbl).
			Scan(&n).Error)
		require.Zero(t, n, "%s leaked DbRecord columns", tbl)
	}

	// explicit result_record_id column (field ResultRecordId) per spec
	require.True(t, db.Migrator().HasColumn(&models.Job{}, "result_record_id"))
}
