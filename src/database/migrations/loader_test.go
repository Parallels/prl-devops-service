package migrations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoadMigrationsFromPath(t *testing.T) {
	// Setup temporary database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Create service
	svc := NewMigrationService(db)

	// Create temporary migration files
	tmpDir := t.TempDir()

	upContent := "CREATE TABLE test (id INT);"
	downContent := "DROP TABLE test;"

	err = os.WriteFile(filepath.Join(tmpDir, "0001_test_migration.up.sql"), []byte(upContent), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "0001_test_migration.down.sql"), []byte(downContent), 0644)
	require.NoError(t, err)

	// Load migrations
	err = svc.LoadMigrationsFromPath(tmpDir)
	assert.NoError(t, err)

	// Verify worker registered
	registered := svc.GetRegisteredSeeds()
	assert.Contains(t, registered, "test_migration")

	// Prepare for running
	// We need to verify that it actually runs
	// But first let's just check registration

	// Verify that we can find the worker
	worker := svc.findWorkerByName("test_migration")
	require.NotNil(t, worker)

	// Verify version
	assert.Equal(t, 1, worker.GetVersion())
}
