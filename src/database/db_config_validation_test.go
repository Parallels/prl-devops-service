package database

import (
	"errors"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PBI #420 task 03 — actionable-message matrix (AC "Validation and errors").
// Pins the user-facing contract forever: every config error names the offending
// setting in BOTH notations (DB_* env and flat file key) plus allowed values.

func resetCfg(t *testing.T) {
	t.Helper()
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	_ = config.New(ctx)
}

func mustFailWith(t *testing.T, err error, wantIs error, mustContain ...string) {
	t.Helper()
	require.Error(t, err)
	if wantIs != nil {
		require.ErrorIs(t, err, wantIs)
	}
	for _, s := range mustContain {
		assert.Containsf(t, err.Error(), s, "message %q must name %q", err.Error(), s)
	}
}

func TestValidation_Matrix(t *testing.T) {
	t.Run("DB_TYPE=mysql names both notations + allowed values", func(t *testing.T) {
		resetCfg(t)
		t.Setenv("DB_TYPE", "mysql")
		_, err := FromApplicationConfig()
		mustFailWith(t, err, ErrUnsupportedDatabaseType, "DB_TYPE", "db_type", `"mysql"`, "sqlite", "postgres")
	})

	t.Run("4KB DB_TYPE is truncated in the echo (log-spam guard)", func(t *testing.T) {
		resetCfg(t)
		t.Setenv("DB_TYPE", strings.Repeat("x", 4096))
		_, err := FromApplicationConfig()
		require.Error(t, err)
		assert.Less(t, len(err.Error()), 200, "echo capped by quoteCapped")
		assert.Contains(t, err.Error(), "…")
		assert.NotContains(t, err.Error(), strings.Repeat("x", 4096))
	})

	t.Run("postgres + DSN unset", func(t *testing.T) {
		resetCfg(t)
		t.Setenv("DB_TYPE", "postgres")
		_, err := FromApplicationConfig()
		mustFailWith(t, err, ErrPostgresDSNRequired, "DB_DSN", "db_dsn")
	})

	t.Run("postgres + DSN blank-only", func(t *testing.T) {
		resetCfg(t)
		t.Setenv("DB_TYPE", "postgres")
		t.Setenv("DB_DSN", "   ") // bridge trims → treated as unset
		_, err := FromApplicationConfig()
		mustFailWith(t, err, ErrPostgresDSNRequired, "DB_DSN", "db_dsn")
	})

	t.Run("DB_AUTO_MIGRATE garbage", func(t *testing.T) {
		resetCfg(t)
		t.Setenv("DB_AUTO_MIGRATE", "maybe")
		_, err := FromApplicationConfig()
		mustFailWith(t, err, nil, "DB_AUTO_MIGRATE", "db_auto_migrate", "true", "false")
	})

	// Parity AC: file-key-only configuration must produce the SAME actionable text.
	t.Run("file-key-only db_type", func(t *testing.T) {
		resetCfg(t)
		config.Get().SetKey("db_type", "mysql")
		_, err := FromApplicationConfig()
		mustFailWith(t, err, ErrUnsupportedDatabaseType, "DB_TYPE", "db_type", `"mysql"`)
	})

	t.Run("file-key-only postgres needs DSN", func(t *testing.T) {
		resetCfg(t)
		config.Get().SetKey("db_type", "postgres")
		_, err := FromApplicationConfig()
		mustFailWith(t, err, ErrPostgresDSNRequired, "DB_DSN", "db_dsn")
	})

	t.Run("file-key-only auto_migrate", func(t *testing.T) {
		resetCfg(t)
		config.Get().SetKey("db_auto_migrate", "maybe")
		_, err := FromApplicationConfig()
		mustFailWith(t, err, nil, "DB_AUTO_MIGRATE", "db_auto_migrate")
	})
}

// sqlite + empty path is unreachable through the bridge (defaultDatabasePath
// always yields a path; DB_PATH only overrides when non-empty). Guard for
// future changes: the sentinel itself must reach callers UNWRAPPED via Validate.
func TestValidation_SQLitePathSentinelGuard(t *testing.T) {
	err := DatabaseConfig{Type: DBTypeSQLite, Path: ""}.Validate()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSQLitePathRequired))
	assert.Same(t, ErrSQLitePathRequired, err) // passthrough unwrapped
}
