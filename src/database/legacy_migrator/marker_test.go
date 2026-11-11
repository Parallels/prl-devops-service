package legacy_migrator

// Marker units (task 01) on #418's real :memory: factory (no GORM mocks,
// feature plan §12). Covers the plan-D-4 check ladder end to end, including
// the two paths that must log (hash-hit info, frozen-legacy warn) — those
// use ctx.EnableLogFile to observe output.

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func memoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := database.NewDatabaseConnection(basecontext.NewRootBaseContext(),
		database.DatabaseConfig{Type: database.DBTypeSQLite, Path: database.MemoryPath})
	require.NoError(t, err)
	t.Cleanup(func() {
		if sqlDB, e := conn.DB(); e == nil {
			_ = sqlDB.Close()
		}
	})
	return conn
}

func seedMarker(t *testing.T, db *gorm.DB, m LegacyMigrationMarker) {
	t.Helper()
	require.NoError(t, MarkMigrated(db, m))
}

// Package-wide log capture (ONE attachment for the whole test binary):
// EnableLogFile adds a handle to the GLOBAL logger — per-test attachments
// leak handles onto deleted temp files, and the dead handles drop later
// writes (observed: ChangedAfter only failed in full-suite runs). One
// TestMain attachment + per-call unique substrings is deterministic.
var sharedLogFile string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "legacy-migrator-logs-")
	if err != nil {
		panic(err)
	}
	sharedLogFile = filepath.Join(dir, "migrator.log")
	basecontext.NewRootBaseContext().EnableLogFile(sharedLogFile)
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func capturingCtx(t *testing.T) (basecontext.ApiContext, func() string) {
	t.Helper()
	return basecontext.NewRootBaseContext(), func() string {
		b, err := os.ReadFile(sharedLogFile)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

// waitForLog polls the capture file: the common-go file logger writes
// through a channel, so an immediate read can race the flush (observed as a
// one-line flake at first run in-container). Bounded wait, never a sleep
// that weakens the assertion.
func waitForLog(t *testing.T, read func() string, substr string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if strings.Contains(read(), substr) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("captured log never contained %q — got: %q", substr, read())
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func writeLegacy(t *testing.T, content []byte) (path string, det Detection) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "data.json")
	require.NoError(t, os.WriteFile(path, content, 0o600))
	d, err := DetectLegacyStore(basecontext.NewRootBaseContext(), path)
	require.NoError(t, err)
	require.True(t, d.Present)
	return path, d
}

func shaOf(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func TestEnsureMarkerTableIdempotent(t *testing.T) {
	db := memoryDB(t)
	require.NoError(t, EnsureMarkerTable(db))
	require.NoError(t, EnsureMarkerTable(db)) // second call is a no-op, not an error
	require.True(t, db.Migrator().HasTable(&LegacyMigrationMarker{}))
	assert.True(t, db.Migrator().HasTable("legacy_migration_marker"))
}

func TestAlreadyMigratedCleanSlate(t *testing.T) {
	db := memoryDB(t)
	require.NoError(t, EnsureMarkerTable(db))

	content := []byte(`{"users":[{"id":"u1"}]}`)
	path, det := writeLegacy(t, content)

	got, err := AlreadyMigrated(basecontext.NewRootBaseContext(), db, det, path, content)
	require.NoError(t, err)
	assert.False(t, got) // present + no marker rows at all ⇒ migrate it
}

func TestAlreadyMigratedAbsentIsTreatedAsDone(t *testing.T) {
	db := memoryDB(t)
	got, err := AlreadyMigrated(basecontext.NewRootBaseContext(), db, Detection{Reason: "absent"}, "/nope/data.json", nil)
	require.NoError(t, err)
	assert.True(t, got) // "don't run" — AC-1 skip folded into the gate
}

func TestAlreadyMigratedFastProbeNoContentRead(t *testing.T) {
	db := memoryDB(t)
	require.NoError(t, EnsureMarkerTable(db))
	content := []byte(`{"users":[{"id":"u1"}]}`)
	path, det := writeLegacy(t, content)

	seedMarker(t, db, LegacyMigrationMarker{
		SourcePath: path, SourceSizeBytes: det.SizeBytes,
		SourceModTimeUnix: det.ModTime.Unix(), ContentSHA256: "stale-but-present",
		CompletedAt: "2026-01-01 00:00:00",
	})

	// nil rawContent: the fast size+mtime probe must short-circuit BEFORE any
	// hashing, proving the happy path never reads data.json (R-3 accepted).
	got, err := AlreadyMigrated(basecontext.NewRootBaseContext(), db, det, path, nil)
	require.NoError(t, err)
	assert.True(t, got)
}

func TestAlreadyMigratedHashHitAfterMtimeTouch(t *testing.T) {
	db := memoryDB(t)
	require.NoError(t, EnsureMarkerTable(db))
	content := []byte(`{"users":[{"id":"u1"}]}`)
	path, _ := writeLegacy(t, content)

	seedMarker(t, db, LegacyMigrationMarker{
		SourcePath: "/somewhere/else/data.json", SourceSizeBytes: 1, SourceModTimeUnix: 1,
		ContentSHA256: shaOf(content), CompletedAt: "2026-01-01 00:00:00",
	})

	// Touch the file: mtime changes ⇒ fast probe misses; content is identical
	// ⇒ sha256 step must hit and log the info fingerprint line.
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, os.Chtimes(path, time.Now(), time.Now()))
	_, det2 := writeLegacySame(t, path)

	ctx, logs := capturingCtx(t)
	got, err := AlreadyMigrated(ctx, db, det2, path, nil) // nil ⇒ hash step reads the file
	require.NoError(t, err)
	assert.True(t, got)
	waitForLog(t, logs, "already migrated (fingerprint")
}

func writeLegacySame(t *testing.T, path string) (string, Detection) {
	t.Helper()
	d, err := DetectLegacyStore(basecontext.NewRootBaseContext(), path)
	require.NoError(t, err)
	require.True(t, d.Present)
	return path, d
}

func TestAlreadyMigratedChangedAfterMigration(t *testing.T) {
	db := memoryDB(t)
	require.NoError(t, EnsureMarkerTable(db)) // seed needs the table; AlreadyMigrated's own ensure runs later
	content := []byte(`{"users":[{"id":"u1"}]}`)
	path, det := writeLegacy(t, content)

	// A marker exists for a DIFFERENT file/content ⇒ frozen legacy changed
	// (OQ-1): warn and do NOT re-migrate.
	seedMarker(t, db, LegacyMigrationMarker{
		SourcePath: "/other/host/data.json", SourceSizeBytes: 999, SourceModTimeUnix: 1,
		ContentSHA256: "different", CompletedAt: "2026-01-01 00:00:00",
	})

	ctx, logs := capturingCtx(t)
	got, err := AlreadyMigrated(ctx, db, det, path, content)
	require.NoError(t, err)
	assert.True(t, got) // "don't run" — never re-migrate deltas
	waitForLog(t, logs, "changed after migration")
	waitForLog(t, logs, "not re-migrating")
}

func TestMarkMigratedUpsertKeepsOneRow(t *testing.T) {
	db := memoryDB(t)
	require.NoError(t, EnsureMarkerTable(db))

	sha := shaOf([]byte("payload"))
	require.NoError(t, MarkMigrated(db, LegacyMigrationMarker{
		SourcePath: "/a/data.json", SourceSizeBytes: 7, SourceModTimeUnix: 100,
		ContentSHA256: sha, CollectionsTotal: 16, CompletedAt: "2026-01-01 00:00:00",
	}))
	require.NoError(t, MarkMigrated(db, LegacyMigrationMarker{
		SourcePath: "/a/data.json", SourceSizeBytes: 7, SourceModTimeUnix: 100,
		ContentSHA256: sha, CollectionsTotal: 16, CompletedAt: "2026-02-02 00:00:00",
	}))

	var rows []LegacyMigrationMarker
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 1) // AC-2 upsert, not duplicate
	assert.Equal(t, "2026-02-02 00:00:00", rows[0].CompletedAt)
	assert.NotEmpty(t, rows[0].ID) // GenerateId() applied
}
