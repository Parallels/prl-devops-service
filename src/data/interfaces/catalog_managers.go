package interfaces

// CatalogManagersService mirrors catalog_managers method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type CatalogManagersService interface {
	GetCatalogManagers() ([]models.CatalogManager, error)
	GetCatalogManager(id string) (*models.CatalogManager, error)
	AddCatalogManager(ctx basecontext.ApiContext, mgr models.CatalogManager) error
	UpdateCatalogManager(ctx basecontext.ApiContext, mgr models.CatalogManager) error
	DeleteCatalogManager(ctx basecontext.ApiContext, id string) error
}
