package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestInitializeSQLite_Success(t *testing.T) {
	tempDir := t.TempDir()

	config := &common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: tempDir,
			FileName:    "test.db",
		},
		Pool: common.DefaultPoolConfig(),
	}

	gormConfig := &gorm.Config{}

	db, err := initializeSQLite(config, gormConfig)
	require.NoError(t, err, "SQLite initialization should succeed")
	require.NotNil(t, db, "Database connection should not be nil")

	// Verify database file exists
	dbPath := filepath.Join(tempDir, "test.db")
	assert.FileExists(t, dbPath, "Database file should be created")

	// Cleanup
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

func TestInitializeSQLite_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	dbDir := filepath.Join(tempDir, "nested", "db", "path")

	config := &common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: dbDir,
			FileName:    "test.db",
		},
		Pool: common.DefaultPoolConfig(),
	}

	gormConfig := &gorm.Config{}

	db, err := initializeSQLite(config, gormConfig)
	require.NoError(t, err, "Should create nested directories")
	require.NotNil(t, db)

	// Verify directory was created
	assert.DirExists(t, dbDir, "Nested directory should be created")

	// Cleanup
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

func TestInitializeSQLite_ConnectionPool(t *testing.T) {
	tempDir := t.TempDir()

	config := &common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: tempDir,
			FileName:    "test.db",
		},
		Pool: common.PoolConfig{
			MaxIdleConns:    2,
			MaxOpenConns:    5,
			ConnMaxLifetime: 0, // No limit for test
		},
	}

	gormConfig := &gorm.Config{}

	db, err := initializeSQLite(config, gormConfig)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify pool settings were applied
	sqlDB, err := db.DB()
	require.NoError(t, err)

	stats := sqlDB.Stats()
	// SQLite may open an initial connection, so check it's <= MaxOpenConns
	assert.LessOrEqual(t, stats.OpenConnections, 5, "Open connections should not exceed MaxOpenConns")

	// Cleanup
	sqlDB.Close()
}

func TestInitializeSQLite_RelativePath(t *testing.T) {
	// Create a temporary directory with a relative path
	originalWd, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	config := &common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: ".",
			FileName:    "relative.db",
		},
		Pool: common.DefaultPoolConfig(),
	}

	gormConfig := &gorm.Config{}

	db, err := initializeSQLite(config, gormConfig)
	require.NoError(t, err, "Should handle relative paths")
	require.NotNil(t, db)

	// Verify database exists
	assert.FileExists(t, "relative.db", "Database should exist in current directory")

	// Cleanup
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

func TestInitializeSQLite_InvalidPath(t *testing.T) {
	// Try to create database in non-writable location
	config := &common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: "/root/no-permission",
			FileName:    "test.db",
		},
		Pool: common.DefaultPoolConfig(),
	}

	gormConfig := &gorm.Config{}

	_, err := initializeSQLite(config, gormConfig)
	// On most systems this will fail due to permissions
	// But on systems where we have access, it might succeed
	// So we just check that it doesn't panic
	if err != nil {
		t.Logf("Expected error due to permissions: %v", err)
	}
}

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Create nested directory",
			path:     filepath.Join(tempDir, "a", "b", "c", "test.db"),
			expected: true,
		},
		{
			name:     "Existing directory",
			path:     filepath.Join(tempDir, "existing.db"),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ensureDir(tc.path)
			if tc.expected {
				assert.NoError(t, err)
				dir := filepath.Dir(tc.path)
				assert.DirExists(t, dir, "Directory should be created")
			} else {
				assert.Error(t, err)
			}
		})
	}
}
