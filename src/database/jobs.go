package database

// Jobs on GORM (PBI #421 task 15). Legacy src/data/jobs.go mirrored
// method-by-method — including the owner-enrichment side effects: the
// three reads re-stamp OwnerName/OwnerEmail via GetUser → GetUsers, which
// carries NO authorisation at the data layer (users.go:30-43 is a plain
// property filter) — ctx never gates enrichment, only owner-existence
// does (err-skip on ErrUserNotFound).
// Quirks mirrored: CreateJob stamps UpdatedAt BEFORE CreatedAt (:83-84,
// two clock calls); the stored copy is un-enriched (the append happens
// before the lookup, :90-96); UpdateJob matches the id EXACTLY (:110,
// unlike Fold in GetJob/DeleteJob) and rewrites exactly its eleven fields
// plus UpdatedAt; DeleteJob("") is silently nil (:135); GetJobs answers
// NON-NIL for zero rows (make(len), :25) while GetJobsByOwner answers nil
// (var-scan, :42). GetJobs/DetectStaleJobs/Recover have NO authz wrap and
// DeleteJobsByState takes no lock — D-9 grep confirms (dataMutex.Lock only
// in Update/Delete/Recover/SetHandler/Detect): the single bulk DELETE
// stays tx-free, the two check-then-act pairs (stale-detect mark-updates)
// get one tx each per D-9's allowance, handler invocations fire AFTER the
// commit mirroring legacy's post-unlock calls.

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
)

// jobsStamp / jobsNowTime route every clock read of this domain through
// the package-private nowFn hook (nil = real clock; stale-window tests
// inject without sleeping — spec §4).
func (g *GormDatabase) jobsNowTime() time.Time {
	if g.nowFn != nil {
		return g.nowFn()
	}
	return time.Now()
}

func (g *GormDatabase) jobsStamp() string {
	if g.nowFn != nil {
		return g.nowFn().UTC().Format(time.RFC3339Nano)
	}
	return helpers.GetUtcCurrentDateTime()
}

// enrichJobs mirrors the per-job GetUser(ctx, job.Owner) lookups: Fold on
// ID || Email || Username against the full user list (one fetch per
// call instead of legacy's per-job refetch — identical observable outcome,
// same authz). Failed lookups leave the stored name/email untouched.
func (g *GormDatabase) enrichJobs(ctx basecontext.ApiContext, jobs []models.Job) {
	if len(jobs) == 0 {
		return
	}
	users, err := g.GetUsers(ctx, "")
	if err != nil {
		return
	}
	for i := range jobs {
		for _, u := range users {
			if strings.EqualFold(u.ID, jobs[i].Owner) || strings.EqualFold(u.Email, jobs[i].Owner) ||
				strings.EqualFold(u.Username, jobs[i].Owner) {
				jobs[i].OwnerName = u.Name
				jobs[i].OwnerEmail = u.Email
				break
			}
		}
	}
}

func (g *GormDatabase) jobRows() ([]models.Job, error) {
	var rows []models.Job
	if err := g.gormDB.Table("jobs").Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

func (g *GormDatabase) GetJobs(ctx basecontext.ApiContext) ([]models.Job, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.jobRows()
	if err != nil {
		return nil, err
	}
	jobs := make([]models.Job, len(rows)) // non-nil at zero rows (:25)
	copy(jobs, rows)
	g.enrichJobs(ctx, jobs)
	return jobs, nil
}

func (g *GormDatabase) GetJobsByOwner(ctx basecontext.ApiContext, owner string) ([]models.Job, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.jobRows()
	if err != nil {
		return nil, err
	}
	var jobs []models.Job // nil at zero matches (:42)
	for _, r := range rows {
		if strings.EqualFold(r.Owner, owner) {
			jobs = append(jobs, r)
		}
	}
	g.enrichJobs(ctx, jobs)
	return jobs, nil
}

func (g *GormDatabase) GetJob(ctx basecontext.ApiContext, id string) (*models.Job, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.jobRows()
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if strings.EqualFold(r.ID, id) { // Fold on read (:62)
			one := []models.Job{r} // enrichment mutates the backing array, not a copy
			g.enrichJobs(ctx, one)
			return &one[0], nil
		}
	}
	return nil, interfaces.ErrJobNotFound
}

func (g *GormDatabase) CreateJob(ctx basecontext.ApiContext, job models.Job) (*models.Job, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	if job.ID == "" {
		job.ID = helpers.GenerateId()
	}
	job.UpdatedAt = helpers.GetUtcCurrentDateTime() // order mirrored (:83-84)
	job.CreatedAt = helpers.GetUtcCurrentDateTime()
	if string(job.State) == "" || !job.State.IsValid() {
		job.State = constants.JobStatePending
	}
	if err := mapGormError(g.gormDB.Table("jobs").Create(&job).Error, nil, nil); err != nil {
		return nil, err
	}
	// Enrich the RETURNED copy only — legacy appends before looking the
	// owner up, the stored row keeps whatever the caller passed (:90-96).
	one := []models.Job{job}
	g.enrichJobs(ctx, one)
	return &one[0], nil
}

func (g *GormDatabase) UpdateJob(ctx basecontext.ApiContext, key models.Job) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	stepsJSON, err := json.Marshal(key.Steps) // map-writes bypass serializer:json — same encoding
	if err != nil {
		return err
	}
	res := g.gormDB.Table("jobs").Where("id = ?", key.ID).Updates(map[string]interface{}{ // EXACT id (:110)
		"state":                 key.State,
		"message":               key.Message,
		"progress":              key.Progress,
		"result":                key.Result,
		"result_record_id":      key.ResultRecordId,
		"result_record_name":    key.ResultRecordName,
		"result_record_link_id": key.ResultRecordLinkId,
		"result_record_type":    key.ResultRecordType,
		"error":                 key.Error,
		"steps":                 string(stepsJSON), // serialised column write (#419 D-12)
		"is_orchestrator_job":   key.IsOrchestratorJob,
		"updated_at":            helpers.GetUtcCurrentDateTime(),
	})
	if res.Error != nil {
		return mapGormError(res.Error, nil, nil)
	}
	if res.RowsAffected == 0 {
		return interfaces.ErrJobNotFound
	}
	return nil
}

func (g *GormDatabase) DeleteJob(ctx basecontext.ApiContext, id string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if id == "" {
		return nil // silent idempotence (:135)
	}
	rows, err := g.jobRows()
	if err != nil {
		return err
	}
	for _, r := range rows {
		if strings.EqualFold(r.ID, id) { // Fold on delete (:142)
			return mapGormError(g.gormDB.Table("jobs").Where("id = ?", r.ID).
				Delete(&models.Job{}).Error, nil, nil)
		}
	}
	return interfaces.ErrJobNotFound
}

func (g *GormDatabase) DeleteJobsByState(ctx basecontext.ApiContext, states ...constants.JobState) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if len(states) == 0 {
		return nil // legacy rebuilds the slice keeping everything (:163-178)
	}
	// Single bulk statement, no tx: legacy holds no lock here (D-9 grep).
	return mapGormError(g.gormDB.Table("jobs").Where("state IN ?", states).
		Delete(&models.Job{}).Error, nil, nil)
}

func (g *GormDatabase) RecoverOngoingJobs(ctx basecontext.ApiContext) {
	if !g.IsConnected() {
		return
	}
	res := g.gormDB.Table("jobs").Where("state IN ?", []constants.JobState{
		constants.JobStateRunning, constants.JobStatePending,
	}).Updates(map[string]interface{}{
		"state":      constants.JobStateFailed,
		"error":      "Service crashed or was restarted",
		"updated_at": g.jobsStamp(),
	})
	if res.Error != nil {
		ctx.LogErrorf("[Database] Job recovery failed: %v", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		ctx.LogInfof("[Database] Recovered and failed ongoing jobs after restart")
	}
}

// SetJobTimeoutHandler registers the post-timeout notification (legacy
// stores it under the data lock; SQL side uses timeoutMu).
func (g *GormDatabase) SetJobTimeoutHandler(handler func(models.Job)) {
	g.timeoutMu.Lock()
	defer g.timeoutMu.Unlock()
	g.onJobTimeout = handler
}

func (g *GormDatabase) DetectStaleJobs(ctx basecontext.ApiContext) {
	if !g.IsConnected() {
		return
	}
	timeoutMinutes := config.Get().DbGhostJobTimeoutMinutes()
	timeout := time.Duration(timeoutMinutes) * time.Minute

	g.timeoutMu.RLock()
	handler := g.onJobTimeout
	g.timeoutMu.RUnlock()

	rows, err := g.jobRows()
	if err != nil {
		return
	}
	stamp := g.jobsStamp() // legacy takes ONE clock sample pre-loop (:228)
	var timedOut []models.Job
	now := g.jobsNowTime()
	for _, r := range rows {
		if r.State == constants.JobStateRunning || r.State == constants.JobStatePending {
			updatedAt, err := time.Parse(time.RFC3339Nano, r.UpdatedAt)
			if err == nil && now.Sub(updatedAt) > timeout { // strictly greater (:232)
				reason := fmt.Sprintf("Job timed out after %d minutes without a progress update. Last update: %s. The operation may still be running on the host; check its status before retrying.", timeoutMinutes, r.UpdatedAt)
				ctx.LogWarnf("[Database] Marking stale job as failed: jobID=%s state=%s progress=%v inactiveFor=%v timeout=%v reason=%q", r.ID, r.State, r.Progress, now.Sub(updatedAt), timeout, reason)
				r.State = constants.JobStateFailed
				r.Error = reason
				r.Message = reason
				r.UpdatedAt = stamp
				timedOut = append(timedOut, r)
			}
		}
	}
	if len(timedOut) == 0 {
		return
	}
	err = g.gormDB.Transaction(func(tx *gorm.DB) error {
		for _, j := range timedOut {
			if err := mapGormError(tx.Table("jobs").Where("id = ?", j.ID).Updates(map[string]interface{}{
				"state": j.State, "error": j.Error, "message": j.Message, "updated_at": j.UpdatedAt,
			}).Error, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		ctx.LogErrorf("[Database] Stale job update failed: %v", err)
		return
	}
	ctx.LogInfof("[Database] Marked %d stale jobs as failed (timeout: %v)", len(timedOut), timeout)
	if handler != nil { // post-commit, once per job (:245-252)
		for _, j := range timedOut {
			handler(j)
		}
	}
}
