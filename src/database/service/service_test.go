package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseService_Initialize_SQLite(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()

	config := common.Config{
		Type:    common.SQLite,
		Debug:   false,
		Migrate: true,
		SQLite: common.SQLiteConfig{
			StoragePath: tempDir,
			FileName:    "test.db",
		},
		Pool: common.PoolConfig{
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: time.Hour,
		},
	}

	// Reset singleton for test
	Reset()

	// Initialize database service
	dbSvc, err := Initialize(config)
	require.NoError(t, err, "Failed to initialize database service")
	require.NotNil(t, dbSvc, "Database service should not be nil")

	// Verify database file was created
	dbPath := filepath.Join(tempDir, "test.db")
	assert.FileExists(t, dbPath, "Database file should exist")

	// Verify service properties
	assert.Equal(t, "database", dbSvc.Name())
	assert.True(t, dbSvc.IsEnabled())
	assert.Empty(t, dbSvc.Dependencies())

	// Cleanup
	dbSvc.Close()
}

func TestDatabaseService_Health(t *testing.T) {
	tempDir := t.TempDir()

	config := common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: tempDir,
			FileName:    "test.db",
		},
		Pool: common.DefaultPoolConfig(),
	}

	Reset()
	dbSvc, err := Initialize(config)
	require.NoError(t, err)
	defer dbSvc.Close()

	// Health check should pass
	err = dbSvc.Health(context.Background())
	assert.NoError(t, err, "Health check should pass for active connection")
}

func TestDatabaseService_GetDB(t *testing.T) {
	tempDir := t.TempDir()

	config := common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: tempDir,
			FileName:    "test.db",
		},
		Pool: common.DefaultPoolConfig(),
	}

	Reset()
	dbSvc, err := Initialize(config)
	require.NoError(t, err)
	defer dbSvc.Close()

	// GetDB should return valid connection
	db := dbSvc.GetDB()
	assert.NotNil(t, db, "GetDB should return valid connection")
}

func TestDatabaseService_Singleton(t *testing.T) {
	tempDir := t.TempDir()

	config := common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: tempDir,
			FileName:    "test.db",
		},
		Pool: common.DefaultPoolConfig(),
	}

	Reset()

	// First initialization
	dbSvc1, err := Initialize(config)
	require.NoError(t, err)
	defer dbSvc1.Close()

	// Second initialization should return same instance
	dbSvc2, err := Initialize(config)
	require.NoError(t, err)

	assert.Same(t, dbSvc1, dbSvc2, "Initialize should return same singleton instance")
	assert.Same(t, dbSvc1, GetInstance(), "GetInstance should return same instance")
}

func TestDatabaseService_InvalidConfig(t *testing.T) {
	Reset()

	// Invalid database type
	config := common.Config{
		Type: "invalid",
		Pool: common.DefaultPoolConfig(),
	}

	_, err := Initialize(config)
	assert.Error(t, err, "Should fail with invalid database type")
}

func TestDatabaseService_Close(t *testing.T) {
	tempDir := t.TempDir()

	config := common.Config{
		Type: common.SQLite,
		SQLite: common.SQLiteConfig{
			StoragePath: tempDir,
			FileName:    "test.db",
		},
		Pool: common.DefaultPoolConfig(),
	}

	Reset()
	dbSvc, err := Initialize(config)
	require.NoError(t, err)

	// Close should succeed
	err = dbSvc.Close()
	assert.NoError(t, err, "Close should succeed")

	// Health check should fail after close
	err = dbSvc.Health(context.Background())
	assert.Error(t, err, "Health check should fail after close")
}

func TestDatabaseService_ConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      common.Config
		expectError bool
	}{
		{
			name: "Valid SQLite config",
			config: common.Config{
				Type: common.SQLite,
				SQLite: common.SQLiteConfig{
					StoragePath: os.TempDir(),
					FileName:    "test.db",
				},
				Pool: common.DefaultPoolConfig(),
			},
			expectError: false,
		},
		{
			name: "Empty database type",
			config: common.Config{
				Type: "",
				Pool: common.DefaultPoolConfig(),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Reset()
			_, err := Initialize(tc.config)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				if err == nil {
					// Cleanup if successful
					if svc := GetInstance(); svc != nil {
						svc.Close()
					}
				}
			}
		})
	}
}
