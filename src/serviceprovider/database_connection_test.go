package serviceprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetProviderConfig(t *testing.T) {
	t.Helper()
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	_ = config.New(ctx)
}

func providerTestCtx() basecontext.ApiContext {
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	return ctx
}

func TestInitGormConnection_HappyPath(t *testing.T) {
	resetProviderConfig(t)
	tmp := t.TempDir()
	t.Setenv("DATABASE_FOLDER", tmp)
	ctx := providerTestCtx()
	s := &ServiceProvider{}

	require.NoError(t, s.initGormConnection(ctx))
	assert.NotNil(t, s.SQLDatabase)
	assert.FileExists(t, filepath.Join(tmp, database.DefaultDatabaseFilename))
	database.Close(ctx, s.SQLDatabase)
}

func TestInitGormConnection_FailureLeavesFieldNil(t *testing.T) {
	resetProviderConfig(t)
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("file, not a dir"), 0o644))
	t.Setenv("DATABASE_FOLDER", filepath.Join(blocker, "sub"))
	ctx := providerTestCtx()
	s := &ServiceProvider{}

	err := s.initGormConnection(ctx)
	require.Error(t, err)
	assert.Nil(t, s.SQLDatabase, "failed init must not leave a half-owned handle")
}

// Method-level test: mustInit on a NON-api command (running command "" after
// config.New) must warn and return — the test surviving at all is the "no
// exit" assertion. The exit-1 branch is grep-audited (single os.Exit gated on
// API_COMMAND); it cannot be exercised without killing the test process.
func TestMustInitGormConnection_NonApiCommandContinues(t *testing.T) {
	resetProviderConfig(t)
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("file, not a dir"), 0o644))
	t.Setenv("DATABASE_FOLDER", filepath.Join(blocker, "sub"))
	ctx := providerTestCtx()
	s := &ServiceProvider{}

	assert.NotPanics(t, func() { s.mustInitGormConnection(ctx) })
	assert.Nil(t, s.SQLDatabase)
}

func TestGetSQLDatabase_MissingProviderAndEmptyProvider(t *testing.T) {
	ctx := providerTestCtx()
	orig := globalProvider
	defer func() { globalProvider = orig }()

	globalProvider = nil
	_, err := GetSQLDatabase(ctx)
	assert.ErrorIs(t, err, database.ErrDatabaseUnreachable)

	globalProvider = &ServiceProvider{}
	_, err = GetSQLDatabase(ctx)
	assert.ErrorIs(t, err, database.ErrDatabaseUnreachable)
}

func TestGetSQLDatabase_ReturnsStoredHandle(t *testing.T) {
	ctx := providerTestCtx()
	orig := globalProvider
	defer func() { globalProvider = orig }()

	db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{Type: database.DBTypeSQLite, Path: database.MemoryPath})
	require.NoError(t, err)
	defer database.Close(ctx, db)

	globalProvider = &ServiceProvider{SQLDatabase: db}
	got, err := GetSQLDatabase(ctx)
	require.NoError(t, err)
	assert.Same(t, db, got)
}
