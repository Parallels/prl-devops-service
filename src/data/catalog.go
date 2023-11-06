package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
)

var (
	ErrCatalogManifestNotFound = errors.NewWithCode("catalog manifest not found", 404)
	ErrCatalogAlreadyExists    = errors.NewWithCode("Catalog Manifest already exists", 404)
)

func (j *JsonDatabase) GetCatalogManifests(ctx basecontext.ApiContext, filter string) ([]models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(j.data.ManifestsCatalog, dbFilter)
	if err != nil {
		return nil, err
	}

	result := GetAuthorizedRecords(ctx, filteredData...)
	return result, nil
}

func (j *JsonDatabase) GetCatalogManifest(ctx basecontext.ApiContext, idOrName string) (*models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, manifest := range catalogManifests {
		if strings.EqualFold(manifest.ID, helpers.NormalizeString(idOrName)) || strings.EqualFold(manifest.Name, idOrName) {
			return &manifest, nil
		}
	}

	return nil, ErrCatalogManifestNotFound
}

func (j *JsonDatabase) CreateCatalogManifest(ctx basecontext.ApiContext, manifest models.CatalogManifest) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}
	fmt.Println(manifest.ID)
	fmt.Println(manifest.Name)

	manifest.ID = helpers.NormalizeStringUpper(manifest.Name)
	if a, _ := j.GetCatalogManifest(ctx, manifest.ID); a != nil {
		fmt.Printf("a: %v\n", a)
		return ErrCatalogAlreadyExists
	}

	if a, _ := j.GetCatalogManifest(ctx, manifest.Name); a != nil {
		fmt.Printf("a: %v\n", a)
		return ErrCatalogAlreadyExists
	}

	// Checking the the required claims and roles exist
	for _, claim := range manifest.RequiredClaims {
		_, err := j.GetClaim(ctx, claim)
		if err != nil {
			return err
		}
	}
	for _, role := range manifest.RequiredRoles {
		_, err := j.GetRole(ctx, role)
		if err != nil {
			return err
		}
	}

	manifest.CreatedAt = helpers.GetUtcCurrentDateTime()
	manifest.UpdatedAt = helpers.GetUtcCurrentDateTime()

	j.data.ManifestsCatalog = append(j.data.ManifestsCatalog, manifest)

	j.Save(ctx)
	return nil
}

func (j *JsonDatabase) DeleteCatalogManifest(ctx basecontext.ApiContext, id string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if id == "" {
		return nil
	}

	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return err
	}

	for i, manifest := range catalogManifests {
		if strings.EqualFold(manifest.ID, id) {
			j.data.ManifestsCatalog = append(j.data.ManifestsCatalog[:i], j.data.ManifestsCatalog[i+1:]...)
			j.Save(ctx)
			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) UpdateCatalogManifest(ctx basecontext.ApiContext, record models.CatalogManifest) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return err
	}

	for i, manifest := range catalogManifests {
		if strings.EqualFold(manifest.ID, record.ID) {
			j.data.ManifestsCatalog[i].VirtualMachineContents = record.VirtualMachineContents
			j.data.ManifestsCatalog[i].PackContents = record.PackContents
			j.data.ManifestsCatalog[i].CreatedAt = record.CreatedAt
			j.data.ManifestsCatalog[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			j.data.ManifestsCatalog[i].LastDownloadedAt = record.LastDownloadedAt
			j.data.ManifestsCatalog[i].LastDownloadedUser = record.LastDownloadedUser
			j.data.ManifestsCatalog[i].Size = record.Size
			j.data.ManifestsCatalog[i].Path = record.Path
			j.data.ManifestsCatalog[i].MetadataFile = record.MetadataFile
			j.data.ManifestsCatalog[i].PackFile = record.PackFile
			j.data.ManifestsCatalog[i].Type = record.Type
			j.data.ManifestsCatalog[i].Tags = record.Tags
			j.data.ManifestsCatalog[i].RequiredClaims = record.RequiredClaims
			j.data.ManifestsCatalog[i].RequiredRoles = record.RequiredRoles

			j.Save(ctx)
			return nil
		}
	}

	return ErrCatalogManifestNotFound
}
