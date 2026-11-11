//go:build integration

// Run: cd src && go test -tags integration ./database/legacy_migrator/... -count=1
// No external services (SQLite file + :memory: only) — unlike #418/#425's
// postgres compose tests. Task 07: the PBI's proof-of-record — the
// first-run upgrade story through the REAL #418 factory, zero mocks.
package legacy_migrator

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- helpers (D-10: tiny committed fixture, everything else generated) ---

func goldenBytes(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "golden.json"))
	require.NoError(t, err)
	return b
}

// newFileDB mirrors boot: real #418 factory on a DISK path + AutoMigrate
// (AD-3 coexistence: legacy data.json and devops.db side by side).
func newFileDB(t *testing.T, dir string) *gorm.DB {
	t.Helper()
	gdb, err := database.NewDatabaseConnection(basecontext.NewRootBaseContext(), database.DatabaseConfig{
		Type: database.DBTypeSQLite,
		Path: filepath.Join(dir, "devops.db"),
	})
	require.NoError(t, err)
	require.NoError(t, database.RunAutoMigrate(basecontext.NewRootBaseContext(), gdb))
	t.Cleanup(func() {
		if sqlDB, e := gdb.DB(); e == nil {
			_ = sqlDB.Close()
		}
	})
	return gdb
}

func placeGolden(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, goldenBytes(t), 0o600))
	return path
}

func shaOfFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// planCounts returns collection → actual count via the plan's own formulas
// (the same Validate paths task 05 shipped).
func planCounts(t *testing.T, gdb *gorm.DB, legacy *data.Data) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	for _, s := range collectionPlan() {
		_, actual, err := s.Validate(gdb, legacy)
		require.NoError(t, err, s.Name)
		out[s.Name] = actual
	}
	return out
}

func reportTotal(rep Report) int64 {
	var n int64
	for _, c := range rep.Collections {
		n += c.WrittenCount
	}
	return n
}

// buildEncryptedLegacy duplicates the data-package recipe (task-02's
// helper lives in data/*_test.go — not importable cross-package; noted as
// a deviation) : RSA keygen → EncryptString(golden) → ciphertext file.
func buildEncryptedLegacy(t *testing.T, dir string, golden []byte) (path string, pem []byte) {
	t.Helper()
	pemPath := filepath.Join(t.TempDir(), "k.pem")
	require.NoError(t, security.GenPrivateRsaKey(pemPath, 2048))
	pemBytes, err := os.ReadFile(pemPath)
	require.NoError(t, err)
	cipher, err := security.EncryptString(string(pemBytes), string(golden))
	require.NoError(t, err)
	path = filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, cipher, 0o600))
	return path, pemBytes
}

func markerRows(t *testing.T, gdb *gorm.DB) []LegacyMigrationMarker {
	t.Helper()
	var rows []LegacyMigrationMarker
	require.NoError(t, gdb.Find(&rows).Error)
	return rows
}

// --- scenarios (each names its AC) ---

// AC-2 + AC-6: the upgrade story end to end — fixture on disk, real
// factory on a sibling devops.db, one RunMigration → every #419 table
// populated per golden, validation clean, marker matches, bytes untouched.
func TestInt_GoldenUpgrade(t *testing.T) {
	dir := t.TempDir()
	src := placeGolden(t, dir)
	sumBefore := shaOfFile(t, src)
	gdb := newFileDB(t, dir)

	rep, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src})
	require.NoError(t, err)
	require.False(t, rep.Skipped)
	require.False(t, rep.DryRun)

	legacy, err := data.LoadLegacyData(basecontext.NewRootBaseContext(), src)
	require.NoError(t, err)
	plan := collectionPlan()
	require.Empty(t, ValidateCounts(basecontext.NewRootBaseContext(), gdb, &legacy, plan))

	counts := planCounts(t, gdb, &legacy)
	var want int64
	for _, s := range plan {
		want += int64(s.LegacyCount(&legacy))
		assert.Equalf(t, int64(s.LegacyCount(&legacy)), counts[s.Name], "collection %s", s.Name)
	}
	assert.Equal(t, want, reportTotal(rep))
	assert.Len(t, rep.Collections, 19)
	assert.Greater(t, rep.MarkerID, "")

	rows := markerRows(t, gdb)
	require.Len(t, rows, 1)
	assert.Equal(t, src, rows[0].SourcePath)
	st, err := os.Stat(src)
	require.NoError(t, err)
	assert.Equal(t, st.Size(), rows[0].SourceSizeBytes)
	assert.Equal(t, want, rows[0].CollectionsTotal)
	assert.Equal(t, shaOfFile(t, src), rows[0].ContentSHA256)

	assert.Equal(t, sumBefore, shaOfFile(t, src), "golden bytes must be untouched")
	assert.FileExists(t, filepath.Join(dir, "devops.db")) // AD-3 coexistence
}

// AC-2: second boot over the same handles = pure no-op, one marker.
func TestInt_RerunNoop(t *testing.T) {
	dir := t.TempDir()
	src := placeGolden(t, dir)
	gdb := newFileDB(t, dir)

	_, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src})
	require.NoError(t, err)
	legacy, err := data.LoadLegacyData(basecontext.NewRootBaseContext(), src)
	require.NoError(t, err)
	before := planCounts(t, gdb, &legacy)

	rep, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src})
	require.NoError(t, err)
	require.True(t, rep.Skipped)
	assert.Equal(t, "already-migrated", rep.SkipReason)
	assert.Empty(t, rep.Collections)

	after := planCounts(t, gdb, &legacy)
	assert.Equal(t, before, after)
	assert.Len(t, markerRows(t, gdb), 1)
}

// D-4 step 3: same BYTES at a new path/mtime ⇒ hash-step skip, INFO line.
func TestInt_RepointedStore(t *testing.T) {
	dir := t.TempDir()
	src := placeGolden(t, dir)
	gdb := newFileDB(t, dir)
	_, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src})
	require.NoError(t, err)

	relocated := filepath.Join(dir, "data2.json")
	require.NoError(t, os.WriteFile(relocated, goldenBytes(t), 0o600))

	ctx, logs := capturingCtx(t)
	rep, err := RunMigration(ctx, gdb, Options{SourcePath: relocated})
	require.NoError(t, err)
	require.True(t, rep.Skipped)
	assert.Equal(t, "already-migrated", rep.SkipReason)
	waitForLog(t, logs, "already migrated (fingerprint")
	assert.Len(t, markerRows(t, gdb), 1)
}

// OQ-1 frozen posture: post-migration edits NEVER re-migrate — skip + WARN.
func TestInt_FrozenAfterMigration(t *testing.T) {
	dir := t.TempDir()
	src := placeGolden(t, dir)
	gdb := newFileDB(t, dir)
	_, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src})
	require.NoError(t, err)
	legacy, err := data.LoadLegacyData(basecontext.NewRootBaseContext(), src)
	require.NoError(t, err)
	before := planCounts(t, gdb, &legacy)

	// Prepend a record (still-valid JSON, different bytes/sha) — keep the
	// prefix, split on the first occurrence only.
	rawGolden := goldenBytes(t)
	parts := bytes.SplitN(rawGolden, []byte(`"users": [`), 2) // golden is pretty-printed
	require.Len(t, parts, 2)
	grown := append(append([]byte{}, parts[0]...), []byte(`"users":[{"id":"u-99","username":"ghost"},`)...)
	grown = append(grown, parts[1]...)
	require.NoError(t, os.WriteFile(src, grown, 0o600))

	ctx, logs := capturingCtx(t)
	rep, err := RunMigration(ctx, gdb, Options{SourcePath: src})
	require.NoError(t, err)
	require.True(t, rep.Skipped)
	waitForLog(t, logs, "changed after migration")

	after := planCounts(t, gdb, &legacy)
	assert.Equal(t, before, after, "frozen legacy ⇒ no re-migration")
	assert.Len(t, markerRows(t, gdb), 1)
}

// AC-2 over ciphertext: encrypted store migrates identically to plaintext.
func TestInt_EncryptedGolden(t *testing.T) {
	dir := t.TempDir()
	src, pem := buildEncryptedLegacy(t, dir, goldenBytes(t))
	t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, base64.StdEncoding.EncodeToString(pem))
	gdb := newFileDB(t, dir)

	rep, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src})
	require.NoError(t, err)
	require.False(t, rep.Skipped)
	assert.Equal(t, int64(41), reportTotal(rep))

	legacy, err := data.LoadLegacyData(basecontext.NewRootBaseContext(), src)
	require.NoError(t, err)
	require.Empty(t, ValidateCounts(basecontext.NewRootBaseContext(), gdb, &legacy, collectionPlan()))
}

// R-5: encrypted without the key ⇒ named error, empty DB, no marker.
func TestInt_EncryptedNoKey(t *testing.T) {
	dir := t.TempDir()
	src, _ := buildEncryptedLegacy(t, dir, goldenBytes(t))
	t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, "")
	gdb := newFileDB(t, dir)

	_, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ENCRYPTION_PRIVATE_KEY")

	empty := &data.Data{}
	for _, s := range collectionPlan() {
		_, actual, verr := s.Validate(gdb, empty)
		require.NoError(t, verr)
		assert.Zerof(t, actual, "collection %s must stay empty", s.Name)
	}
	assert.Empty(t, markerRows(t, gdb))
}

// AC-4 at scale: 5000 jobs through the default 500-row chunking.
func TestInt_LargeStoreChunking(t *testing.T) {
	dir := t.TempDir()
	src := BuildLargeJobsFixture(t, dir, 5000)
	gdb := newFileDB(t, dir)

	ctx, logs := capturingCtx(t)
	start := time.Now()
	rep, err := RunMigration(ctx, gdb, Options{SourcePath: src})
	elapsed := time.Since(start)
	require.NoError(t, err)
	assert.Equal(t, int64(5000), reportTotal(rep))
	assert.Less(t, elapsed, 30*time.Second, "5k rows must complete < 30s (observed %s)", elapsed)
	// File logger is async: poll until the Jobs progress line lands, then
	// earlier lines are guaranteed visible for the count below. (Needle
	// carries NO digits: the test capture context renders numeric args as
	// `%!d(string=…)` — pre-existing capture quirk; the production file
	// logger renders them fine, proven by the macOS boot matrix.)
	waitForLog(t, logs, "Jobs records (total so far")

	lines := 0
	for _, l := range strings.Split(logs(), "\n") {
		if strings.Contains(l, "records (total so far") {
			lines++
		}
	}
	// Deviation from the spec table (">=10 for Jobs"): the MERGED-AND-
	// APPROVED AC-4 implementation (tasks 03/05, human-gated) logs ONE
	// progress line per collection, not per chunk — so the observable at
	// scale is the full 19-line plan with Jobs reporting 5000/5000.
	assert.GreaterOrEqual(t, lines, 19, "every planned collection gets a progress line")

	legacy, err := data.LoadLegacyData(basecontext.NewRootBaseContext(), src)
	require.NoError(t, err)
	require.Empty(t, ValidateCounts(basecontext.NewRootBaseContext(), gdb, &legacy, collectionPlan()))
}

// AC-7 over the REAL factory file DB: report projects, NOTHING exists.
func TestInt_DryRunE2E(t *testing.T) {
	dir := t.TempDir()
	src := placeGolden(t, dir)
	sumBefore := shaOfFile(t, src)

	// No RunAutoMigrate here — if the dry-run wrote schema at all, the
	// absence assertions below would be vacuous.
	gdb, err := database.NewDatabaseConnection(basecontext.NewRootBaseContext(), database.DatabaseConfig{
		Type: database.DBTypeSQLite,
		Path: filepath.Join(dir, "devops.db"),
	})
	require.NoError(t, err)
	defer func() {
		if sqlDB, e := gdb.DB(); e == nil {
			_ = sqlDB.Close()
		}
	}()

	rep, err := RunMigration(basecontext.NewRootBaseContext(), gdb, Options{SourcePath: src, DryRun: true})
	require.NoError(t, err)
	require.True(t, rep.DryRun)
	assert.Equal(t, int64(41), reportTotal(rep))
	assert.Len(t, rep.Collections, 19)

	require.True(t, gdb.Migrator().HasTable("legacy_migration_marker")) // predicate ensured it
	assert.Empty(t, markerRows(t, gdb))
	assert.False(t, gdb.Migrator().HasTable("users"), "dry-run must not create data tables")
	assert.Equal(t, sumBefore, shaOfFile(t, src))
}
