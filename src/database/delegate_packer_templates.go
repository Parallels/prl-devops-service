package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/packer_templates delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetPackerTemplates(ctx basecontext.ApiContext, filter string) ([]models.PackerTemplate, error) {
	return g.legacy.GetPackerTemplates(ctx, filter)
}

func (g *GormDatabase) GetPackerTemplate(ctx basecontext.ApiContext, nameOrId string) (*models.PackerTemplate, error) {
	return g.legacy.GetPackerTemplate(ctx, nameOrId)
}

func (g *GormDatabase) AddPackerTemplate(ctx basecontext.ApiContext, template *models.PackerTemplate) (*models.PackerTemplate, error) {
	return g.legacy.AddPackerTemplate(ctx, template)
}

func (g *GormDatabase) DeletePackerTemplate(ctx basecontext.ApiContext, nameOrId string) error {
	return g.legacy.DeletePackerTemplate(ctx, nameOrId)
}

func (g *GormDatabase) UpdatePackerTemplate(ctx basecontext.ApiContext, template *models.PackerTemplate) (*models.PackerTemplate, error) {
	return g.legacy.UpdatePackerTemplate(ctx, template)
}
