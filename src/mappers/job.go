package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	api_models "github.com/Parallels/prl-devops-service/models"
)

func MapJobToApiJob(job data_models.Job) *api_models.JobResponse {
	return &api_models.JobResponse{
		ID:               job.ID,
		Owner:            job.Owner,
		OwnerName:        job.OwnerName,
		OwnerEmail:       job.OwnerEmail,
		State:            job.State,
		Progress:         job.Progress,
		JobType:          job.JobType,
		JobOperation:     job.JobOperation,
		Action:           job.Action,
		ActionMessage:    job.ActionMessage,
		ActionValue:      job.ActionValue,
		ActionPercentage: job.ActionPercentage,
		ActionTotal:      job.ActionTotal,
		ActionETA:        job.ActionETA,
		ActionValueUnit:  job.ActionValueUnit,
		Result:           job.Result,
		Error:            job.Error,
		CreatedAt:        job.CreatedAt,
		UpdatedAt:        job.UpdatedAt,
	}
}
