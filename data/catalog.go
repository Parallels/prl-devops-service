package data

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/data/models"
	"context"
	"errors"
	"strings"
	"time"
)

func (j *JsonDatabase) GetCatalogManifests(ctx context.Context) ([]models.CatalogVirtualMachineManifest, error) {
	authContext := basecontext.GetAuthorizationContext(ctx)
	if authContext == nil {
		return nil, errors.New("no auth context")
	}
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	result := make([]models.CatalogVirtualMachineManifest, 0)
	for _, manifest := range j.data.ManifestsCatalog {
		if manifest.IsAuthorized(authContext) {
			result = append(result, manifest)
		}
	}

	return result, nil
}

func (j *JsonDatabase) GetCatalogManifest(ctx context.Context, idOrName string) (*models.CatalogVirtualMachineManifest, error) {
	authContext := basecontext.GetAuthorizationContext(ctx)
	if authContext == nil {
		return nil, errors.New("no auth context")
	}

	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	for _, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, idOrName) || strings.EqualFold(manifest.Name, idOrName) {
			if manifest.IsAuthorized(authContext) {
				return &manifest, nil
			}
		}
	}

	return nil, errors.New("catalog manifest not found")
}

func (j *JsonDatabase) CreateCatalogManifest(ctx context.Context, manifest *models.CatalogVirtualMachineManifest) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if a, _ := j.GetCatalogManifest(ctx, manifest.ID); a != nil {
		return errors.New("manifest already exists")
	}

	if a, _ := j.GetCatalogManifest(ctx, manifest.Name); a != nil {
		return errors.New("manifest already exists")
	}

	manifest.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	manifest.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	j.data.ManifestsCatalog = append(j.data.ManifestsCatalog, *manifest)
	j.save()

	return nil
}

func (j *JsonDatabase) RemoveCatalogManifest(ctx context.Context, id string) error {
	authContext := basecontext.GetAuthorizationContext(ctx)
	if authContext == nil {
		return errors.New("no auth context")
	}

	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if id == "" {
		return nil
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, id) {
			if manifest.IsAuthorized(authContext) {
				j.data.ManifestsCatalog = append(j.data.ManifestsCatalog[:i], j.data.ManifestsCatalog[i+1:]...)
				j.save()
				return nil
			} else {
				return errors.New("not authorized")
			}
		}
	}

	return errors.New("catalog Manifest not found")
}

func (j *JsonDatabase) UpdateCatalogManifest(ctx context.Context, record models.CatalogVirtualMachineManifest) error {
	authContext := basecontext.GetAuthorizationContext(ctx)
	if authContext == nil {
		return errors.New("no auth context")
	}
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, record.ID) {
			if !manifest.IsAuthorized(authContext) {
				return errors.New("not authorized")
			}

			j.data.ManifestsCatalog[i].Contents = record.Contents
			j.data.ManifestsCatalog[i].CreatedAt = record.CreatedAt
			j.data.ManifestsCatalog[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			j.data.ManifestsCatalog[i].LastDownloadedAt = record.LastDownloadedAt
			j.data.ManifestsCatalog[i].LastDownloadedUser = record.LastDownloadedUser
			j.data.ManifestsCatalog[i].Size = record.Size
			j.data.ManifestsCatalog[i].Path = record.Path
			j.data.ManifestsCatalog[i].MetadataPath = record.MetadataPath
			j.data.ManifestsCatalog[i].Type = record.Type
			j.data.ManifestsCatalog[i].Tags = record.Tags
			j.data.ManifestsCatalog[i].RequiredClaims = record.RequiredClaims
			j.data.ManifestsCatalog[i].RequiredRoles = record.RequiredRoles

			j.save()
			return nil
		}
	}

	return errors.New("catalog Manifest not found")
}
