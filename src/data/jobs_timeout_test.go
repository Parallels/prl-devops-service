package data

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/require"
)

func TestDetectStaleJobsNotifiesWithTimeoutReason(t *testing.T) {
	t.Setenv(constants.GhostJobTimeoutMinutesEnvVar, "30")
	db, dir := setupTestDB(t)
	defer cleanupTestDB(t, dir, db)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	old := time.Now().Add(-31 * time.Minute).UTC().Format(time.RFC3339Nano)
	recent := time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339Nano)
	db.data.Jobs = []models.Job{
		{ID: "stale", State: constants.JobStateRunning, UpdatedAt: old, Progress: 69, IsOrchestratorJob: true, ResultRecordId: "vm-id", Steps: []models.JobStep{{Name: "register", State: constants.JobStateRunning}}},
		{ID: "pending", State: constants.JobStatePending, UpdatedAt: old},
		{ID: "recent", State: constants.JobStateRunning, UpdatedAt: recent},
		{ID: "completed", State: constants.JobStateCompleted, UpdatedAt: old},
		{ID: "failed", State: constants.JobStateFailed, UpdatedAt: old, Error: "original failure"},
	}
	var notifications []models.Job
	db.SetJobTimeoutHandler(func(job models.Job) {
		// Acquiring this lock also verifies callbacks run outside the DB lock.
		db.dataMutex.RLock()
		defer db.dataMutex.RUnlock()
		notifications = append(notifications, job)
	})
	db.DetectStaleJobs(ctx)
	require.Len(t, notifications, 2)
	job := notifications[0]
	require.Equal(t, constants.JobStateFailed, job.State)
	require.Contains(t, job.Error, "Job timed out after 30 minutes without a progress update")
	require.Contains(t, job.Error, old)
	require.Equal(t, job.Error, job.Message)
	require.Equal(t, 69, job.Progress)
	require.True(t, job.IsOrchestratorJob)
	require.Equal(t, "vm-id", job.ResultRecordId)
	require.Len(t, job.Steps, 1)
	require.Equal(t, constants.JobStateRunning, db.data.Jobs[2].State)
	require.Equal(t, constants.JobStateCompleted, db.data.Jobs[3].State)
	require.Equal(t, "original failure", db.data.Jobs[4].Error)
	db.DetectStaleJobs(ctx)
	require.Len(t, notifications, 2, "terminal jobs must not time out again")
}

func TestUpdateJobPersistsForwardedFailureMessage(t *testing.T) {
	db, dir := setupTestDB(t)
	defer cleanupTestDB(t, dir, db)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	job, err := db.CreateJob(ctx, models.Job{ID: "forwarded", State: constants.JobStateRunning})
	require.NoError(t, err)
	job.State = constants.JobStateFailed
	job.Error = "Job timed out after 30 minutes without a progress update."
	job.Message = job.Error
	require.NoError(t, db.UpdateJob(ctx, *job))
	stored, err := db.GetJob(ctx, job.ID)
	require.NoError(t, err)
	require.Equal(t, job.Error, stored.Error)
	require.Equal(t, job.Message, stored.Message)
}
