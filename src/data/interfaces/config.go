package interfaces

// ConfigService mirrors config method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type ConfigService interface {
	GetConfiguration(ctx basecontext.ApiContext) (*models.Configuration, error)
	GetId(ctx basecontext.ApiContext) (string, error)
	SeedId(ctx basecontext.ApiContext) (string, error)
	SetId(ctx basecontext.ApiContext, id string) (string, error)
}
