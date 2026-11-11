package testutil

import (
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	dsnSeq      atomic.Uint64
	dsnSanitize = regexp.MustCompile(`[^a-z0-9]`)
)

// memoryDSN implements the D-2 scheme: file:t425-<name>-<seq>?mode=memory&cache=shared
// Name sanitised to [a-z0-9-], capped at 48 chars; seq guarantees distinct
// DSNs even under -count=N (same t.Name() repeats).
func memoryDSN(t *testing.T) string {
	t.Helper()
	name := strings.ToLower(dsnSanitize.ReplaceAllString(t.Name(), "-"))
	if len(name) > 48 {
		name = name[:48]
	}
	return fmt.Sprintf("file:t425-%s-%d?mode=memory&cache=shared", name, dsnSeq.Add(1))
}

// OpenMemoryDB opens a uniquely-named in-memory SQLite DB through #418's
// factory. t.Cleanup closes it — a named shared-cache memory DB otherwise
// lives until every connection closes.
func OpenMemoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	rc, ctx := RecordingContext()
	db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
		Type: database.DBTypeSQLite,
		Path: memoryDSN(t),
	})
	require.NoErrorf(t, err, "open memory db failed: %s", strings.Join(rc.Lines(), "\n"))
	t.Cleanup(func() { database.Close(ctx, db) })
	return db
}

// MigrateSchema runs the #419 AutoMigrate entrypoint (database.RunAutoMigrate).
func MigrateSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	rc, ctx := RecordingContext()
	require.NoErrorf(t, database.RunAutoMigrate(ctx, db), "automigrate failed: %s", strings.Join(rc.Lines(), "\n"))
}

// WithCleanConfig isolates the config global (D-3 rule 3): fresh
// DATABASE_FOLDER + config.New(ctx), mirroring src/data/main_test.go.
// Uses t.Setenv ⇒ the calling test is sequential by construction.
func WithCleanConfig(t *testing.T) basecontext.ApiContext {
	t.Helper()
	t.Setenv("DATABASE_FOLDER", t.TempDir())
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	config.New(ctx)
	return ctx
}
