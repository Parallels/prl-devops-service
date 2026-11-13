package database

import (
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FromApplicationConfig resolve/defaults matrix (PBI #420 task 02).
// resetConfig from config_test.go gives isolation (config.New resets globalConfig).
// t.Setenv auto-restores and blocks parallel subtests.

func TestFromApplicationConfig_UnsetMatchesDefaultByteIdentical(t *testing.T) {
	resetConfig(t)
	cfg, err := FromApplicationConfig()
	require.NoError(t, err)
	assert.Equal(t, DefaultDatabaseConfig(), cfg) // D-5: defaults delegated, byte-identical
	assert.True(t, cfg.AutoMigrate)
	assert.Equal(t, DBTypeSQLite, cfg.Type)
}

func TestFromApplicationConfig_PostgresDSNPassthrough(t *testing.T) {
	resetConfig(t)
	t.Setenv("DB_TYPE", "postgres")
	dsn := testDSN("u", "p", "db.internal:5432", "devops?sslmode=require")
	t.Setenv("DB_DSN", dsn)
	cfg, err := FromApplicationConfig()
	require.NoError(t, err)
	assert.Equal(t, DBTypePostgres, cfg.Type)
	assert.Equal(t, dsn, cfg.DSN) // DSN reaches the struct verbatim (redaction is #418's)
}

func TestFromApplicationConfig_PostgresWithoutDSN(t *testing.T) {
	resetConfig(t)
	t.Setenv("DB_TYPE", "postgres")
	_, err := FromApplicationConfig()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPostgresDSNRequired) // AC-3: required-when-postgres
	assert.Contains(t, err.Error(), "DB_DSN")
	assert.Contains(t, err.Error(), "db_dsn") // actionable: names both spellings
}

func TestFromApplicationConfig_UnsupportedType(t *testing.T) {
	resetConfig(t)
	t.Setenv("DB_TYPE", "mysql")
	_, err := FromApplicationConfig()
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedDatabaseType)
	assert.Contains(t, err.Error(), "mysql") // echoes the user's own value
	assert.True(t, strings.Contains(err.Error(), "sqlite|postgres") ||
		strings.Contains(err.Error(), "sqlite"), "message names allowed values")
}

func TestFromApplicationConfig_CaseInsensitiveValues(t *testing.T) {
	resetConfig(t)
	t.Setenv("DB_TYPE", " POSTGRES ") // trimmed + lowered by bridge
	t.Setenv("DB_DSN", "postgres://x")
	cfg, err := FromApplicationConfig()
	require.NoError(t, err)
	assert.Equal(t, DBTypePostgres, cfg.Type)
}

func TestFromApplicationConfig_AutoMigrateFalse(t *testing.T) {
	resetConfig(t)
	t.Setenv("DB_AUTO_MIGRATE", "false")
	cfg, err := FromApplicationConfig()
	require.NoError(t, err) // false is a VALID value, not a parse error (D-3)
	assert.False(t, cfg.AutoMigrate)
}

func TestFromApplicationConfig_AutoMigrateGarbageFails(t *testing.T) {
	resetConfig(t)
	t.Setenv("DB_AUTO_MIGRATE", "maybe")
	cfg, err := FromApplicationConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db_auto_migrate")
	assert.Equal(t, DatabaseConfig{}, cfg) // zero config returned on error
}

func TestFromApplicationConfig_DBPathBeatsDatabaseFolder(t *testing.T) {
	resetConfig(t)
	folder := t.TempDir()
	path := "/srv/parallels-devops/alt.db"
	t.Setenv("DATABASE_FOLDER", folder)
	t.Setenv("DB_PATH", path) // OQ-420-4: DB_PATH wins
	cfg, err := FromApplicationConfig()
	require.NoError(t, err)
	assert.Equal(t, path, cfg.Path)
	assert.NotEqual(t, DefaultDatabaseConfig().Path, cfg.Path)
}

// Precedence through the bridge: env beats file key per key (AC parity + precedence).
func TestFromApplicationConfig_EnvBeatsConfigFileKeys(t *testing.T) {
	resetConfig(t)
	c := config.Get()
	c.SetKey("db_type", "sqlite") // SetKey lowercases → flat key (D-1)
	c.SetKey("db_path", "/from/file.db")
	cfg, err := FromApplicationConfig()
	require.NoError(t, err)
	assert.Equal(t, DBTypeSQLite, cfg.Type)
	assert.Equal(t, "/from/file.db", cfg.Path)

	t.Setenv("DB_TYPE", "postgres")
	t.Setenv("DB_DSN", "postgres://env/wins")
	cfg, err = FromApplicationConfig()
	require.NoError(t, err)
	assert.Equal(t, DBTypePostgres, cfg.Type) // env overrode the file key
}
