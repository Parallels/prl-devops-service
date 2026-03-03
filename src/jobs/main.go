package jobs

import (
	"context"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	data_models "github.com/Parallels/prl-devops-service/data/models"
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
	if globalJobManagerService == nil {
		globalJobManagerService = New(ctx)
	}

	return globalJobManagerService
}

func New(ctx basecontext.ApiContext) *JobManagerService {
	db, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
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
		Action:       action,
		Progress:     0,
	}

	createdJob, err := jms.db.CreateJob(jms.apiCtx, job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_CREATED", createdJob)
	return createdJob, nil
}

func (jms *JobManagerService) UpdateJobProgress(jobId string, action string, progress int, state constants.JobState) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	if action != "" {
		if job.Action != action {
			job.Action = action
			// Reset the granular action details for the new main action
			job.ActionMessage = ""
			job.ActionPercentage = 0
			job.ActionValue = 0
			job.ActionTotal = 0
			job.ActionETA = ""
		}
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

func (jms *JobManagerService) UpdateJobActionProgress(jobId string, actionMessage string, value int64, percentage int, total int64, eta string, unit string) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	job.ActionMessage = actionMessage
	job.ActionValue = value
	job.ActionPercentage = percentage
	job.ActionTotal = total
	job.ActionETA = eta
	if unit != "" {
		job.ActionValueUnit = unit
	}

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

func (jms *JobManagerService) UpdateJobActionMessage(jobId string, actionMessage string) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	job.ActionMessage = actionMessage

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

func (jms *JobManagerService) UpdateJobActionPercent(jobId string, percentage int) (*data_models.Job, error) {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
	if err != nil {
		return nil, err
	}

	job.ActionPercentage = percentage

	err = jms.db.UpdateJob(jms.apiCtx, *job)
	if err != nil {
		return nil, err
	}

	jms.emitEvent("JOB_UPDATED", job)
	return job, nil
}

func (jms *JobManagerService) MarkJobComplete(jobId string, result string) error {
	job, err := jms.db.GetJob(jms.apiCtx, jobId)
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

	jms.emitEvent("JOB_FAILED", job)
	return nil
}

func (jms *JobManagerService) emitEvent(message string, job *data_models.Job) {
	emitter := serviceprovider.GetEventEmitter()
	if emitter != nil && emitter.IsRunning() {
		msg := global_models.NewEventMessage(constants.EventTypeJobManager, message, job)
		go func() {
			_ = emitter.Broadcast(msg)
		}()
	}
}
