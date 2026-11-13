package interfaces

// PackerTemplatesService mirrors packer_template method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type PackerTemplatesService interface {
	GetPackerTemplates(ctx basecontext.ApiContext, filter string) ([]models.PackerTemplate, error)
	GetPackerTemplate(ctx basecontext.ApiContext, nameOrId string) (*models.PackerTemplate, error)
	AddPackerTemplate(ctx basecontext.ApiContext, template *models.PackerTemplate) (*models.PackerTemplate, error)
	DeletePackerTemplate(ctx basecontext.ApiContext, nameOrId string) error
	UpdatePackerTemplate(ctx basecontext.ApiContext, template *models.PackerTemplate) (*models.PackerTemplate, error)
}
