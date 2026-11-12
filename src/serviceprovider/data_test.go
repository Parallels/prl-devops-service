package serviceprovider

// PBI #426 task 02 — runtime deprecation warns (D-3). Tests ship despite the
// loose DoD box: the AC ("deprecation warning logged at warn level on
// startup", OQ-1 frozen variant) is unprovable without them; string-EXACT
// assertions, recording fake capturing LogWarnf.

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/database/legacy_migrator"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// warns returns the captured warn-level lines, recording-prefix stripped —
// the exact emitted strings, in order.
func warns(rec *testutil.RecordingApiContext) []string {
	var out []string
	for _, l := range rec.Lines() {
		if s, ok := strings.CutPrefix(l, "warn: "); ok {
			out = append(out, s)
		}
	}
	return out
}

func TestLegacyWarns_SelectionOnly(t *testing.T) { // row a
	resetProviderConfig(t)
	t.Setenv("DB_TYPE", "legacy")
	rec, ctx := testutil.RecordingContext()
	s := &ServiceProvider{}
	path := filepath.Join(t.TempDir(), "test.json")

	s.selectDatabaseBackend(ctx, func() *data.JsonDatabase {
		return data.NewJsonDatabase(ctx, path)
	})
	defer func() { _ = s.Legacy().Disconnect(ctx) }()

	assert.Equal(t, []string{fmt.Sprintf(legacyDeprecationWarnMsg, "legacy")}, warns(rec),
		"legacy selection must emit exactly the pinned selection warn")
	// Version drift pin (D-6): the strings carry the deprecation/removal versions.
	assert.Contains(t, fmt.Sprint(warns(rec)), "v1.2.0")
	assert.Contains(t, fmt.Sprint(warns(rec)), "v1.3.0") // owner decision 2026-11-12 (D-6 amended)
}

func TestLegacyWarns_FrozenStore(t *testing.T) { // row b
	resetProviderConfig(t)
	t.Setenv("DB_TYPE", "legacy")
	rec, ctx := testutil.RecordingContext()
	s := &ServiceProvider{}
	path := filepath.Join(t.TempDir(), "test.json")

	db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
		Type: database.DBTypeSQLite, Path: database.MemoryPath})
	require.NoError(t, err)
	defer database.Close(ctx, db)
	s.SQLDatabase = db // injected: the frozen probe needs a reachable handle

	// Insert the completion marker through #424's exported API — never copied.
	require.NoError(t, legacy_migrator.EnsureMarkerTable(db))
	require.NoError(t, legacy_migrator.MarkMigrated(db, legacy_migrator.LegacyMigrationMarker{
		ID: "test-marker", SourcePath: path, CompletedAt: "2026-11-12T00:00:00Z",
	}))

	s.selectDatabaseBackend(ctx, func() *data.JsonDatabase {
		return data.NewJsonDatabase(ctx, path)
	})
	defer func() { _ = s.Legacy().Disconnect(ctx) }()

	assert.Equal(t, []string{
		fmt.Sprintf(legacyDeprecationWarnMsg, "legacy"),
		fmt.Sprintf(legacyFrozenWarnMsg, s.Legacy().Filename()),
	}, warns(rec), "frozen hazard: both warns, in order, exact strings")
}

func TestLegacyWarns_MarkerAbsent(t *testing.T) { // row c
	resetProviderConfig(t)
	t.Setenv("DB_TYPE", "legacy")
	rec, ctx := testutil.RecordingContext()
	s := &ServiceProvider{}
	path := filepath.Join(t.TempDir(), "test.json")

	db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
		Type: database.DBTypeSQLite, Path: database.MemoryPath})
	require.NoError(t, err)
	defer database.Close(ctx, db)
	s.SQLDatabase = db // reachable, migrated table ensured by the probe, zero rows

	s.selectDatabaseBackend(ctx, func() *data.JsonDatabase {
		return data.NewJsonDatabase(ctx, path)
	})
	defer func() { _ = s.Legacy().Disconnect(ctx) }()

	assert.Equal(t, []string{fmt.Sprintf(legacyDeprecationWarnMsg, "legacy")}, warns(rec),
		"migrated-table-with-zero-rows is not frozen — selection warn only")
}

func TestLegacyWarns_SqlModeSilent(t *testing.T) { // row d — positive-space proof
	resetProviderConfig(t)
	t.Setenv("DATABASE_FOLDER", t.TempDir()) // DB_TYPE unset ⇒ sqlite default
	rec, ctx := testutil.RecordingContext()
	s := &ServiceProvider{}

	s.selectDatabaseBackend(ctx, func() *data.JsonDatabase {
		t.Fatal("SQL mode must not build a legacy engine")
		return nil
	})
	defer func() { _ = s.JsonDatabase.Disconnect(ctx) }()

	assert.Empty(t, warns(rec), "SQL selection must emit zero legacy warns")
	assert.IsType(t, &database.GormDatabase{}, s.JsonDatabase)
}

func TestLegacyWarns_ProbeErrorDegradesNotFatal(t *testing.T) { // row e
	resetProviderConfig(t)
	t.Setenv("DB_TYPE", "legacy")
	rec, ctx := testutil.RecordingContext()
	s := &ServiceProvider{}
	path := filepath.Join(t.TempDir(), "test.json")

	db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
		Type: database.DBTypeSQLite, Path: database.MemoryPath})
	require.NoError(t, err)
	database.Close(ctx, db) // detached handle ⇒ probe fails
	s.SQLDatabase = db

	require.NotPanics(t, func() {
		s.selectDatabaseBackend(ctx, func() *data.JsonDatabase {
			return data.NewJsonDatabase(ctx, path)
		})
	})
	defer func() { _ = s.Legacy().Disconnect(ctx) }()

	w := warns(rec)
	require.Len(t, w, 2, "deprecation warn + exactly one probe-fail warn")
	assert.Equal(t, fmt.Sprintf(legacyDeprecationWarnMsg, "legacy"), w[0])
	assert.True(t, strings.HasPrefix(w[1], legacyProbeFailWarnMsg[:len(legacyProbeFailWarnMsg)-3]),
		"second warn is the probe-failure line: %q", w[1])
	// Selection still succeeded on legacy — the error was never fatal:
	assert.Same(t, s.Legacy(), s.JsonDatabase)
	assert.NotNil(t, s.Legacy())
}
