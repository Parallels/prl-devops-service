package serviceprovider

// Runtime-derived module reporting: the hardware/host endpoints advertise the
// runtime-derived sql_database module and its engine type only when a live SQL
// (GORM) connection exists, and the placement module only when capacity
// admission is actually guarding dispatch — never from operator configuration
// alone.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSQLDatabaseType_NoLiveConnection(t *testing.T) {
	orig := globalProvider
	defer func() { globalProvider = orig }()

	globalProvider = nil
	assert.Equal(t, "", GetSQLDatabaseType(), "missing provider must report no engine")

	globalProvider = &ServiceProvider{}
	assert.Equal(t, "", GetSQLDatabaseType(), "legacy posture (nil SQLDatabase) must report no engine")
}

func TestGetSQLDatabaseType_LiveSqliteConnection(t *testing.T) {
	ctx := providerTestCtx()
	orig := globalProvider
	defer func() { globalProvider = orig }()

	db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{Type: database.DBTypeSQLite, Path: database.MemoryPath})
	require.NoError(t, err)
	defer database.Close(ctx, db)

	globalProvider = &ServiceProvider{SQLDatabase: db}
	assert.Equal(t, database.DBTypeSQLite, GetSQLDatabaseType())
}

func TestEnabledModulesForReport_LegacyOmitsSqlDatabase(t *testing.T) {
	resetProviderConfig(t)
	orig := globalProvider
	defer func() { globalProvider = orig }()

	globalProvider = &ServiceProvider{}
	modules := EnabledModulesForReport()
	assert.Contains(t, modules, constants.API_MODE, "base module list must stay intact")
	assert.NotContains(t, modules, constants.SQL_DATABASE_MODE)
}

func TestEnabledModulesForReport_AppendsSqlDatabaseModule(t *testing.T) {
	resetProviderConfig(t)
	ctx := providerTestCtx()
	orig := globalProvider
	defer func() { globalProvider = orig }()

	db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{Type: database.DBTypeSQLite, Path: database.MemoryPath})
	require.NoError(t, err)
	defer database.Close(ctx, db)

	globalProvider = &ServiceProvider{SQLDatabase: db}

	base := config.Get().GetEnabledModules()
	modules := EnabledModulesForReport()
	assert.Equal(t, append(base, constants.SQL_DATABASE_MODE), modules,
		"report must equal the operator list plus the appended sql_database module")
}

// withSqlConn installs a live in-memory SQL handle (or none) behind the
// provider, which is what makes a runtime-derived module appear or disappear.
func withSqlConn(t *testing.T, live bool) {
	t.Helper()

	orig := globalProvider
	t.Cleanup(func() { globalProvider = orig })

	if !live {
		globalProvider = &ServiceProvider{}

		return
	}

	db, err := database.NewDatabaseConnection(providerTestCtx(),
		database.DatabaseConfig{Type: database.DBTypeSQLite, Path: database.MemoryPath})
	require.NoError(t, err)
	t.Cleanup(func() { database.Close(providerTestCtx(), db) })

	globalProvider = &ServiceProvider{SQLDatabase: db}
}

// TestEnabledModulesForReport_PlacementIsReportedOnlyWhenEnforcing pins the
// predicate, not just the happy path. The flag alone must not be enough: on the
// legacy JSON store the reservation ledger cannot exist, so advertising
// placement would tell an operator checking whether the over-commit bug is closed
// that it is guarded when nothing is guarding anything.
func TestEnabledModulesForReport_PlacementIsReportedOnlyWhenEnforcing(t *testing.T) {
	cases := []struct {
		name        string
		modules     string
		sql         bool
		enabled     string
		wantVisible bool
	}{
		{"orchestrator with ledger", "api,orchestrator", true, "true", true},
		{"switch off", "api,orchestrator", true, "false", false},
		{"no ledger behind the switch", "api,orchestrator", false, "true", false},
		{"a host does not admit", "api,host", true, "true", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ENABLED_MODULES", tc.modules)
			t.Setenv(constants.PlacementReservationEnabledEnvVar, tc.enabled)
			resetProviderConfig(t)
			withSqlConn(t, tc.sql)

			modules := EnabledModulesForReport()
			if tc.wantVisible {
				assert.Contains(t, modules, constants.PLACEMENT_MODE)
				// Appended, never replacing what the operator configured.
				assert.Contains(t, modules, constants.SQL_DATABASE_MODE)

				return
			}
			assert.NotContains(t, modules, constants.PLACEMENT_MODE)
		})
	}
}

// TestPlacementModeIsNotAnOperationalModule: placement reports itself, it is not
// switched. Listing it in ENABLED_MODULES must do nothing — the real switch is
// PLACEMENT_RESERVATION_ENABLED — and a module round-trip must not be able to
// write it back as if it were configured.
func TestPlacementModeIsNotAnOperationalModule(t *testing.T) {
	assert.NotContains(t, constants.VALID_MODULES, constants.PLACEMENT_MODE,
		"placement must not become switchable through ENABLED_MODULES")

	t.Setenv("ENABLED_MODULES", "api,orchestrator,placement")
	resetProviderConfig(t)
	withSqlConn(t, true)

	assert.NotContains(t, config.Get().GetEnabledModules(), constants.PLACEMENT_MODE,
		"an operator typing placement into the module list must not be believed")

	// And with the switch off it stays absent even though it was requested.
	t.Setenv(constants.PlacementReservationEnabledEnvVar, "false")
	resetProviderConfig(t)
	assert.NotContains(t, EnabledModulesForReport(), constants.PLACEMENT_MODE)
}
