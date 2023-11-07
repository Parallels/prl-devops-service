package catalog

import (
	"path/filepath"

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
		ctx.LogError("Error creating temp dir: %v", err)
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
			ctx.LogError("Error checking remote service %v: %v", rs.Name(), checkErr)
			response.AddError(checkErr)
			break
		}

		if check {
			foundProvider = true
			response.CleanupRequest.RemoteStorageService = rs
			dir := r.ID
			metaFileName := s.getMetaFilename(r.ID)
			packFileName := s.getPackFilename(r.ID)
			metaExists, err := rs.FileExists(ctx, dir, metaFileName)
			if err != nil {
				ctx.LogError("Error checking if meta file %v exists: %v", r.ID, err)
				response.AddError(err)
				break
			}
			if !metaExists {
				err := errors.Newf("meta file %v does not exist", r.ID)
				response.AddError(err)
				break
			}
			packExists, err := rs.FileExists(ctx, dir, packFileName)
			if err != nil {
				ctx.LogError("Error checking if pack file %v exists: %v", r.ID, err)
				response.AddError(err)
				break
			}
			if !packExists {
				err := errors.Newf("pack file %v does not exist", r.ID)
				response.AddError(err)
				break
			}

			ctx.LogInfo("Getting manifest from remote service %v", rs.Name())
			if err := rs.PullFile(ctx, dir, metaFileName, "/tmp"); err != nil {
				ctx.LogError("Error pulling file %v from remote service %v: %v", r.ID, rs.Name(), err)
				response.AddError(err)
				break
			}

			ctx.LogInfo("Loading manifest from file %v", r.ID)
			tmpCatalogManifestFilePath := filepath.Join("/tmp", metaFileName)
			response.CleanupRequest.AddLocalFileCleanupOperation(tmpCatalogManifestFilePath, false)
			catalogManifest, err := s.readManifestFromFile(tmpCatalogManifestFilePath)
			if err != nil {
				ctx.LogError("Error reading manifest from file %v: %v", tmpCatalogManifestFilePath, err)
				response.AddError(err)
				break
			}

			dto := mappers.CatalogManifestToDto(*catalogManifest)

			// Importing claims and roles
			for _, claim := range dto.RequiredClaims {
				exists, err := db.GetClaim(ctx, claim)
				if err != nil {
					if errors.GetSystemErrorCode(err) != 404 {
						ctx.LogError("Error getting claim %v: %v", claim, err)
						response.AddError(err)
						break
					}
				}
				if exists == nil {
					ctx.LogInfo("Creating claim %v", claim)
					newClaim := data_models.Claim{
						ID:   claim,
						Name: claim,
					}
					db.CreateClaim(ctx, newClaim)
				}
			}
			for _, role := range dto.RequiredRoles {
				exists, err := db.GetRole(ctx, role)
				if err != nil {
					if errors.GetSystemErrorCode(err) != 404 {
						ctx.LogError("Error getting role %v: %v", role, err)
						response.AddError(err)
						break
					}
				}
				if exists == nil {
					ctx.LogInfo("Creating role %v", role)
					newRole := data_models.Role{
						ID:   role,
						Name: role,
					}
					db.CreateRole(ctx, newRole)
				}
			}

			if err := db.CreateCatalogManifest(ctx, dto); err != nil {
				ctx.LogError("Error creating catalog manifest: %v", err)
				response.AddError(err)
				break
			}

			cat, err := db.GetCatalogManifest(ctx, dto.ID)
			if err != nil {
				ctx.LogError("Error getting catalog manifest: %v", err)
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

	//Cleaning up
	s.CleanImportRequest(ctx, r, response)

	return response
}

func (s *CatalogManifestService) CleanImportRequest(ctx basecontext.ApiContext, r *models.ImportCatalogManifestRequest, response *models.ImportCatalogManifestResponse) {
	if cleanErrors := response.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		ctx.LogError("Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	}
}
