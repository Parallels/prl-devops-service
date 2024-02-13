package catalog

import (
	"path/filepath"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/data"
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/serviceprovider"
)

func (s *CatalogManifestService) Import(ctx basecontext.ApiContext, r *models.ImportCatalogManifestRequest) *models.ImportCatalogManifestResponse {
	foundProvider := false
	response := models.NewImportCatalogManifestResponse()
	serviceProvider := serviceprovider.Get()
	db := serviceProvider.JsonDatabase
	if db == nil {
		err := data.ErrDatabaseNotConnected
		response.AddError(err)
		return response
	}
	if err := db.Connect(ctx); err != nil {
		response.AddError(err)
		return response
	}

	if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
		ctx.LogErrorf("Error creating temp dir: %v", err)
		response.AddError(err)
		return response
	}

	provider := models.CatalogManifestProvider{}
	if err := provider.Parse(r.Connection); err != nil {
		response.AddError(err)
		return response
	}

	if provider.IsRemote() {
		err := errors.New("remote providers are not supported")
		response.AddError(err)
		return response
	}

	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(ctx, provider.String())
		if checkErr != nil {
			ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			response.AddError(checkErr)
			break
		}

		if check {
			foundProvider = true
			response.CleanupRequest.RemoteStorageService = rs
			dir := strings.ToLower(r.CatalogId)
			metaFileName := s.getMetaFilename(r.Name())
			packFileName := s.getPackFilename(r.Name())
			metaExists, err := rs.FileExists(ctx, dir, metaFileName)
			if err != nil {
				ctx.LogErrorf("Error checking if meta file %v exists: %v", r.CatalogId, err)
				response.AddError(err)
				break
			}
			if !metaExists {
				err := errors.Newf("meta file %v does not exist", r.CatalogId)
				response.AddError(err)
				break
			}
			packExists, err := rs.FileExists(ctx, dir, packFileName)
			if err != nil {
				ctx.LogErrorf("Error checking if pack file %v exists: %v", r.CatalogId, err)
				response.AddError(err)
				break
			}
			if !packExists {
				err := errors.Newf("pack file %v does not exist", r.CatalogId)
				response.AddError(err)
				break
			}

			ctx.LogInfof("Getting manifest from remote service %v", rs.Name())
			if err := rs.PullFile(ctx, dir, metaFileName, "/tmp"); err != nil {
				ctx.LogErrorf("Error pulling file %v from remote service %v: %v", r.CatalogId, rs.Name(), err)
				response.AddError(err)
				break
			}

			ctx.LogInfof("Loading manifest from file %v", r.CatalogId)
			tmpCatalogManifestFilePath := filepath.Join("/tmp", metaFileName)
			response.CleanupRequest.AddLocalFileCleanupOperation(tmpCatalogManifestFilePath, false)
			catalogManifest, err := s.readManifestFromFile(tmpCatalogManifestFilePath)
			if err != nil {
				ctx.LogErrorf("Error reading manifest from file %v: %v", tmpCatalogManifestFilePath, err)
				response.AddError(err)
				break
			}

			catalogManifest.Version = r.Version
			catalogManifest.CatalogId = r.CatalogId
			catalogManifest.Architecture = r.Architecture
			if err := catalogManifest.Validate(); err != nil {
				ctx.LogErrorf("Error validating manifest: %v", err)
				response.AddError(err)
				break
			}
			exists, err := db.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogManifest.Name, catalogManifest.Version, catalogManifest.Architecture)
			if err != nil {
				if errors.GetSystemErrorCode(err) != 404 {
					ctx.LogErrorf("Error getting catalog manifest: %v", err)
					response.AddError(err)
					break
				}
			}
			if exists != nil {
				ctx.LogErrorf("Catalog manifest already exists: %v", catalogManifest.Name)
				response.AddError(errors.Newf("Catalog manifest already exists: %v", catalogManifest.Name))
				break
			}

			dto := mappers.CatalogManifestToDto(*catalogManifest)

			// Importing claims and roles
			for _, claim := range dto.RequiredClaims {
				exists, err := db.GetClaim(ctx, claim)
				if err != nil {
					if errors.GetSystemErrorCode(err) != 404 {
						ctx.LogErrorf("Error getting claim %v: %v", claim, err)
						response.AddError(err)
						break
					}
				}
				if exists == nil {
					ctx.LogInfof("Creating claim %v", claim)
					newClaim := data_models.Claim{
						ID:   claim,
						Name: claim,
					}
					if _, err := db.CreateClaim(ctx, newClaim); err != nil {
						ctx.LogErrorf("Error creating claim %v: %v", claim, err)
						response.AddError(err)
						break
					}
				}
			}
			for _, role := range dto.RequiredRoles {
				exists, err := db.GetRole(ctx, role)
				if err != nil {
					if errors.GetSystemErrorCode(err) != 404 {
						ctx.LogErrorf("Error getting role %v: %v", role, err)
						response.AddError(err)
						break
					}
				}
				if exists == nil {
					ctx.LogInfof("Creating role %v", role)
					newRole := data_models.Role{
						ID:   role,
						Name: role,
					}
					if _, err := db.CreateRole(ctx, newRole); err != nil {
						ctx.LogErrorf("Error creating role %v: %v", role, err)
						response.AddError(err)
						break
					}
				}
			}

			result, err := db.CreateCatalogManifest(ctx, dto)
			if err != nil {
				ctx.LogErrorf("Error creating catalog manifest: %v", err)
				response.AddError(err)
				break
			}

			cat, err := db.GetCatalogManifestByName(ctx, result.ID)
			if err != nil {
				ctx.LogErrorf("Error getting catalog manifest: %v", err)
				response.AddError(err)
				break
			}

			response.ID = cat.ID
		}
	}

	if !foundProvider {
		err := errors.Newf("provider %v not found", provider.String())
		response.AddError(err)
	}

	// Cleaning up
	s.CleanImportRequest(ctx, r, response)

	return response
}

func (s *CatalogManifestService) CleanImportRequest(ctx basecontext.ApiContext, r *models.ImportCatalogManifestRequest, response *models.ImportCatalogManifestResponse) {
	if cleanErrors := response.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		ctx.LogErrorf("Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	}
}
