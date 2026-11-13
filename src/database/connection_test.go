package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseConnection_PostgresRoutesToDial(t *testing.T) {
	// Task 03's "not wired yet" interim test retired by task 04: the
	// postgres case now dials. Offline proof of real dialing lives in
	// postgres_test.go (unreachable host → ErrDatabaseUnreachable wrap).
	ctx := testCtx()

	_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: "postgres", DSN: testDSN("u", "p", "127.0.0.1:1", "db")})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDatabaseUnreachable)
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
