package data

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
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

	orderedResult, err := OrderByProperty(result, &Order{Property: "UpdatedAt", Direction: OrderDirectionDesc})
	if err != nil {
		return nil, err
	}

	return orderedResult, nil
}

func (j *JsonDatabase) GetCatalogManifestByName(ctx basecontext.ApiContext, idOrName string) (*models.CatalogManifest, error) {
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

func (j *JsonDatabase) GetCatalogManifestByTag(ctx basecontext.ApiContext, catalogId string, tag string) (*models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, manifest := range catalogManifests {
		if strings.EqualFold(manifest.CatalogId, helpers.NormalizeString(catalogId)) {
			for _, t := range manifest.Tags {
				if strings.EqualFold(t, tag) {
					return &manifest, nil
				}
			}
		}
	}

	return nil, ErrCatalogManifestNotFound
}

func (j *JsonDatabase) GetCatalogManifestsByCatalogId(ctx basecontext.ApiContext, catalogId string) ([]models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	result := make([]models.CatalogManifest, 0)
	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, manifest := range catalogManifests {
		if strings.EqualFold(manifest.ID, helpers.NormalizeString(catalogId)) ||
			strings.EqualFold(manifest.Name, helpers.NormalizeString(catalogId)) ||
			strings.EqualFold(manifest.CatalogId, helpers.NormalizeString(catalogId)) {
			result = append(result, manifest)
		}
	}

	return result, nil
}

func (j *JsonDatabase) GetCatalogManifestsByCatalogIdAndVersion(ctx basecontext.ApiContext, catalogId string, version string) ([]models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	result := make([]models.CatalogManifest, 0)
	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, manifest := range catalogManifests {
		if (strings.EqualFold(manifest.ID, helpers.NormalizeString(catalogId)) ||
			strings.EqualFold(manifest.CatalogId, helpers.NormalizeString(catalogId)) ||
			strings.EqualFold(manifest.Name, helpers.NormalizeString(catalogId))) &&
			strings.EqualFold(manifest.Version, version) {
			result = append(result, manifest)
		}
	}

	if len(result) == 0 {
		return nil, ErrCatalogManifestNotFound
	}

	return result, nil
}

func (j *JsonDatabase) GetCatalogManifestsByCatalogIdVersionAndArch(ctx basecontext.ApiContext, catalogId string, version string, arch string) (*models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, manifest := range catalogManifests {
		if (strings.EqualFold(manifest.ID, helpers.NormalizeString(catalogId)) ||
			strings.EqualFold(manifest.CatalogId, helpers.NormalizeString(catalogId)) ||
			strings.EqualFold(manifest.Name, helpers.NormalizeString(catalogId))) &&
			strings.EqualFold(manifest.Version, version) &&
			strings.EqualFold(manifest.Architecture, arch) {
			return &manifest, nil
		}
	}

	return nil, ErrCatalogManifestNotFound
}

func (j *JsonDatabase) CreateCatalogManifest(ctx basecontext.ApiContext, manifest models.CatalogManifest) (*models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}
	manifest.ID = helpers.GenerateId()
	if manifest.Version == "" {
		manifest.Version = constants.LATEST_TAG
	}

	manifest.Name = helpers.NormalizeString(manifest.Name)
	manifest.CatalogId = helpers.NormalizeStringUpper(manifest.CatalogId)
	manifest.AddTag(constants.LATEST_TAG)

	// Getting all of the siblings, we need to check for tag clashing
	siblings, err := j.GetCatalogManifestsByCatalogId(ctx, manifest.CatalogId)
	if err != nil {
		return nil, err
	}
	// The rule of the tag is simple, only one can exist per catalog item
	// If the current tag exists, we will remove the sibling tag
	// the manifest files will be overridden by the new one as the file name is the same
	for _, sibling := range siblings {
		if strings.EqualFold(sibling.Version, manifest.Version) {
			continue
		}
		if sibling.HasTag(constants.LATEST_TAG) {
			sibling.RemoveTag(constants.LATEST_TAG)
			if err := j.UpdateCatalogManifestTags(ctx, sibling); err != nil {
				return nil, err
			}
		}
	}

	exists, err := j.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, manifest.CatalogId, manifest.Version, manifest.Architecture)
	if err != nil {
		if errors.GetSystemErrorCode(err) != 404 {
			return nil, err
		}
	}

	if exists != nil {
		manifest.ID = exists.ID
		manifest.Name = exists.Name
		manifest.CatalogId = exists.CatalogId
		manifest.Architecture = exists.Architecture
		r, err := j.UpdateCatalogManifest(ctx, manifest)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// Checking the the required claims and roles exist
	for _, claim := range manifest.RequiredClaims {
		_, err := j.GetClaim(ctx, claim)
		if err != nil {
			return nil, err
		}
	}
	for _, role := range manifest.RequiredRoles {
		_, err := j.GetRole(ctx, role)
		if err != nil {
			return nil, err
		}
	}

	manifest.CreatedAt = helpers.GetUtcCurrentDateTime()
	manifest.UpdatedAt = helpers.GetUtcCurrentDateTime()

	j.data.ManifestsCatalog = append(j.data.ManifestsCatalog, manifest)
	return &manifest, nil
}

func (j *JsonDatabase) DeleteCatalogManifest(ctx basecontext.ApiContext, catalogIdOrId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if catalogIdOrId == "" {
		return ErrCatalogManifestNotFound
	}

	deletedItems := make([]string, 0)

	for {
		catalogManifests, err := j.GetCatalogManifests(ctx, "")
		if err != nil {
			return err
		}

		deletedSomething := false
		for _, manifest := range catalogManifests {
			if strings.EqualFold(manifest.ID, catalogIdOrId) ||
				strings.EqualFold(manifest.CatalogId, catalogIdOrId) ||
				strings.EqualFold(manifest.Name, catalogIdOrId) {
				index, err := GetRecordIndex(j.data.ManifestsCatalog, "id", manifest.ID)
				if err != nil {
					continue
				}
				j.data.ManifestsCatalog = append(j.data.ManifestsCatalog[:index], j.data.ManifestsCatalog[index+1:]...)
				deletedSomething = true
				deletedItems = append(deletedItems, catalogIdOrId)
				break
			}
		}

		if !deletedSomething {
			break
		}
	}

	if len(deletedItems) == 0 {
		return ErrCatalogManifestNotFound
	}

	return nil
}

func (j *JsonDatabase) DeleteCatalogManifestVersion(ctx basecontext.ApiContext, catalogIdOrId string, version string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if catalogIdOrId == "" {
		return ErrCatalogManifestNotFound
	}

	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return err
	}

	for _, manifest := range catalogManifests {
		i, err := GetRecordIndex(j.data.ManifestsCatalog, "id", manifest.ID)
		if err != nil {
			continue
		}

		if (strings.EqualFold(manifest.ID, catalogIdOrId) || strings.EqualFold(manifest.CatalogId, catalogIdOrId) || strings.EqualFold(manifest.Name, catalogIdOrId)) &&
			strings.EqualFold(manifest.Version, version) {
			j.data.ManifestsCatalog = append(j.data.ManifestsCatalog[:i], j.data.ManifestsCatalog[i+1:]...)
			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) DeleteCatalogManifestVersionArch(ctx basecontext.ApiContext, catalogIdOrId string, version string, architecture string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if catalogIdOrId == "" {
		return ErrCatalogManifestNotFound
	}

	catalogManifests, err := j.GetCatalogManifests(ctx, "")
	if err != nil {
		return err
	}

	for _, manifest := range catalogManifests {
		if (strings.EqualFold(manifest.ID, catalogIdOrId) || strings.EqualFold(manifest.CatalogId, catalogIdOrId) || strings.EqualFold(manifest.Name, catalogIdOrId)) &&
			strings.EqualFold(manifest.Version, version) &&
			strings.EqualFold(manifest.Architecture, architecture) {
			i, err := GetRecordIndex(j.data.ManifestsCatalog, "id", manifest.ID)
			if err != nil {
				continue
			}
			j.data.ManifestsCatalog = append(j.data.ManifestsCatalog[:i], j.data.ManifestsCatalog[i+1:]...)
			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) UpdateCatalogManifest(ctx basecontext.ApiContext, record models.CatalogManifest) (*models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, record.ID) {
			if !strings.EqualFold(manifest.Version, record.Version) {
				return nil, errors.Newf("cannot update version of catalog manifest %s, it is trying to change the version %s to version %s, not allowed", record.ID, j.data.ManifestsCatalog[i].Version, record.Version)
			}
			if !strings.EqualFold(manifest.Architecture, record.Architecture) {
				return nil, errors.Newf("cannot update architecture of catalog manifest %s, it is trying to change the version %s to version %s, not allowed", record.ID, j.data.ManifestsCatalog[i].Architecture, record.Architecture)
			}

			j.data.ManifestsCatalog[i].Name = record.Name
			j.data.ManifestsCatalog[i].VirtualMachineContents = record.VirtualMachineContents
			j.data.ManifestsCatalog[i].PackContents = record.PackContents
			j.data.ManifestsCatalog[i].CreatedAt = manifest.CreatedAt
			j.data.ManifestsCatalog[i].Architecture = record.Architecture
			j.data.ManifestsCatalog[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
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

			return &j.data.ManifestsCatalog[i], nil
		}
	}

	return nil, ErrCatalogManifestNotFound
}

func (j *JsonDatabase) UpdateCatalogManifestTags(ctx basecontext.ApiContext, record models.CatalogManifest) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, record.ID) || (strings.EqualFold(manifest.CatalogId, record.CatalogId) && strings.EqualFold(manifest.Version, record.Version)) {
			j.data.ManifestsCatalog[i].Tags = record.Tags

			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) UpdateCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, recordId) {
			found := false
			for _, role := range roles {
				for _, r := range manifest.RequiredRoles {
					if strings.EqualFold(r, role) {
						found = true
						break
					}
				}

				if !found {
					j.data.ManifestsCatalog[i].RequiredRoles = append(j.data.ManifestsCatalog[i].RequiredRoles, role)
				}
			}

			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) AddCatalogManifestRequiredClaims(ctx basecontext.ApiContext, recordId string, claims ...string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, recordId) {
			found := false
			for _, claim := range claims {
				claimExists := false
				for _, existClaim := range j.data.Claims {
					if strings.EqualFold(existClaim.ID, claim) {
						claimExists = true
						break
					}
				}
				if !claimExists {
					return errors.Newf("claim %s does not exist in the system", claim)
				}

				for _, r := range manifest.RequiredClaims {
					if strings.EqualFold(r, claim) {
						found = true
						break
					}
				}

				if !found {
					j.data.ManifestsCatalog[i].RequiredClaims = append(j.data.ManifestsCatalog[i].RequiredClaims, claim)
				}
			}

			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) RemoveCatalogManifestRequiredClaims(ctx basecontext.ApiContext, recordId string, claims ...string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, recordId) {
			for _, claim := range claims {
				foundAt := -1
				for index, r := range manifest.RequiredClaims {
					if strings.EqualFold(r, claim) {
						foundAt = index
						break
					}
				}

				if foundAt != -1 {
					j.data.ManifestsCatalog[i].RequiredClaims = append(j.data.ManifestsCatalog[i].RequiredClaims[:foundAt], j.data.ManifestsCatalog[i].RequiredClaims[foundAt+1:]...)
				}
			}
			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) AddCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, recordId) {
			found := false
			for _, role := range roles {
				roleExists := false
				for _, existRole := range j.data.Roles {
					if strings.EqualFold(existRole.ID, role) {
						roleExists = true
						break
					}
				}
				if !roleExists {
					return errors.Newf("role %s does not exist in the system", role)
				}

				for _, r := range manifest.RequiredRoles {
					if strings.EqualFold(r, role) {
						found = true
						break
					}
				}

				if !found {
					j.data.ManifestsCatalog[i].RequiredRoles = append(j.data.ManifestsCatalog[i].RequiredRoles, role)
				}
			}

			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) RemoveCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, recordId) {
			for _, role := range roles {
				foundAt := -1
				for index, r := range manifest.RequiredRoles {
					if strings.EqualFold(r, role) {
						foundAt = index
						break
					}
				}

				if foundAt != -1 {
					j.data.ManifestsCatalog[i].RequiredRoles = append(j.data.ManifestsCatalog[i].RequiredRoles[:foundAt], j.data.ManifestsCatalog[i].RequiredRoles[foundAt+1:]...)
				}
			}

			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) AddCatalogManifestTags(ctx basecontext.ApiContext, recordId string, tags ...string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, recordId) {
			found := false
			for _, tag := range tags {
				for _, r := range manifest.Tags {
					if strings.EqualFold(r, tag) {
						found = true
						break
					}
				}

				if !found {
					j.data.ManifestsCatalog[i].Tags = append(j.data.ManifestsCatalog[i].Tags, tag)
				}
			}

			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) RemoveCatalogManifestTags(ctx basecontext.ApiContext, recordId string, tags ...string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, recordId) {
			for _, tag := range tags {
				foundAt := -1
				for index, r := range manifest.Tags {
					if strings.EqualFold(r, tag) {
						foundAt = index
						break
					}
				}

				if foundAt != -1 {
					j.data.ManifestsCatalog[i].Tags = append(j.data.ManifestsCatalog[i].Tags[:foundAt], j.data.ManifestsCatalog[i].Tags[foundAt+1:]...)
				}
			}

			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) UpdateCatalogManifestDownloadCount(ctx basecontext.ApiContext, catalogId, version string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	downloadUser := "unknown"
	authContext := ctx.GetAuthorizationContext()
	if authContext.User != nil && authContext.User.Username != "" {
		downloadUser = authContext.User.Username
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.CatalogId, catalogId) && strings.EqualFold(manifest.Version, version) {
			j.data.ManifestsCatalog[i].LastDownloadedAt = helpers.GetUtcCurrentDateTime()
			j.data.ManifestsCatalog[i].LastDownloadedUser = downloadUser
			j.data.ManifestsCatalog[i].DownloadCount = j.data.ManifestsCatalog[i].DownloadCount + 1

			return nil
		}
	}

	return ErrCatalogManifestNotFound
}

func (j *JsonDatabase) TaintCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	taintUser := "unknown"
	authContext := ctx.GetAuthorizationContext()
	if authContext.User != nil && authContext.User.Username != "" {
		taintUser = authContext.User.Username
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.CatalogId, catalogId) && strings.EqualFold(manifest.Version, version) {
			j.data.ManifestsCatalog[i].TaintedAt = helpers.GetUtcCurrentDateTime()
			j.data.ManifestsCatalog[i].Tainted = true
			j.data.ManifestsCatalog[i].TaintedBy = taintUser

			return &j.data.ManifestsCatalog[i], nil
		}
	}

	return nil, ErrCatalogManifestNotFound
}

func (j *JsonDatabase) UnTaintCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	unTaintUser := "unknown"
	authContext := ctx.GetAuthorizationContext()
	if authContext.User != nil && authContext.User.Username != "" {
		unTaintUser = authContext.User.Username
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.CatalogId, catalogId) && strings.EqualFold(manifest.Version, version) {
			j.data.ManifestsCatalog[i].TaintedAt = ""
			j.data.ManifestsCatalog[i].Tainted = false
			j.data.ManifestsCatalog[i].UnTaintedBy = unTaintUser
			j.data.ManifestsCatalog[i].TaintedBy = ""

			return &j.data.ManifestsCatalog[i], nil
		}
	}

	return nil, ErrCatalogManifestNotFound
}

func (j *JsonDatabase) RevokeCatalogManifestVersion(ctx basecontext.ApiContext, catalogId string, version string) (*models.CatalogManifest, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	revokeUser := "unknown"
	authContext := ctx.GetAuthorizationContext()
	if authContext.User != nil && authContext.User.Username != "" {
		revokeUser = authContext.User.Username
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.CatalogId, catalogId) && strings.EqualFold(manifest.Version, version) {
			j.data.ManifestsCatalog[i].RevokedAt = helpers.GetUtcCurrentDateTime()
			j.data.ManifestsCatalog[i].Revoked = true
			j.data.ManifestsCatalog[i].RevokedBy = revokeUser

			return &j.data.ManifestsCatalog[i], nil
		}
	}

	return nil, ErrCatalogManifestNotFound
}

func (j *JsonDatabase) UpdateCatalogManifestProvider(ctx basecontext.ApiContext, recordId string, provider models.CatalogManifestProvider) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, manifest := range j.data.ManifestsCatalog {
		if strings.EqualFold(manifest.ID, recordId) {
			j.data.ManifestsCatalog[i].Provider = &provider
			return nil
		}
	}

	return ErrCatalogManifestNotFound
}
