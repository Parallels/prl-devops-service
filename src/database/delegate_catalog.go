package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/catalog delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetCatalogManifests(ctx basecontext.ApiContext, filter string) ([]models.CatalogManifest, error) {
	return g.legacy.GetCatalogManifests(ctx, filter)
}

func (g *GormDatabase) GetCatalogManifestByName(ctx basecontext.ApiContext, idOrName string) (*models.CatalogManifest, error) {
	return g.legacy.GetCatalogManifestByName(ctx, idOrName)
}

func (g *GormDatabase) GetCatalogManifestByTag(ctx basecontext.ApiContext, catalogId string, tag string) (*models.CatalogManifest, error) {
	return g.legacy.GetCatalogManifestByTag(ctx, catalogId, tag)
}

func (g *GormDatabase) GetCatalogManifestsByCatalogId(ctx basecontext.ApiContext, catalogId string) ([]models.CatalogManifest, error) {
	return g.legacy.GetCatalogManifestsByCatalogId(ctx, catalogId)
}

func (g *GormDatabase) GetCatalogManifestsByCatalogIdAndVersion(ctx basecontext.ApiContext, catalogId string, version string) ([]models.CatalogManifest, error) {
	return g.legacy.GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)
}

func (g *GormDatabase) GetCatalogManifestsByCatalogIdVersionAndArch(ctx basecontext.ApiContext, catalogId string, version string, arch string) (*models.CatalogManifest, error) {
	return g.legacy.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, arch)
}

func (g *GormDatabase) CreateCatalogManifest(ctx basecontext.ApiContext, manifest models.CatalogManifest) (*models.CatalogManifest, error) {
	return g.legacy.CreateCatalogManifest(ctx, manifest)
}

func (g *GormDatabase) DeleteCatalogManifest(ctx basecontext.ApiContext, catalogIdOrId string) error {
	return g.legacy.DeleteCatalogManifest(ctx, catalogIdOrId)
}

func (g *GormDatabase) DeleteCatalogManifestVersion(ctx basecontext.ApiContext, catalogIdOrId string, version string) error {
	return g.legacy.DeleteCatalogManifestVersion(ctx, catalogIdOrId, version)
}

func (g *GormDatabase) DeleteCatalogManifestVersionArch(ctx basecontext.ApiContext, catalogIdOrId string, version string, architecture string) error {
	return g.legacy.DeleteCatalogManifestVersionArch(ctx, catalogIdOrId, version, architecture)
}

func (g *GormDatabase) UpdateCatalogManifest(ctx basecontext.ApiContext, record models.CatalogManifest) (*models.CatalogManifest, error) {
	return g.legacy.UpdateCatalogManifest(ctx, record)
}

func (g *GormDatabase) UpdateCatalogManifestTags(ctx basecontext.ApiContext, record models.CatalogManifest) error {
	return g.legacy.UpdateCatalogManifestTags(ctx, record)
}

func (g *GormDatabase) UpdateCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	return g.legacy.UpdateCatalogManifestRequiredRoles(ctx, recordId, roles...)
}

func (g *GormDatabase) AddCatalogManifestRequiredClaims(ctx basecontext.ApiContext, recordId string, claims ...string) error {
	return g.legacy.AddCatalogManifestRequiredClaims(ctx, recordId, claims...)
}

func (g *GormDatabase) RemoveCatalogManifestRequiredClaims(ctx basecontext.ApiContext, recordId string, claims ...string) error {
	return g.legacy.RemoveCatalogManifestRequiredClaims(ctx, recordId, claims...)
}

func (g *GormDatabase) AddCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	return g.legacy.AddCatalogManifestRequiredRoles(ctx, recordId, roles...)
}

func (g *GormDatabase) RemoveCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	return g.legacy.RemoveCatalogManifestRequiredRoles(ctx, recordId, roles...)
}

func (g *GormDatabase) AddCatalogManifestTags(ctx basecontext.ApiContext, recordId string, tags ...string) error {
	return g.legacy.AddCatalogManifestTags(ctx, recordId, tags...)
}

func (g *GormDatabase) RemoveCatalogManifestTags(ctx basecontext.ApiContext, recordId string, tags ...string) error {
	return g.legacy.RemoveCatalogManifestTags(ctx, recordId, tags...)
}

func (g *GormDatabase) UpdateCatalogManifestDownloadCount(ctx basecontext.ApiContext, catalogId, version string) error {
	return g.legacy.UpdateCatalogManifestDownloadCount(ctx, catalogId, version)
}

func (g *GormDatabase) TaintCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	return g.legacy.TaintCatalogManifestVersion(ctx, catalogId, version)
}

func (g *GormDatabase) UnTaintCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	return g.legacy.UnTaintCatalogManifestVersion(ctx, catalogId, version)
}

func (g *GormDatabase) RevokeCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	return g.legacy.RevokeCatalogManifestVersion(ctx, catalogId, version)
}

func (g *GormDatabase) UpdateCatalogManifestMetadata(ctx basecontext.ApiContext, catalogId, version, architecture string, description *string, tags, claims, roles []string) (*models.CatalogManifest, error) {
	return g.legacy.UpdateCatalogManifestMetadata(ctx, catalogId, version, architecture, description, tags, claims, roles)
}

func (g *GormDatabase) UpdateCatalogManifestProvider(ctx basecontext.ApiContext, recordId string, provider models.CatalogManifestProvider) error {
	return g.legacy.UpdateCatalogManifestProvider(ctx, recordId, provider)
}
