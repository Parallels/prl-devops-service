package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseConnection_PostgresReturnsNotWired(t *testing.T) {
	ctx := testCtx()

	_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres, DSN: "host=whatever"})
	require.Error(t, err)
	// Explicit interim error — never a silent sqlite fallback.
	assert.ErrorIs(t, err, ErrPostgresNotWiredYet)
}

func TestNewDatabaseConnection_InvalidType(t *testing.T) {
	ctx := testCtx()

	_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: "mongodb", Path: "/tmp/x.db"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedDatabaseType)

	// Empty-postgres-DSN validation still precedes the switch:
	_, err = NewDatabaseConnection(ctx, DatabaseConfig{Type: "postgres", DSN: "  "})
	assert.ErrorIs(t, err, ErrPostgresDSNRequired)
}

func TestClose_NilSafe(t *testing.T) {
	ctx := testCtx()
	assert.NotPanics(t, func() { Close(ctx, nil) })
}
