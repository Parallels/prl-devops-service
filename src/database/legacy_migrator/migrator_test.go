package legacy_migrator

// Engine units (task 03): the AC matrix on a real :memory: handle —
// idempotency, single-tx rollback, progress lines, marker content,
// dry-run semantics, chunking. Fixtures from task 02's testdata (golden);
// large volumes generated at test time (D-10). Logs are observed through a
// recording ctx wrapper (spec-mandated) instead of the global file logger.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type recordingCtx struct {
	basecontext.ApiContext
	mu    sync.Mutex
	lines []string
}

func newRecordingCtx() *recordingCtx {
	return &recordingCtx{ApiContext: basecontext.NewRootBaseContext()}
}

func (r *recordingCtx) record(level, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines = append(r.lines, level+" "+msg)
}
func (r *recordingCtx) LogInfof(f string, a ...interface{})  { r.record("INFO", fmt.Sprintf(f, a...)) }
func (r *recordingCtx) LogWarnf(f string, a ...interface{})  { r.record("WARN", fmt.Sprintf(f, a...)) }
func (r *recordingCtx) LogDebugf(f string, a ...interface{}) { r.record("DEBUG", fmt.Sprintf(f, a...)) }
func (r *recordingCtx) LogErrorf(f string, a ...interface{}) { r.record("ERROR", fmt.Sprintf(f, a...)) }
func (r *recordingCtx) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.lines...)
}

func (r *recordingCtx) count(pred func(string) bool) int {
	n := 0
	for _, l := range r.all() {
		if pred(l) {
			n++
		}
	}
	return n
}

func migDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := memoryDB(t) // task-01 helper: real #418 :memory: factory + cleanup
	require.NoError(t, database.RunAutoMigrate(basecontext.NewRootBaseContext(), db))
	require.NoError(t, EnsureMarkerTable(db))
	return db
}

// goldenAt copies the golden fixture into a temp dir so the SourcePath is a
// private, deletable file (marker stores it verbatim).
func goldenAt(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "golden.json"))
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "data.json")
	require.NoError(t, os.WriteFile(path, raw, 0o600))
	return path
}

func countModel(t *testing.T, db *gorm.DB, model interface{}) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(model).Count(&n).Error)
	return n
}

// scalarCounts is the expected golden outcome: 10 collections, 12 records.
func scalarCounts() map[string]int64 {
	return map[string]int64{
		"Configuration": 1, "ApiKeys": 2, "PackerTemplates": 1,
		"OrchestratorHosts": 1, "ReverseProxy": 1, "ReverseProxyHosts": 1,
		"CatalogManagers": 1, "Jobs": 2, "EnrollmentTokens": 1, "UserConfigs": 1,
	}
}

func countAllScalars(t *testing.T, db *gorm.DB) map[string]int64 {
	t.Helper()
	return map[string]int64{
		"Configuration":     countModel(t, db, &models.Configuration{}),
		"ApiKeys":           countModel(t, db, &models.ApiKey{}),
		"PackerTemplates":   countModel(t, db, &models.PackerTemplate{}),
		"OrchestratorHosts": countModel(t, db, &models.OrchestratorHost{}),
		"ReverseProxy":      countModel(t, db, &models.ReverseProxy{}),
		"ReverseProxyHosts": countModel(t, db, &models.ReverseProxyHost{}),
		"CatalogManagers":   countModel(t, db, &models.CatalogManager{}),
		"Jobs":              countModel(t, db, &models.Job{}),
		"EnrollmentTokens":  countModel(t, db, &models.OrchestratorEnrollmentToken{}),
		"UserConfigs":       countModel(t, db, &models.UserConfig{}),
		"__marker":          countModel(t, db, &LegacyMigrationMarker{}),
	}
}

func TestRunMigrationGoldenScalars(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)

	rep, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{SourcePath: path})
	require.NoError(t, err)
	assert.False(t, rep.Skipped)
	assert.NotEmpty(t, rep.MarkerID)

	require.Len(t, rep.Collections, 10)
	for _, c := range rep.Collections {
		assert.Equal(t, scalarCounts()[c.Name], int64(c.LegacyCount), "legacy count for %s", c.Name)
		assert.Equal(t, scalarCounts()[c.Name], c.WrittenCount, "written count for %s", c.Name)
	}
	got := countAllScalars(t, db)
	for name, want := range scalarCounts() {
		assert.Equal(t, want, got[name], "table count for %s", name)
	}
	assert.Equal(t, int64(1), got["__marker"])
}

func TestRunMigrationIdempotent(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	ctx := basecontext.NewRootBaseContext()

	_, err := RunMigration(ctx, db, Options{SourcePath: path})
	require.NoError(t, err)
	first := countAllScalars(t, db)

	rep2, err := RunMigration(ctx, db, Options{SourcePath: path})
	require.NoError(t, err)
	assert.True(t, rep2.Skipped)
	assert.Equal(t, "already-migrated", rep2.SkipReason)
	assert.Equal(t, first, countAllScalars(t, db), "second run changed nothing")

	// The upsert engine itself, independent of the marker gate (AC-2 core):
	// replaying every scalar spec a second time must neither duplicate nor
	// error.
	legacy, err := data.LoadLegacyData(ctx, path)
	require.NoError(t, err)
	for _, s := range collectionPlan() {
		_, err := s.Migrate(ctx, db, &legacy, defaultChunkSize)
		require.NoError(t, err, s.Name)
		// RowsAffected on update-in-place is dialect-flavoured (sqlite
		// reports 0 for identical rewrites — see upsertAll caveat); the
		// AC-2 guarantee is the unchanged table below, not the counter.
	}
	assert.Equal(t, first, countAllScalars(t, db), "replay changed nothing")
}

func TestRunMigrationCrashRollback(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)

	rep, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{
		SourcePath: path,
		FailAfter: func(name string, done int) error {
			if name == "PackerTemplates" { // second collection
				return fmt.Errorf("synthetic crash after %d rows", done)
			}
			return nil
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "synthetic crash")
	assert.Empty(t, rep.Collections, "failed run reports nothing")

	got := countAllScalars(t, db)
	for name, n := range got {
		assert.Zero(t, n, "%s must be empty after rollback", name) // includes __marker (AC-3+AC-5)
	}
}

func TestRunMigrationProgressLogs(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	ctx := newRecordingCtx()

	_, err := RunMigration(ctx, db, Options{SourcePath: path})
	require.NoError(t, err)

	progress := ctx.count(func(l string) bool { return strings.HasPrefix(l, "INFO migrated ") })
	assert.Equal(t, len(collectionPlan()), progress, "one AC-4 line per collection")
	assert.Contains(t, strings.Join(ctx.all(), "\n"), "migrated 1/12 Configuration records (total so far 1/12")
	assert.Contains(t, strings.Join(ctx.all(), "\n"), "migrated 2/12 Jobs records (total so far 10/12")
}

func TestRunMigrationMarkerContent(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)

	rep, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{SourcePath: path})
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	sum := sha256.Sum256(raw)
	wantSha := hex.EncodeToString(sum[:])

	var rows []LegacyMigrationMarker
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 1)
	m := rows[0]
	assert.Equal(t, path, m.SourcePath)
	assert.Equal(t, info.Size(), m.SourceSizeBytes)
	assert.Equal(t, info.ModTime().Unix(), m.SourceModTimeUnix)
	assert.Equal(t, wantSha, m.ContentSHA256)
	assert.Equal(t, int64(12), m.CollectionsTotal)
	assert.NotEmpty(t, m.CompletedAt)
	assert.Equal(t, rep.MarkerID, m.ID)
}

func TestRunMigrationDryRun(t *testing.T) {
	db := migDB(t)
	path := goldenAt(t)
	before, err := os.ReadFile(path)
	require.NoError(t, err)
	ctx := newRecordingCtx()

	rep, err := RunMigration(ctx, db, Options{SourcePath: path, DryRun: true})
	require.NoError(t, err)
	assert.True(t, rep.DryRun)
	assert.False(t, rep.Skipped)
	require.Len(t, rep.Collections, 10)
	for _, c := range rep.Collections {
		assert.Equal(t, int64(c.LegacyCount), c.WrittenCount, "projection count for %s", c.Name)
	}

	got := countAllScalars(t, db)
	for name, n := range got {
		assert.Zero(t, n, "%s must stay empty in dry-run", name)
	}
	assert.Empty(t, rep.MarkerID)

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, before, after, "source file untouched")
	assert.Equal(t, 1, ctx.count(func(l string) bool { return strings.Contains(l, "nothing written") }))
}

func TestRunMigrationChunking(t *testing.T) {
	db := migDB(t)
	var creates int64
	require.NoError(t, db.Callback().Create().After("gorm:create").
		Register("test:create_count", func(tx *gorm.DB) { atomic.AddInt64(&creates, 1) }))

	path := BuildLargeJobsFixture(t, t.TempDir(), 5000)
	rep, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{SourcePath: path, ChunkSize: 500})
	require.NoError(t, err)

	assert.Equal(t, int64(5000), countModel(t, db, &models.Job{}))
	require.Len(t, rep.Collections, len(collectionPlan())) // every collection reports, zero-row ones included
	var jobsRep CollectionReport
	for _, c := range rep.Collections {
		if c.Name == "Jobs" {
			jobsRep = c
		} else {
			assert.Zero(t, c.WrittenCount, c.Name)
		}
	}
	assert.Equal(t, int64(5000), jobsRep.WrittenCount)
	// 5000/500 = 10 job batches + marker create; nothing else had rows.
	assert.GreaterOrEqual(t, atomic.LoadInt64(&creates), int64(11), "chunking actually batched")
}

func TestRunMigrationNoSourcePath(t *testing.T) {
	db := migDB(t)
	require.NotPanics(t, func() {
		_, err := RunMigration(basecontext.NewRootBaseContext(), db, Options{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "SourcePath is required")
	})
}
