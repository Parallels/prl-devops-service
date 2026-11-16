package serviceprovider

// sql_database module reporting: the hardware/host endpoints advertise the
// runtime-derived sql_database module and its engine type only when a live
// SQL (GORM) connection exists — never from operator configuration alone.

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
