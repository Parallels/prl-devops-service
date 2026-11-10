package interfaces

// CatalogService mirrors catalog method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type CatalogService interface {
	GetCatalogManifests(ctx basecontext.ApiContext, filter string) ([]models.CatalogManifest, error)
	GetCatalogManifestByName(ctx basecontext.ApiContext, idOrName string) (*models.CatalogManifest, error)
	GetCatalogManifestByTag(ctx basecontext.ApiContext, catalogId string, tag string) (*models.CatalogManifest, error)
	GetCatalogManifestsByCatalogId(ctx basecontext.ApiContext, catalogId string) ([]models.CatalogManifest, error)
	GetCatalogManifestsByCatalogIdAndVersion(ctx basecontext.ApiContext, catalogId string, version string) ([]models.CatalogManifest, error)
	GetCatalogManifestsByCatalogIdVersionAndArch(ctx basecontext.ApiContext, catalogId string, version string, arch string) (*models.CatalogManifest, error)
	CreateCatalogManifest(ctx basecontext.ApiContext, manifest models.CatalogManifest) (*models.CatalogManifest, error)
	DeleteCatalogManifest(ctx basecontext.ApiContext, catalogIdOrId string) error
	DeleteCatalogManifestVersion(ctx basecontext.ApiContext, catalogIdOrId string, version string) error
	DeleteCatalogManifestVersionArch(ctx basecontext.ApiContext, catalogIdOrId string, version string, architecture string) error
	UpdateCatalogManifest(ctx basecontext.ApiContext, record models.CatalogManifest) (*models.CatalogManifest, error)
	UpdateCatalogManifestTags(ctx basecontext.ApiContext, record models.CatalogManifest) error
	UpdateCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error
	AddCatalogManifestRequiredClaims(ctx basecontext.ApiContext, recordId string, claims ...string) error
	RemoveCatalogManifestRequiredClaims(ctx basecontext.ApiContext, recordId string, claims ...string) error
	AddCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error
	RemoveCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error
	AddCatalogManifestTags(ctx basecontext.ApiContext, recordId string, tags ...string) error
	RemoveCatalogManifestTags(ctx basecontext.ApiContext, recordId string, tags ...string) error
	UpdateCatalogManifestDownloadCount(ctx basecontext.ApiContext, catalogId, version string) error
	TaintCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error)
	UnTaintCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error)
	RevokeCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error)
	UpdateCatalogManifestMetadata(ctx basecontext.ApiContext, catalogId, version, architecture string, description *string, tags, claims, roles []string) (*models.CatalogManifest, error)
	UpdateCatalogManifestProvider(ctx basecontext.ApiContext, recordId string, provider models.CatalogManifestProvider) error
}
