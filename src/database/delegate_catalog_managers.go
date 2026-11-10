package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/catalog_managers delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetCatalogManagers() ([]models.CatalogManager, error) {
	return g.legacy.GetCatalogManagers()
}

func (g *GormDatabase) GetCatalogManager(id string) (*models.CatalogManager, error) {
	return g.legacy.GetCatalogManager(id)
}

func (g *GormDatabase) AddCatalogManager(ctx basecontext.ApiContext, mgr models.CatalogManager) error {
	return g.legacy.AddCatalogManager(ctx, mgr)
}

func (g *GormDatabase) UpdateCatalogManager(ctx basecontext.ApiContext, mgr models.CatalogManager) error {
	return g.legacy.UpdateCatalogManager(ctx, mgr)
}

func (g *GormDatabase) DeleteCatalogManager(ctx basecontext.ApiContext, id string) error {
	return g.legacy.DeleteCatalogManager(ctx, id)
}
