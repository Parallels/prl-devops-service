package jobs

import (
	"context"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/mappers"
	global_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

var globalJobManagerService *JobManagerService

type JobManagerService struct {
	apiCtx basecontext.ApiContext
	db     *data.JsonDatabase
	ctx    context.Context
	cancel context.CancelFunc
}

func Get(ctx basecontext.ApiContext) *JobManagerService {
	db, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		globalJobManagerService = nil
		return nil
	}

	if globalJobManagerService == nil || globalJobManagerService.db != db {
		globalJobManagerService = &JobManagerService{
			apiCtx: ctx,
			db:     db,
		}
	} else {
		globalJobManagerService.apiCtx = ctx
	}

	return globalJobManagerService
}

func New(ctx basecontext.ApiContext) *JobManagerService {
	db, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		globalJobManagerService = nil
		ctx.LogErrorf("Error getting database service for job manager: %v", err)
		return nil
	}

	globalJobManagerService = &JobManagerService{
		apiCtx: ctx,
		db:     db,
	}

	return globalJobManagerService
}

func (jms *JobManagerService) Start() error {
	jms.apiCtx.LogInfof("[Job Manager] Starting Job Manager Service")
	jms.ctx, jms.cancel = context.WithCancel(context.Background())
	return nil
}

func (jms *JobManagerService) Stop() error {
	jms.apiCtx.LogInfof("[Job Manager] Stopping Job Manager Service")
	if jms.cancel != nil {
		jms.cancel()
	}
	return nil
}

func (jms *JobManagerService) CreateNewJob(owner string, jobType string, jobOperation string, action string) (*data_models.Job, error) {
	job := data_models.Job{
		Owner:        owner,
		State:        constants.JobStatePending,
		JobType:      jobType,
		JobOperation: jobOperation,
		Progress:     0,
		Steps:        make([]data_models.JobStep, 0),
	}

	createdJob, err := jms.db.CreateJob(jms.apiCtx, job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_CREATED", createdJob)
	return createdJob, nil
}

func (jms *JobManagerService) CreateOrchestratorJob(owner string, jobType string, jobOperation string, action string, externalJobID string) (*data_models.Job, error) {
	job := data_models.Job{
		ID:                externalJobID,
		Owner:             owner,
		State:             constants.JobStatePending,
		JobType:           jobType,
		JobOperation:      jobOperation,
		Progress:          0,
		IsOrchestratorJob: true,
		Steps:             make([]data_models.JobStep, 0),
	}

	createdJob, err := jms.db.CreateJob(jms.apiCtx, job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_CREATED", createdJob)
	return createdJob, nil
}

func (jms *JobManagerService) InitJob(jobId string) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	job.State = constants.JobStateInit

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

func (jms *JobManagerService) UpdateJobProgress(jobId string, progress int, state constants.JobState) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	if job.State == constants.JobStateCompleted || job.State == constants.JobStateFailed {
		return job, nil
	}

	job.Progress = progress
	job.State = state

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

func (jms *JobManagerService) UpdateJobSteps(jobId string, steps []data_models.JobStep) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	if job.State == constants.JobStateCompleted || job.State == constants.JobStateFailed {
		return job, nil
	}

	job.Steps = steps

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

// UpdateJobProgressAndSteps atomically writes both the overall progress percentage
// and the step snapshot in one DB update and one event broadcast.
// This avoids the race condition where separate UpdateJobSteps + UpdateJobProgress
// calls could emit a stale "no steps" event because the second read happened
// before the first write was visible.
func (jms *JobManagerService) UpdateJobProgressAndSteps(jobId string, progress int, state constants.JobState, steps []data_models.JobStep) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	if job.State == constants.JobStateCompleted || job.State == constants.JobStateFailed {
		return job, nil
	}

	job.Progress = progress
	job.State = state
	job.Steps = steps

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

// UpdateJobProgressStepsAndMessage atomically syncs progress, state, steps,
// and message in a single DB write and a single JOB_UPDATED event. This is
// used when mirroring a host job into an orchestrator job so the UI sees the
// full rich body (steps, percentages, filenames, etc.) in one update.
func (jms *JobManagerService) UpdateJobProgressStepsAndMessage(jobId string, progress int, state constants.JobState, steps []data_models.JobStep, message string) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	if job.State == constants.JobStateCompleted || job.State == constants.JobStateFailed {
		return job, nil
	}

	job.Progress = progress
	job.State = state
	job.Steps = steps
	if message != "" {
		job.Message = message
	}

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

func (jms *JobManagerService) UpdateJobMessage(jobId string, message string) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	job.Message = message

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

func (jms *JobManagerService) UpdateJobResultRecord(jobId string, recordId string, recordName string, recordType string, recordLinkId string) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	job.ResultRecordId = recordId
	job.ResultRecordType = recordType
	job.ResultRecordName = recordName
	job.ResultRecordLinkId = recordLinkId

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	// Deliberately not emitting an event here, per user request, this is just a silent DB update
	return job, nil
}

func (jms *JobManagerService) MarkJobComplete(jobID string, result string) error {
	// Applying a slowdown to allow other messages for this job
	// to be processed before the "completed" event is emitted,
	// so the UI can show the final progress and steps before
	// transitioning to completed.
	time.Sleep(50 * time.Millisecond)
	job, err := jms.db.GetJob(jms.apiCtx, jobID)
	if err != nil {
		return err
	}

	job.State = constants.JobStateCompleted
	job.Progress = 100
	job.Result = result

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return err
	}

	jms.emitEvent("JOB_UPDATED", job)
	// Applying a slowdown to allow other messages for this job
	// to be processed before the "completed" event is emitted,
	// so the UI can show the final progress and steps before
	// transitioning to completed.
	time.Sleep(500 * time.Millisecond)
	jms.emitEvent("JOB_COMPLETED", job)
	return nil
}

func (jms *JobManagerService) MarkJobCompleteWithRecord(jobID string, result string, recordID string, recordName string, recordType string, recordLinkId string) error {
	// Applying a slowdown to allow other messages for this job
	// to be processed before the "completed" event is emitted,
	// so the UI can show the final progress and steps before
	// transitioning to completed.
	time.Sleep(50 * time.Millisecond)
	job, err := jms.db.GetJob(jms.apiCtx, jobID)
	if err != nil {
		return err
	}

	job.State = constants.JobStateCompleted
	job.Progress = 100
	job.Result = result
	job.ResultRecordId = recordID
	job.ResultRecordType = recordType
	job.ResultRecordName = recordName
	job.ResultRecordLinkId = recordLinkId

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return err
	}

	jms.emitEvent("JOB_UPDATED", job)
	// Applying a slowdown to allow other messages for this job
	// to be processed before the "completed" event is emitted,
	// so the UI can show the final progress and steps before
	// transitioning to completed.
	time.Sleep(500 * time.Millisecond)
	jms.emitEvent("JOB_COMPLETED", job)
	return nil
}

func (jms *JobManagerService) MarkJobError(jobId string, jobErr error) error {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return err
	}

	job.State = constants.JobStateFailed
	if jobErr != nil {
		job.Error = jobErr.Error()
	}

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return nil
}

func (jms *JobManagerService) DeleteJob(jobId string) error {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return err
	}

	if err := jms.db.DeleteJob(jms.apiCtx, jobId); err != nil {
		return err
	}

	jms.emitEvent("JOB_DELETED", job)
	return nil
}

func (jms *JobManagerService) emitEvent(message string, job *data_models.Job) {
	if job == nil {
		jms.apiCtx.LogDebugf("[Orchestrator] [Jobs] emitEvent called with nil job, message=%s", message)
		return
	}
	emitter := serviceprovider.GetEventEmitter()
	if emitter == nil {
		jms.apiCtx.LogDebugf("[Orchestrator] [Jobs] emitEvent: emitter is nil, message=%s jobID=%s", message, job.ID)
		return
	}
	if !emitter.IsRunning() {
		jms.apiCtx.LogDebugf("[Orchestrator] [Jobs] emitEvent: emitter not running, message=%s jobID=%s", message, job.ID)
		return
	}
	jms.apiCtx.LogDebugf("[Orchestrator] [Jobs] emitEvent: message=%s jobID=%s jobState=%s progress=%d", message, job.ID, job.State, job.Progress)
	// Always broadcast the mapped API model so the UI always receives
	// the full schema including Steps (never the raw DB struct).
	// NOTE: Broadcast is synchronous (not in a goroutine) to preserve event
	// ordering. The Go scheduler's LIFO goroutine scheduling would otherwise
	// cause MarkJobComplete's "completed" event to be delivered before the
	// preceding "running" step-update event, then overwritten when the
	// "running" goroutine runs after, leaving the UI stuck showing "running".
	apiJob := mappers.MapJobToApiJob(*job)
	msg := global_models.NewEventMessage(constants.EventTypeJobManager, message, apiJob)
	if err := emitter.Broadcast(msg); err != nil {
		jms.apiCtx.LogDebugf("[Orchestrator] [Jobs] emitEvent: broadcast failed for message=%s jobID=%s: %v", message, job.ID, err)
	}
}
