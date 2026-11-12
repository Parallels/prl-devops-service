// Package legacy_migrator_test — #425 task 06: the D-7 behaviour matrix
// against #424's PUBLIC API only (exported surface + sanctioned seams:
// Options.FailAfter, LegacyMigrationMarker/EnsureMarkerTable,
// DetectLegacyStore, ValidateCounts). No internals copied; the encrypted
// recipe duplicates the data-package one via exported security.* (same
// deviation note as #424's own suite — key is synthetic, Risk-8).
// Source-hash invariant (OQ-1 frozen-source) closes EVERY scenario.
// Fixtures REUSED from #424 testdata: golden.json, corrupt.json
// (partial.json intentionally — its states are covered live-constructed).
package legacy_migrator_test

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/database/legacy_migrator"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/Parallels/prl-devops-service/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	goldenTotal   = 41 // golden.json records, matches #424's own count
	goldenColls   = 19 // #424 plan specs: 16 legacy collections + 3 join-table specs
	fixtureSource = "testdata/golden.json"
)

// --- shared harness (D-2: memory target; fixture copies = sanctioned disk exception) ---

func fixtureCopy(t *testing.T, src string) (path string, hash [sha256.Size]byte) {
	t.Helper()
	raw, err := os.ReadFile(src) // go test runs inside the package dir
	require.NoError(t, err)
	path = filepath.Join(t.TempDir(), "data.json")
	require.NoError(t, os.WriteFile(path, raw, 0o600))
	return path, sha256.Sum256(raw)
}

func hashFile(t *testing.T, path string) [sha256.Size]byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	return sha256.Sum256(raw)
}

func newTarget(t *testing.T) *gorm.DB {
	t.Helper()
	gdb := testutil.OpenMemoryDB(t)
	testutil.MigrateSchema(t, gdb)
	require.NoError(t, legacy_migrator.EnsureMarkerTable(gdb))
	return gdb
}

func tableCount(t *testing.T, gdb *gorm.DB, table string) int64 {
	t.Helper()
	var n int64
	require.NoErrorf(t, gdb.Table(table).Count(&n).Error, "count %s", table)
	return n
}

func distinctEqTotal(t *testing.T, gdb *gorm.DB, table string) {
	t.Helper()
	var total, distinct int64
	require.NoError(t, gdb.Table(table).Count(&total).Error)
	require.NoError(t, gdb.Table(table).Distinct().Count(&distinct).Error)
	assert.Equalf(t, total, distinct, "%s: duplicate ids after re-entry", table)
}

func markerCount(t *testing.T, gdb *gorm.DB) int64 {
	t.Helper()
	var rows []legacy_migrator.LegacyMigrationMarker
	require.NoError(t, gdb.Find(&rows).Error)
	return int64(len(rows))
}

func writtenTotal(rep legacy_migrator.Report) int64 {
	var n int64
	for _, c := range rep.Collections {
		n += c.WrittenCount
	}
	return n
}

// --- D-7 matrix rows ---

// #424 AC-1. Trap guard per spec: the run must route through #424's
// detector — a file-creating "detector" (IsDataFileEmpty) would leave an
// empty data.json behind; os.Stat must still say IsNotExist after the run.
func TestMigration_EmptyStoreSilentSkip(t *testing.T) {
	t.Run("absent_file", func(t *testing.T) {
		rec, ctx := testutil.RecordingContext()
		gdb := newTarget(t)
		missing := filepath.Join(t.TempDir(), "data.json")
		rep, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: missing})
		require.NoError(t, err)
		assert.True(t, rep.Skipped)
		assert.Equal(t, "absent", rep.SkipReason)
		assert.Zero(t, markerCount(t, gdb))
		assert.Zero(t, tableCount(t, gdb, "jobs"))
		_, statErr := os.Stat(missing) // the trap: file must NOT have been created
		assert.True(t, os.IsNotExist(statErr), "detector must not create the source file")
		assert.NotEmpty(t, rec.Containing("skipped"))
	})
	t.Run("zero_byte", func(t *testing.T) {
		_, ctx := testutil.RecordingContext()
		gdb := newTarget(t)
		empty := filepath.Join(t.TempDir(), "data.json")
		require.NoError(t, os.WriteFile(empty, nil, 0o600))
		rep, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: empty})
		require.NoError(t, err)
		assert.True(t, rep.Skipped)
		assert.Equal(t, "empty", rep.SkipReason)
		assert.Zero(t, markerCount(t, gdb))
	})
}

// #424 AC-2/AC-4/AC-6: 16 collections incl. pointer singletons
// (Configuration, ReverseProxy) and embedded slices (User.Roles/Claims)
// land; validation clean (no warn lines at all).
func TestMigration_HappyPathAllCollections(t *testing.T) {
	src, hash := fixtureCopy(t, fixtureSource)
	rec, ctx := testutil.RecordingContext()
	gdb := newTarget(t)
	rep, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src})
	require.NoError(t, err)
	require.False(t, rep.Skipped)
	assert.Len(t, rep.Collections, goldenColls)
	assert.EqualValues(t, goldenTotal, writtenTotal(rep))
	assert.Empty(t, rep.Validation)
	assert.Empty(t, rec.Containing("warn: "), "happy path must log zero warnings")
	assert.NotEmpty(t, rec.Containing(" records"), "AC-4 per-collection info lines")
	assert.Equal(t, int64(1), tableCount(t, gdb, "configurations"), "pointer singleton landed")
	assert.Positive(t, tableCount(t, gdb, "reverse_proxy_hosts"))
	assert.Positive(t, tableCount(t, gdb, "user_roles"), "embedded User.Roles landed")
	assert.Positive(t, tableCount(t, gdb, "user_claims"), "embedded User.Claims landed")
	assert.EqualValues(t, 1, markerCount(t, gdb))
	require.Equal(t, hash, hashFile(t, src), "source must never be modified")
}

// #424 AC-2 over ciphertext (AD-5 loader path). Synthetic RSA key —
// test-only by construction (Risk-8).
func TestMigration_EncryptedStore(t *testing.T) {
	encryptFixture := func(t *testing.T, dir string) (string, string) {
		t.Helper()
		raw, err := os.ReadFile(fixtureSource)
		require.NoError(t, err)
		pemPath := filepath.Join(t.TempDir(), "k.pem")
		require.NoError(t, security.GenPrivateRsaKey(pemPath, 2048))
		pemBytes, err := os.ReadFile(pemPath)
		require.NoError(t, err)
		cipher, err := security.EncryptString(string(pemBytes), string(raw))
		require.NoError(t, err)
		path := filepath.Join(dir, "data.json")
		require.NoError(t, os.WriteFile(path, cipher, 0o600))
		return path, base64.StdEncoding.EncodeToString(pemBytes)
	}
	t.Run("with_key_rows_land", func(t *testing.T) {
		dir := t.TempDir()
		src, key := encryptFixture(t, dir)
		t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, key)
		before := hashFile(t, src)
		gdb := newTarget(t)
		_, ctx := testutil.RecordingContext()
		rep, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src})
		require.NoError(t, err)
		assert.EqualValues(t, goldenTotal, writtenTotal(rep), "ciphertext migrates identically to plaintext")
		assert.Equal(t, int64(1), markerCount(t, gdb))
		require.Equal(t, before, hashFile(t, src), "source must never be modified")
	})
	t.Run("no_key_named_error_empty_db", func(t *testing.T) {
		dir := t.TempDir()
		src, _ := encryptFixture(t, dir)
		t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR, "")
		gdb := newTarget(t)
		_, ctx := testutil.RecordingContext()
		_, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ENCRYPTION_PRIVATE_KEY")
		assert.Zero(t, markerCount(t, gdb))
		assert.Zero(t, tableCount(t, gdb, "users"))
	})
}

// #424 R-6 posture: undecodable store ⇒ named error, no panic, target untouched.
func TestMigration_CorruptJSON(t *testing.T) {
	src, hash := fixtureCopy(t, "testdata/corrupt.json")
	_, ctx := testutil.RecordingContext()
	gdb := newTarget(t)
	rep, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src})
	require.Error(t, err)
	assert.False(t, rep.Skipped)
	assert.Zero(t, markerCount(t, gdb))
	assert.Zero(t, tableCount(t, gdb, "users"))
	assert.Zero(t, tableCount(t, gdb, "jobs"))
	require.Equal(t, hash, hashFile(t, src), "source must never be modified")
}
