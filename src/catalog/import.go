package catalog

import (
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/data"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *CatalogManifestService) Import(r *models.ImportCatalogManifestRequest) *models.ImportCatalogManifestResponse {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}
	foundProvider := false
	response := models.NewImportCatalogManifestResponse()
	serviceProvider := serviceprovider.Get()
	db := serviceProvider.JsonDatabase
	if db == nil {
		err := data.ErrDatabaseNotConnected
		response.AddError(err)
		return response
	}
	if err := db.Connect(s.ctx); err != nil {
		response.AddError(err)
		return response
	}

	if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
		s.ns.NotifyErrorf("Error creating temp dir: %v", err)
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
		check, checkErr := rs.Check(s.ctx, provider.String())
		if checkErr != nil {
			s.ns.NotifyErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			response.AddError(checkErr)
			break
		}

		if !check {
			continue
		}

		foundProvider = true
		response.CleanupRequest.RemoteStorageService = rs
		dir := strings.ToLower(r.CatalogId)
		metaFileName := s.getMetaFilename(r.Name())
		packFileName := s.getPackFilename(r.Name())
		metaExists, err := rs.FileExists(s.ctx, dir, metaFileName)
		if err != nil {
			s.ns.NotifyErrorf("Error checking if meta file %v exists: %v", r.CatalogId, err)
			response.AddError(err)
			break
		}
		if !metaExists {
			err := errors.Newf("meta file %v does not exist", r.CatalogId)
			response.AddError(err)
			break
		}
		packExists, err := rs.FileExists(s.ctx, dir, packFileName)
		if err != nil {
			s.ns.NotifyErrorf("Error checking if pack file %v exists: %v", r.CatalogId, err)
			response.AddError(err)
			break
		}
		if !packExists {
			err := errors.Newf("pack file %v does not exist", r.CatalogId)
			response.AddError(err)
			break
		}

		s.ns.NotifyInfof("Getting manifest from remote service %v", rs.Name())
		if err := rs.PullFile(s.ctx, dir, metaFileName, "/tmp"); err != nil {
			s.ns.NotifyErrorf("Error pulling file %v from remote service %v: %v", r.CatalogId, rs.Name(), err)
			response.AddError(err)
			break
		}

		s.ns.NotifyInfof("Loading manifest from file %v", r.CatalogId)
		tmpCatalogManifestFilePath := filepath.Join("/tmp", metaFileName)
		response.CleanupRequest.AddLocalFileCleanupOperation(tmpCatalogManifestFilePath, false)
		catalogManifest, err := s.readManifestFromFile(tmpCatalogManifestFilePath)
		if err != nil {
			s.ns.NotifyErrorf("Error reading manifest from file %v: %v", tmpCatalogManifestFilePath, err)
			response.AddError(err)
			break
		}

		catalogManifest.Version = r.Version
		catalogManifest.CatalogId = r.CatalogId
		catalogManifest.Architecture = r.Architecture
		if err := catalogManifest.Validate(false); err != nil {
			s.ns.NotifyErrorf("Error validating manifest: %v", err)
			response.AddError(err)
			break
		}
		exists, err := db.GetCatalogManifestsByCatalogIdVersionAndArch(s.ctx, catalogManifest.Name, catalogManifest.Version, catalogManifest.Architecture)
		if err != nil {
			if errors.GetSystemErrorCode(err) != 404 {
				s.ns.NotifyErrorf("Error getting catalog manifest: %v", err)
				response.AddError(err)
				break
			}
		}
		if exists != nil {
			s.ns.NotifyErrorf("Catalog manifest already exists: %v", catalogManifest.Name)
			response.AddError(errors.Newf("Catalog manifest already exists: %v", catalogManifest.Name))
			break
		}

		dto := mappers.CatalogManifestToDto(*catalogManifest)
		dto.Provider = &data_models.CatalogManifestProvider{
			Type: provider.Type,
			Meta: provider.Meta,
		}

		// Importing claims and roles
		for _, claim := range dto.RequiredClaims {
			if claim == "" {
				continue
			}
			exists, err := db.GetClaim(s.ctx, claim)
			if err != nil {
				if errors.GetSystemErrorCode(err) != 404 {
					s.ns.NotifyErrorf("Error getting claim %v: %v", claim, err)
					response.AddError(err)
					break
				}
			}
			if exists == nil {
				s.ns.NotifyInfof("Creating claim %v", claim)
				newClaim := data_models.Claim{
					ID:   claim,
					Name: claim,
				}
				if _, err := db.CreateClaim(s.ctx, newClaim); err != nil {
					s.ns.NotifyErrorf("Error creating claim %v: %v", claim, err)
					response.AddError(err)
					break
				}
			}
		}
		for _, role := range dto.RequiredRoles {
			if role == "" {
				continue
			}
			exists, err := db.GetRole(s.ctx, role)
			if err != nil {
				if errors.GetSystemErrorCode(err) != 404 {
					s.ns.NotifyErrorf("Error getting role %v: %v", role, err)
					response.AddError(err)
					break
				}
			}
			if exists == nil {
				s.ns.NotifyInfof("Creating role %v", role)
				newRole := data_models.Role{
					ID:   role,
					Name: role,
				}
				if _, err := db.CreateRole(s.ctx, newRole); err != nil {
					s.ns.NotifyErrorf("Error creating role %v: %v", role, err)
					response.AddError(err)
					break
				}
			}
		}

		result, err := db.CreateCatalogManifest(s.ctx, dto)
		if err != nil {
			s.ns.NotifyErrorf("Error creating catalog manifest: %v", err)
			response.AddError(err)
			break
		}

		cat, err := db.GetCatalogManifestByName(s.ctx, result.ID)
		if err != nil {
			s.ns.NotifyErrorf("Error getting catalog manifest: %v", err)
			response.AddError(err)
			break
		}

		db.SaveNow(s.ctx)
		response.ID = cat.ID
	}

	if !foundProvider {
		err := errors.Newf("provider %v not found", provider.String())
		response.AddError(err)
	}

	// Cleaning up
	s.CleanImportRequest(s.ctx, r, response)

	return response
}

func (s *CatalogManifestService) CleanImportRequest(ctx basecontext.ApiContext, r *models.ImportCatalogManifestRequest, response *models.ImportCatalogManifestResponse) {
	if cleanErrors := response.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		s.ns.NotifyErrorf("Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	}
}
