package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

var (
	ErrJobNotFound = errors.NewWithCode("job not found", 404)
)

func (j *JsonDatabase) GetJobs(ctx basecontext.ApiContext) ([]models.Job, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	jobs := make([]models.Job, len(j.data.Jobs))
	for i, job := range j.data.Jobs {
		if owner, err := j.GetUser(ctx, job.Owner); err == nil {
			job.OwnerName = owner.Name
			job.OwnerEmail = owner.Email
		}
		jobs[i] = job
	}

	return jobs, nil
}

func (j *JsonDatabase) GetJobsByOwner(ctx basecontext.ApiContext, owner string) ([]models.Job, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	var jobs []models.Job
	for _, job := range j.data.Jobs {
		if strings.EqualFold(job.Owner, owner) {
			if user, err := j.GetUser(ctx, job.Owner); err == nil {
				job.OwnerName = user.Name
				job.OwnerEmail = user.Email
			}
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func (j *JsonDatabase) GetJob(ctx basecontext.ApiContext, id string) (*models.Job, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for _, job := range j.data.Jobs {
		if strings.EqualFold(job.ID, id) {
			if user, err := j.GetUser(ctx, job.Owner); err == nil {
				job.OwnerName = user.Name
				job.OwnerEmail = user.Email
			}
			return &job, nil
		}
	}

	return nil, ErrJobNotFound
}

func (j *JsonDatabase) CreateJob(ctx basecontext.ApiContext, job models.Job) (*models.Job, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if job.ID == "" {
		job.ID = helpers.GenerateId()
	}

	job.UpdatedAt = helpers.GetUtcCurrentDateTime()
	job.CreatedAt = helpers.GetUtcCurrentDateTime()

	if string(job.State) == "" || !job.State.IsValid() {
		job.State = constants.JobStatePending
	}

	j.data.Jobs = append(j.data.Jobs, job)

	// Enrich the returned struct so the immediately emitted event has the names
	if user, err := j.GetUser(ctx, job.Owner); err == nil {
		job.OwnerName = user.Name
		job.OwnerEmail = user.Email
	}

	return &job, nil
}

func (j *JsonDatabase) UpdateJob(ctx basecontext.ApiContext, key models.Job) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	for i, job := range j.data.Jobs {
		if job.ID == key.ID {
			j.data.Jobs[i].State = key.State
			j.data.Jobs[i].Message = key.Message
			j.data.Jobs[i].Progress = key.Progress
			j.data.Jobs[i].Result = key.Result
			j.data.Jobs[i].ResultRecordId = key.ResultRecordId
			j.data.Jobs[i].ResultRecordName = key.ResultRecordName
			j.data.Jobs[i].ResultRecordLinkId = key.ResultRecordLinkId
			j.data.Jobs[i].ResultRecordType = key.ResultRecordType
			j.data.Jobs[i].Error = key.Error
			j.data.Jobs[i].Steps = key.Steps
			j.data.Jobs[i].IsOrchestratorJob = key.IsOrchestratorJob
			j.data.Jobs[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
			return nil
		}
	}

	return ErrJobNotFound
}

func (j *JsonDatabase) DeleteJob(ctx basecontext.ApiContext, id string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if id == "" {
		return nil
	}

	j.dataMutex.Lock()
	found := false
	for i, job := range j.data.Jobs {
		if strings.EqualFold(job.ID, id) {
			j.data.Jobs = append(j.data.Jobs[:i], j.data.Jobs[i+1:]...)
			found = true
			break
		}
	}
	j.dataMutex.Unlock()

	if !found {
		return ErrJobNotFound
	}

	_ = j.SaveNow(ctx)
	return nil
}

func (j *JsonDatabase) DeleteJobsByState(ctx basecontext.ApiContext, states ...constants.JobState) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	stateMap := make(map[constants.JobState]bool)
	for _, s := range states {
		stateMap[s] = true
	}

	var newJobs []models.Job
	for i, job := range j.data.Jobs {
		if stateMap[job.State] {
			// Do not include it in the new jobs list
			continue
		}
		newJobs = append(newJobs, j.data.Jobs[i])
	}

	// We replace the entire slice
	j.data.Jobs = newJobs

	return nil
}

func (j *JsonDatabase) RecoverOngoingJobs(ctx basecontext.ApiContext) {
	if !j.IsConnected() {
		return
	}

	j.dataMutex.Lock()
	updated := false
	for i, job := range j.data.Jobs {
		if job.State == constants.JobStateRunning || job.State == constants.JobStatePending {
			j.data.Jobs[i].State = constants.JobStateFailed
			j.data.Jobs[i].Error = "Service crashed or was restarted"
			j.data.Jobs[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
			updated = true
		}
	}
	j.dataMutex.Unlock()

	if updated {
		ctx.LogInfof("[Database] Recovered and failed ongoing jobs after restart")
		_ = j.SaveNow(ctx)
	}
}

// SetJobTimeoutHandler registers a notification invoked after timed-out jobs
// are updated, outside the database lock.
func (j *JsonDatabase) SetJobTimeoutHandler(handler func(models.Job)) {
	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()
	j.onJobTimeout = handler
}

// DetectStaleJobs marks pending/running jobs as failed after the configured
// inactivity timeout, preserving their progress, steps, and result metadata.
func (j *JsonDatabase) DetectStaleJobs(ctx basecontext.ApiContext) {
	if !j.IsConnected() {
		return
	}

	cfg := config.Get()
	timeoutMinutes := cfg.DbGhostJobTimeoutMinutes()
	timeout := time.Duration(timeoutMinutes) * time.Minute

	j.dataMutex.Lock()
	var timedOutJobs []models.Job
	handler := j.onJobTimeout
	now := helpers.GetUtcCurrentDateTime()
	for i, job := range j.data.Jobs {
		if job.State == constants.JobStateRunning || job.State == constants.JobStatePending {
			updatedAt, err := time.Parse(time.RFC3339Nano, job.UpdatedAt)
			if err == nil && time.Since(updatedAt) > timeout {
				reason := fmt.Sprintf("Job timed out after %d minutes without a progress update. Last update: %s. The operation may still be running on the host; check its status before retrying.", timeoutMinutes, job.UpdatedAt)
				ctx.LogWarnf("[Database] Marking stale job as failed: jobID=%s state=%s progress=%v inactiveFor=%v timeout=%v reason=%q", job.ID, job.State, job.Progress, time.Since(updatedAt), timeout, reason)
				j.data.Jobs[i].State = constants.JobStateFailed
				j.data.Jobs[i].Error = reason
				j.data.Jobs[i].Message = reason
				j.data.Jobs[i].UpdatedAt = now
				timedOutJobs = append(timedOutJobs, j.data.Jobs[i])
			}
		}
	}
	j.dataMutex.Unlock()

	if len(timedOutJobs) > 0 {
		ctx.LogInfof("[Database] Marked %d stale jobs as failed (timeout: %v)", len(timedOutJobs), timeout)
		_ = j.SaveNow(ctx)
		if handler != nil {
			for _, job := range timedOutJobs {
				handler(job)
			}
		}
	}
}
