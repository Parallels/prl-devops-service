package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/jobs delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetJobs(ctx basecontext.ApiContext) ([]models.Job, error) {
	return g.legacy.GetJobs(ctx)
}

func (g *GormDatabase) GetJobsByOwner(ctx basecontext.ApiContext, owner string) ([]models.Job, error) {
	return g.legacy.GetJobsByOwner(ctx, owner)
}

func (g *GormDatabase) GetJob(ctx basecontext.ApiContext, id string) (*models.Job, error) {
	return g.legacy.GetJob(ctx, id)
}

func (g *GormDatabase) CreateJob(ctx basecontext.ApiContext, job models.Job) (*models.Job, error) {
	return g.legacy.CreateJob(ctx, job)
}

func (g *GormDatabase) UpdateJob(ctx basecontext.ApiContext, key models.Job) error {
	return g.legacy.UpdateJob(ctx, key)
}

func (g *GormDatabase) DeleteJob(ctx basecontext.ApiContext, id string) error {
	return g.legacy.DeleteJob(ctx, id)
}

func (g *GormDatabase) DeleteJobsByState(ctx basecontext.ApiContext, states ...constants.JobState) error {
	return g.legacy.DeleteJobsByState(ctx, states...)
}

func (g *GormDatabase) RecoverOngoingJobs(ctx basecontext.ApiContext) {
	g.legacy.RecoverOngoingJobs(ctx)
}

func (g *GormDatabase) SetJobTimeoutHandler(handler func(models.Job)) {
	g.legacy.SetJobTimeoutHandler(handler)
}

func (g *GormDatabase) DetectStaleJobs(ctx basecontext.ApiContext) {
	g.legacy.DetectStaleJobs(ctx)
}
