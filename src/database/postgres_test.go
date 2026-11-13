package database

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// No real server needed: gorm.io/driver/postgres v1.6 dials during
// gorm.Open, so the factory tests use an unroutable port (instant refusal,
// no timeout wait); applyPool is engine-agnostic database/sql configuration
// and is proven offline through a memory-sqlite handle.

func TestNewDatabaseConnection_Postgres_UnreachableWrapsSentinel(t *testing.T) {
	ctx := testCtx()

	_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres, DSN: testDSN("u", "p", "127.0.0.1:1", "db")})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDatabaseUnreachable), "factory error must keep %w chain to sentinel")

	msg := err.Error()
	assert.Contains(t, msg, "u:***", "error must quote the REDACTED dsn")
	assert.NotContains(t, msg, "u:p@", "raw credential must never appear")
}

func TestNewDatabaseConnection_Postgres_MalformedDSNNoPanic(t *testing.T) {
	ctx := testCtx()

	assert.NotPanics(t, func() {
		_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres, DSN: "postgres://%zz%zz\\"})
		assert.Error(t, err) // pgx parse or ping failure — either error surface is acceptable
	})
}

func offlineHandle(t *testing.T) *gorm.DB {
	t.Helper()
	// applyPool only touches database/sql pool setters — engine-independent,
	// so a private memory-sqlite handle proves the pure helper offline.
	db, err := NewDatabaseConnection(testCtx(), DatabaseConfig{Type: DBTypeSQLite, Path: MemoryPath})
	require.NoError(t, err)
	return db
}

func TestApplyPool_DefaultsAndCustom(t *testing.T) {
	db := offlineHandle(t)
	defer Close(testCtx(), db)
	require.NoError(t, applyPool(db, DatabaseConfig{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	assert.Equal(t, DefaultPostgresMaxOpenConns, sqlDB.Stats().MaxOpenConnections)

	custom := DatabaseConfig{Postgres: PostgresPoolConfig{MaxOpenConns: 7, MaxIdleConns: 2, ConnMaxLifetime: 60_000_000_000}}
	require.NoError(t, applyPool(db, custom))
	assert.Equal(t, 7, sqlDB.Stats().MaxOpenConnections)
}

// Sanity that sqlite half still routes after the switch edit (no server needed).
func TestNewDatabaseConnection_RoutingIntact(t *testing.T) {
	ctx := testCtx()
	db, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypeSQLite, Path: filepath.Join(t.TempDir(), "r.db")})
	require.NoError(t, err)
	defer Close(ctx, db)
	assert.Equal(t, "sqlite", db.Dialector.Name())
}
