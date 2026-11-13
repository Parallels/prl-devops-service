package parity

// Engine-parity scenarios — jobs + config + user configs (PBI #425 task 04).
// Provenance: data/jobs_timeout_test.go + data/config_test.go (Tier 1) and
// the #421 task-15 ghost-job semantics; owner-scoping lands on the #421
// task-09 fixture pattern (per-owner EqualFold slug scoping). The
// ghost-job timeout needs PRE-STAMPED rows (CreateJob stamps
// UpdatedAt/CreatedAt itself, and the legacy CreateJob quirk even orders
// UpdatedAt BEFORE CreatedAt) ⇒ E-4 hook engines pre-seed the exact
// stale-matrix rows and the suite asserts the observable state
// transitions + handler payload, per spec §2 ("state transition, not the
// goroutine" — the goroutine/ticker legs stay in the #421 watcher suite).
// Split noted in the ledger.
//
// Exclusion note (D-1): TestDeleteOrchestratorVirtualMachine's no-op-bug
// leg stays Tier-1-only — asserting it cross-engine would pin a known
// legacy bug (its own comment); wave covers the engine-API VM paths.

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestParityJobsConfigInfraWave(t *testing.T) {
	testutil.RunOnEngines(t, jobsScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
	testutil.RunOnEngines(t, configScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
}

func jobsScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "jobs/state_walk", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateJob(ctx, models.Job{ID: "tj-1", Owner: "u1", State: constants.JobStatePending})
			require.NoError(t, err)
			require.Equal(t, "tj-1", created.ID, "job IDs are caller-assigned (both engines)")
			created.State = constants.JobStateRunning
			created.Progress = 50
			require.NoError(t, db.UpdateJob(ctx, *created))
			running, err := db.GetJob(ctx, "tj-1")
			require.NoError(t, err)
			assert.Equal(t, constants.JobStateRunning, running.State)
			assert.Equal(t, 50, running.Progress)
			// forwarded-failure message persistence (legacy jobs_timeout :52-68):
			running.State = constants.JobStateFailed
			running.Error = "Job timed out after 30 minutes without a progress update."
			running.Message = running.Error
			require.NoError(t, db.UpdateJob(ctx, *running))
			stored, err := db.GetJob(ctx, "tj-1")
			require.NoError(t, err)
			assert.Equal(t, running.Error, stored.Error)
			assert.Equal(t, stored.Error, stored.Message, "Error==Message round-trip")
			owned, err := db.GetJobsByOwner(ctx, "u1")
			require.NoError(t, err)
			assert.Len(t, owned, 1)
			foreign, err := db.GetJobsByOwner(ctx, "somebody-else")
			require.NoError(t, err)
			assert.Empty(t, foreign)
			_, err = db.GetJob(ctx, "no-such-job")
			assert.Equal(t, data.ErrJobNotFound, err)
			require.NoError(t, db.DeleteJob(ctx, ""), "DeleteJob(\"\") silently nil (legacy :135 quirk)")
			require.NoError(t, db.DeleteJob(ctx, "tj-1"))
		}},
		{Name: "jobs/delete_by_state", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			for _, j := range []models.Job{
				{ID: "ds-1", State: constants.JobStateCompleted},
				{ID: "ds-2", State: constants.JobStateCompleted},
				{ID: "ds-3", State: constants.JobStateRunning},
			} {
				_, err := db.CreateJob(ctx, j)
				require.NoError(t, err)
			}
			require.NoError(t, db.DeleteJobsByState(ctx, constants.JobStateCompleted))
			left, err := db.GetJobs(ctx)
			require.NoError(t, err)
			require.Len(t, left, 1, "count parity: exactly the two completed rows pruned")
			assert.Equal(t, "ds-3", left[0].ID)
			require.NoError(t, db.DeleteJobsByState(ctx, constants.JobStateRunning, constants.JobStatePending),
				"multi-state form")
			left, err = db.GetJobs(ctx)
			require.NoError(t, err)
			assert.Empty(t, left)
		}},
		{Name: "userconfig/owner_scoping", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			_, err := db.CreateUserConfig(ctx, models.UserConfig{UserID: "user-a", Slug: "theme", Name: "Theme", Type: "string", Value: "light"})
			require.NoError(t, err)
			_, err = db.CreateUserConfig(ctx, models.UserConfig{UserID: "user-b", Slug: "theme", Name: "Theme", Type: "string", Value: "dark"})
			require.NoError(t, err, "same slug under a different owner is allowed (#421 t09)")
			aList, err := db.GetUserConfigs(ctx, "user-a", "")
			require.NoError(t, err)
			require.Len(t, aList, 1, "no cross-owner leakage")
			assert.Equal(t, "light", aList[0].Value)
			bCfg, err := db.GetUserConfig(ctx, "user-b", "theme")
			require.NoError(t, err)
			assert.Equal(t, "dark", bCfg.Value)
			_, err = db.GetUserConfig(ctx, "user-a", "nope")
			assert.Equal(t, data.ErrUserConfigNotFound, err)
			_, err = db.CreateUserConfig(ctx, models.UserConfig{UserID: "user-a", Slug: "theme"})
			assert.Equal(t, data.ErrUserConfigAlreadyExists, err, "dup by owner+slug (500 quirk — matrix)")
			require.NoError(t, db.DeleteUserConfig(ctx, "user-a", "theme"))
			bStill, err := db.GetUserConfig(ctx, "user-b", "theme") // delete stays owner-scoped
			require.NoError(t, err)
			assert.Equal(t, "dark", bStill.Value)
		}},
	}
}

func configScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "config/singleton_and_schema", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			// SetId FIRST: GetConfiguration's create-on-read seeds a
			// random ID, which would then lock SetId forever.
			id, err := db.SetId(ctx, "custom-id")
			require.NoError(t, err)
			assert.Equal(t, "custom-id", id)
			_, err = db.SetId(ctx, "another-id")
			require.Error(t, err)
			assert.Equal(t, "ID already exists", err.Error(), "legacy message pin (config_test :44)")
			cfg, err := db.GetConfiguration(ctx) // singleton now visible
			require.NoError(t, err)
			require.NotEmpty(t, cfg.ID)
			again, err := db.GetConfiguration(ctx)
			require.NoError(t, err)
			assert.Equal(t, cfg.ID, again.ID, "single row across reads")
			assert.Equal(t, "custom-id", cfg.ID)
			require.NoError(t, db.UpdateSchemaVersion(ctx, "417.4"))
			require.NoError(t, db.UpdateSchemaVersion(ctx, "417.5"), "schema row upserts in place")
			seeded, err := db.SeedId(ctx)
			require.NoError(t, err)
			got, err := db.GetId(ctx)
			require.NoError(t, err)
			assert.Equal(t, seeded, got)
			assert.Equal(t, "custom-id", got, "SeedId never overwrites an existing id")
		}},
	}
}

// TestParityGhostJobTimeout is the E-4 shape: identical stale-matrix rows
// pre-seeded on BOTH engines (legacy pre-Write file / gorm raw table
// insert). It pins the boot-recovery DIVERGENCE explicitly: legacy
// Connect auto-converts every running/pending row to failed at boot
// (ghost-timeout is a post-boot safety net — its Tier-1 test mutates
// internals after Connect precisely because of this), while the SQL
// engine defers to the explicit RecoverOngoingJobs/DetectStaleJobs legs
// (#421 jobs_watcher.go) so the ghost timeout is the PRIMARY recovery
// path in SQL mode. Terminal convergence (stale rows end `failed` on
// both) is the shared parity claim; trigger mechanics are pinned
// per-engine. Ledger records the split ("state transition, not the
// goroutine" — ticker legs stay in the #421 watcher suite).
func TestParityGhostJobTimeout(t *testing.T) {
	t.Setenv(constants.GhostJobTimeoutMinutesEnvVar, "30") // before engine creation
	stale := time.Now().Add(-31 * time.Minute).UTC().Format(time.RFC3339Nano)
	recent := time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339Nano)
	rows := []models.Job{
		{ID: "stale-pend", State: constants.JobStatePending, UpdatedAt: stale, Progress: 69,
			IsOrchestratorJob: true, ResultRecordId: "vm-id",
			Steps: []models.JobStep{{Name: "register", State: constants.JobStateRunning}}},
		{ID: "recent-pend", State: constants.JobStatePending, UpdatedAt: recent},
		{ID: "done", State: constants.JobStateCompleted, UpdatedAt: stale, Result: "keep"},
		{ID: "failed", State: constants.JobStateFailed, UpdatedAt: stale, Error: "original failure"},
		{ID: "boot-run", State: constants.JobStateRunning, UpdatedAt: stale},
	}
	seedLegacy := func(t *testing.T, dbPath string) {
		raw, err := json.Marshal(data.Data{Jobs: rows})
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dbPath, raw, 0o600))
	}
	seedGorm := func(t *testing.T, gdb *gorm.DB) {
		for i := range rows {
			require.NoError(t, gdb.Table("jobs").Create(&rows[i]).Error)
		}
	}
	testutil.RunOnEngines(t, []testutil.Scenario{{Name: "jobs/ghost_timeout", Run: func(t *testing.T, _ ctxT, db dbT) {
		ctx := tplCtxRoot()
		var fired []models.Job
		db.SetJobTimeoutHandler(func(j models.Job) { fired = append(fired, j) })
		db.DetectStaleJobs(ctx)
		if _, isLegacy := db.(*data.JsonDatabase); isLegacy {
			// Boot posture: legacy already converted the in-flight rows at
			// Connect — DetectStaleJobs finds nothing left to time out.
			assert.Empty(t, fired, "legacy: boot recovery pre-empted the ghost timeout")
		} else {
			// SQL posture: ghost timeout is the primary recovery.
			require.Len(t, fired, 2, "stale pending + stale running both time out")
			byID := map[string]models.Job{}
			for _, j := range fired {
				byID[j.ID] = j
			}
			sr := byID["stale-pend"]
			assert.Equal(t, constants.JobStateFailed, sr.State)
			assert.Contains(t, sr.Error, "Job timed out after 30 minutes without a progress update")
			assert.Contains(t, sr.Error, stale, "reason carries the last-update stamp")
			assert.Equal(t, sr.Error, sr.Message)
			assert.Equal(t, 69, sr.Progress, "unrelated fields survive the timeout write")
			assert.True(t, sr.IsOrchestratorJob)
			assert.Equal(t, "vm-id", sr.ResultRecordId)
			require.Len(t, sr.Steps, 1)
			assert.Equal(t, constants.JobStateRunning, sr.Steps[0].State)
			assert.Equal(t, constants.JobStateFailed, byID["boot-run"].State)
			recentJob, err := db.GetJob(ctx, "recent-pend")
			require.NoError(t, err)
			assert.Equal(t, constants.JobStatePending, recentJob.State, "recent job untouched")
			db.DetectStaleJobs(ctx)
			assert.Len(t, fired, 2, "terminal jobs must not time out again")
		}
		// Terminal convergence (the shared cross-engine claim): every
		// in-flight pre-seed ends failed on BOTH engines after one cycle:
		for _, id := range []string{"stale-pend", "boot-run"} {
			j, err := db.GetJob(ctx, id)
			require.NoError(t, err)
			assert.Equal(t, constants.JobStateFailed, j.State, "%s converges to failed", id)
		}
		done, err := db.GetJob(ctx, "done")
		require.NoError(t, err)
		assert.Equal(t, "keep", done.Result, "completed job untouched")
		failed, err := db.GetJob(ctx, "failed")
		require.NoError(t, err)
		assert.Equal(t, "original failure", failed.Error, "failed job keeps its error")
	}}}, authLegacyEngine(seedLegacy), authGormEngine(seedGorm))
}

// TestParityAD4BothTablesExist — SQL-leg only: the legacy schema_metadata
// table and the configurations singleton coexist post-write (AD-4:
// distinct from #424's marker table; cheap existence guard, spec §2).
func TestParityAD4BothTablesExist(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	gdb := testutil.OpenMemoryDB(t)
	testutil.MigrateSchema(t, gdb)
	svc := database.NewGormDatabase(ctx, gdb)
	require.NoError(t, svc.Connect(ctx))
	t.Cleanup(func() { _ = svc.Disconnect(ctx) })
	require.NoError(t, svc.UpdateSchemaVersion(ctx, "417.6"))
	_, err := svc.GetConfiguration(ctx) // ensure the singleton row path runs
	require.NoError(t, err)
	require.True(t, gdb.Migrator().HasTable("schema_metadata"))
	require.True(t, gdb.Migrator().HasTable("configurations"))
}
