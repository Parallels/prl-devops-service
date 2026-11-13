package interfaces

// JobsService mirrors jobs method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
)

type JobsService interface {
	GetJobs(ctx basecontext.ApiContext) ([]models.Job, error)
	GetJobsByOwner(ctx basecontext.ApiContext, owner string) ([]models.Job, error)
	GetJob(ctx basecontext.ApiContext, id string) (*models.Job, error)
	CreateJob(ctx basecontext.ApiContext, job models.Job) (*models.Job, error)
	UpdateJob(ctx basecontext.ApiContext, key models.Job) error
	DeleteJob(ctx basecontext.ApiContext, id string) error
	DeleteJobsByState(ctx basecontext.ApiContext, states ...constants.JobState) error
	RecoverOngoingJobs(ctx basecontext.ApiContext)
	SetJobTimeoutHandler(handler func(models.Job))
	DetectStaleJobs(ctx basecontext.ApiContext)
}
