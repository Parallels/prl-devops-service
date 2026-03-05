package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	api_models "github.com/Parallels/prl-devops-service/models"
)

func MapJobToApiJob(job data_models.Job) *api_models.JobResponse {
	response := &api_models.JobResponse{
		ID:               job.ID,
		Owner:            job.Owner,
		OwnerName:        job.OwnerName,
		OwnerEmail:       job.OwnerEmail,
		State:            job.State,
		Message:          job.Message,
		Progress:         job.Progress,
		JobType:          job.JobType,
		JobOperation:     job.JobOperation,
		Result:           job.Result,
		ResultRecordId:   job.ResultRecordId,
		ResultRecordType: job.ResultRecordType,
		Error:            job.Error,
		CreatedAt:        job.CreatedAt,
		UpdatedAt:        job.UpdatedAt,
		Steps:            make([]api_models.JobStepResponse, 0),
	}

	for _, step := range job.Steps {
		response.Steps = append(response.Steps, api_models.JobStepResponse{
			Name:              step.Name,
			Weight:            step.Weight,
			Parallel:          step.Parallel,
			HasPercentage:     step.HasPercentage,
			State:             step.State,
			CurrentPercentage: step.CurrentPercentage,
			Value:             step.Value,
			Total:             step.Total,
			ETA:               step.ETA,
			Message:           step.Message,
			Error:             step.Error,
			Filename:          step.Filename,
			Unit:              step.Unit,
		})
	}

	return response
}
