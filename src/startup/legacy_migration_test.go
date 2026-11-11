package startup

// Boot-hook units (task 06): seams stubbed so no ServiceProvider is built,
// no os.Exit path exists (running command is never api here), and no real
// filesystem conventions are needed — just the policy the hook encodes:
// migrate-once, dry-run-writes-nothing, failure-propagates-unexited.

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/database/legacy_migrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// stubHook points all three seams at a fresh :memory: handle + a temp copy
// of the golden fixture; wrap (optional) decorates RunMigration for failure
// injection. Returns the handle for assertions.
func stubHook(t *testing.T, wrap func(orig migratorFn) migratorFn) *gorm.DB {
	t.Helper()
	gdb, err := database.NewDatabaseConnection(basecontext.NewRootBaseContext(), database.DatabaseConfig{
		Type: database.DBTypeSQLite, Path: database.MemoryPath,
	})
	require.NoError(t, err)
	require.NoError(t, database.RunAutoMigrate(basecontext.NewRootBaseContext(), gdb))
	src := goldenCopy(t)

	oldGet, oldSrc, oldRun := getSQLDatabase, legacyStoreSource, runMigration
	t.Cleanup(func() { getSQLDatabase, legacyStoreSource, runMigration = oldGet, oldSrc, oldRun })
	getSQLDatabase = func(ctx basecontext.ApiContext) (*gorm.DB, error) { return gdb, nil }
	legacyStoreSource = func() (string, error) { return src, nil }
	if wrap != nil {
		runMigration = wrap(oldRun)
	}
	return gdb
}

func goldenCopy(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "database", "legacy_migrator", "testdata", "golden.json"))
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "data.json")
	require.NoError(t, os.WriteFile(path, raw, 0o600))
	return path
}

func counts(t *testing.T, db *gorm.DB) (users, marker int64) {
	t.Helper()
	require.NoError(t, db.Model(&models.User{}).Count(&users).Error)
	require.NoError(t, db.Table("legacy_migration_marker").Count(&marker).Error)
	return
}

func TestRunLegacyDataMigrationFreshStore(t *testing.T) {
	db := stubHook(t, nil)

	require.NoError(t, runLegacyDataMigration(basecontext.NewRootBaseContext()))

	users, marker := counts(t, db)
	assert.Equal(t, int64(3), users)
	assert.Equal(t, int64(1), marker)

	// second boot = idempotent skip (marker gate inside the migrator)
	require.NoError(t, runLegacyDataMigration(basecontext.NewRootBaseContext()))
	users2, marker2 := counts(t, db)
	assert.Equal(t, []int64{users, marker}, []int64{users2, marker2})
}

func TestRunLegacyDataMigrationDryRun(t *testing.T) {
	db := stubHook(t, nil)
	t.Setenv(constants.MIGRATE_DRY_RUN_ENV_VAR, "true")

	require.NoError(t, runLegacyDataMigration(basecontext.NewRootBaseContext()))

	users, marker := counts(t, db)
	assert.Zero(t, users, "dry-run wrote nothing")
	assert.Zero(t, marker, "dry-run left no marker")
}

func TestRunLegacyDataMigrationFailurePropagates(t *testing.T) {
	db := stubHook(t, func(orig migratorFn) migratorFn {
		return func(ctx basecontext.ApiContext, gdb *gorm.DB, opts legacy_migrator.Options) (legacy_migrator.Report, error) {
			opts.FailAfter = func(name string, done int) error {
				if name == "Roles" {
					return fmt.Errorf("synthetic boot-time crash")
				}
				return nil
			}
			return orig(ctx, gdb, opts)
		}
	})

	err := runLegacyDataMigration(basecontext.NewRootBaseContext())
	require.Error(t, err) // non-API posture: error returned, NOT os.Exit (call-site gate is review-verified)
	assert.Contains(t, err.Error(), "synthetic boot-time crash")

	users, marker := counts(t, db)
	assert.Zero(t, users, "hook propagated the rolled-back failure")
	assert.Zero(t, marker)
}

func TestDefaultLegacyStoreSourceHonoursDatabaseFolder(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(constants.DATABASE_FOLDER_ENV_VAR, dir)
	// provider.JsonDatabase is nil in this process (no provider init ran) ⇒
	// the convention branch resolves under DATABASE_FOLDER.
	path, err := defaultLegacyStoreSource()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "data.json"), path)
}
