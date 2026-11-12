// behaviour_2_test.go — D-7 matrix rows 5–9 (split guard: ≤240 ln per
// file; harness + rows 1–4 in behaviour_test.go, same package).
package legacy_migrator_test

import (
	"fmt"
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/legacy_migrator"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// #424 AC-5: second run is the cheap no-op — marker gate short-circuits
// BEFORE loading collections: zero per-collection info lines in run 2.
func TestMigration_RerunMarkerShortCircuit(t *testing.T) {
	src, hash := fixtureCopy(t, fixtureSource)
	gdb := newTarget(t)
	_, ctx1 := testutil.RecordingContext()
	_, err := legacy_migrator.RunMigration(ctx1, gdb, legacy_migrator.Options{SourcePath: src})
	require.NoError(t, err)
	users1 := tableCount(t, gdb, "users")
	rec2, ctx2 := testutil.RecordingContext()
	rep2, err := legacy_migrator.RunMigration(ctx2, gdb, legacy_migrator.Options{SourcePath: src})
	require.NoError(t, err)
	assert.True(t, rep2.Skipped)
	assert.Equal(t, "already-migrated", rep2.SkipReason)
	assert.Empty(t, rec2.Containing(" records"), "run 2 must not open any collection (skip line names the marker, no collection work)")
	assert.Equal(t, users1, tableCount(t, gdb, "users"))
	assert.EqualValues(t, 1, markerCount(t, gdb))
	require.Equal(t, hash, hashFile(t, src), "source must never be modified")
}

// #424 AC-3/AC-5 crash posture, both constructions the spec names:
// (a) FailAfter seam mid-plan ⇒ ONE-tx rollback (marker INSIDE the tx) ⇒
// clean re-run; (b) marker row deleted post-commit (upsert re-entry over
// full existing rows) ⇒ zero duplicates either way.
func TestMigration_CrashReEntryNoDuplicates(t *testing.T) {
	t.Run("fail_after_rollback_then_rerun", func(t *testing.T) {
		src, hash := fixtureCopy(t, fixtureSource)
		_, ctx := testutil.RecordingContext()
		gdb := newTarget(t)
		detonate := legacy_migrator.Options{SourcePath: src, FailAfter: func(name string, done int) error {
			return fmt.Errorf("constructed crash after %s (%d rows)", name, done)
		}}
		_, err := legacy_migrator.RunMigration(ctx, gdb, detonate)
		require.Error(t, err)
		assert.Zero(t, markerCount(t, gdb), "marker lives INSIDE the tx: crash leaves nothing")
		assert.Zero(t, tableCount(t, gdb, "users"))
		rep, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src})
		require.NoError(t, err)
		assert.EqualValues(t, goldenTotal, writtenTotal(rep), "re-run completes from scratch")
		require.Equal(t, hash, hashFile(t, src), "source must never be modified")
	})
	t.Run("marker_deleted_upsert_reentry", func(t *testing.T) {
		src, hash := fixtureCopy(t, fixtureSource)
		_, ctx := testutil.RecordingContext()
		gdb := newTarget(t)
		rep1, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src})
		require.NoError(t, err)
		require.NoError(t, gdb.Where("id = ?", rep1.MarkerID).Delete(&legacy_migrator.LegacyMigrationMarker{}).Error)
		rep2, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src})
		require.NoError(t, err)
		assert.False(t, rep2.Skipped, "marker absent ⇒ migration re-runs (upsert posture)")
		for _, tbl := range []string{"users", "jobs", "roles", "claims", "api_keys", "orchestrator_hosts"} {
			distinctEqTotal(t, gdb, tbl) // upsert-by-ID: no duplicate PKs
		}
		assert.Equal(t, int64(1), markerCount(t, gdb), "re-run re-marked")
		require.Equal(t, hash, hashFile(t, src), "source must never be modified")
	})
}

// #424 AC-6 warn-not-fail, constructed from OUTSIDE: pre-existing
// target-side rows the legacy store does not know about skew the
// post-commit counts ⇒ MISMATCH warn naming the collection, err == nil,
// committed data intact.
func TestMigration_ValidationMismatchWarnsNotFails(t *testing.T) {
	src, hash := fixtureCopy(t, fixtureSource)
	rec, ctx := testutil.RecordingContext()
	gdb := newTarget(t)
	require.NoError(t, gdb.Create(&models.User{
		ID: "pre-existing-stray", Username: "stray", Email: "stray@example.com",
	}).Error)
	rep, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src})
	require.NoError(t, err, "validation NEVER fails a committed migration")
	var skew *legacy_migrator.ValidationIssue
	for i := range rep.Validation {
		if rep.Validation[i].Collection == "Users" { // spec name, capitalised
			skew = &rep.Validation[i]
		}
	}
	require.NotNil(t, skew, "report must name the skewed collection")
	assert.Equal(t, skew.Expected+1, skew.Actual, "stray row = exactly the injected skew")
	warn := rec.Containing("MISMATCH Users")
	require.NotEmpty(t, warn, "warn line naming the collection")
	assert.Positive(t, tableCount(t, gdb, "users"), "data stayed committed")
	require.Equal(t, hash, hashFile(t, src), "source must never be modified")
}

// #424 AC-7: projection only — target empty, no marker, report at info,
// source byte-identical.
func TestMigration_DryRunInert(t *testing.T) {
	src, hash := fixtureCopy(t, fixtureSource)
	rec, ctx := testutil.RecordingContext()
	gdb := newTarget(t)
	rep, err := legacy_migrator.RunMigration(ctx, gdb, legacy_migrator.Options{SourcePath: src, DryRun: true})
	require.NoError(t, err)
	assert.True(t, rep.DryRun)
	assert.Len(t, rep.Collections, goldenColls)
	assert.Zero(t, markerCount(t, gdb))
	assert.Zero(t, tableCount(t, gdb, "users"))
	assert.Zero(t, tableCount(t, gdb, "jobs"))
	assert.NotEmpty(t, rec.Containing("dry-run"))
	require.Equal(t, hash, hashFile(t, src), "dry-run leaves the source byte-identical")
}
