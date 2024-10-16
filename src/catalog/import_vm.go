package catalog

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/interfaces"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/data"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper"
)

type ImportVmManifestDetails struct {
	HasMetaFile      bool
	FilePath         string
	MetadataFilename string
	HasPackFile      bool
	MachineFilename  string
	MachineFileSize  int64
}

func (s *CatalogManifestService) ImportVm(ctx basecontext.ApiContext, r *models.ImportVmRequest) *models.ImportVmResponse {
	foundProvider := false
	response := models.NewImportVmRequestResponse()
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

		if !check {
			continue
		}

		foundProvider = true
		response.CleanupRequest.RemoteStorageService = rs
		var catalogManifest *models.VirtualMachineCatalogManifest
		fileDetails, err := s.checkForFiles(ctx, r, rs)
		if err != nil {
			response.AddError(err)
			break
		}
		if !fileDetails.HasPackFile {
			err := errors.Newf("pack file %v does not exist", r.CatalogId)
			response.AddError(err)
			break
		}

		if fileDetails.HasMetaFile {
			if r.Force {
				ctx.LogInfof("Force flag is set, removing existing manifest")
				if err := rs.DeleteFile(ctx, fileDetails.FilePath, fileDetails.MetadataFilename); err != nil {
					ctx.LogErrorf("Error deleting file %v: %v", fileDetails.MetadataFilename, err)
					response.AddError(err)
					break
				}
				catalogManifest = models.NewVirtualMachineCatalogManifest()

				catalogManifest.Name = fmt.Sprintf("%v-%v", r.CatalogId, r.Version)
				catalogManifest.Type = r.Type
				catalogManifest.Description = r.Description
				catalogManifest.RequiredClaims = r.RequiredClaims
				catalogManifest.RequiredRoles = r.RequiredRoles
				catalogManifest.Tags = r.Tags
			} else {
				ctx.LogInfof("Loading manifest from file %v", r.CatalogId)
				tmpCatalogManifestFilePath := filepath.Join("/tmp", fileDetails.MetadataFilename)
				response.CleanupRequest.AddLocalFileCleanupOperation(tmpCatalogManifestFilePath, false)
				if err := rs.PullFile(ctx, fileDetails.FilePath, fileDetails.MetadataFilename, "/tmp"); err != nil {
					ctx.LogErrorf("Error pulling file %v from remote service %v: %v", fileDetails.MetadataFilename, rs.Name(), err)
					response.AddError(err)
					break
				}
				catalogManifest, err = s.readManifestFromFile(tmpCatalogManifestFilePath)
				if err != nil {
					ctx.LogErrorf("Error reading manifest from file %v: %v", tmpCatalogManifestFilePath, err)
					response.AddError(err)
					break
				}

			}
		} else {
			catalogManifest = models.NewVirtualMachineCatalogManifest()
			catalogManifest.Name = fmt.Sprintf("%v-%v", r.CatalogId, r.Version)
			catalogManifest.Type = r.Type
			catalogManifest.Description = r.Description
			catalogManifest.RequiredClaims = r.RequiredClaims
			catalogManifest.RequiredRoles = r.RequiredRoles
			catalogManifest.Tags = r.Tags
		}

		ctx.LogInfof("Getting manifest from remote service %v", rs.Name())
		catalogManifest.Version = r.Version
		catalogManifest.CatalogId = r.CatalogId
		catalogManifest.Architecture = r.Architecture
		catalogManifest.Path = fileDetails.FilePath
		catalogManifest.PackRelativePath = fileDetails.MachineFilename
		catalogManifest.PackFile = fileDetails.MachineFilename

		catalogManifest.Provider = &models.CatalogManifestProvider{
			Type: provider.Type,
			Meta: provider.Meta,
		}

		if !strings.HasPrefix(catalogManifest.Path, "/") {
			catalogManifest.Path = "/" + catalogManifest.Path
		}
		catalogManifest.IsCompressed = r.IsCompressed
		vmChecksum, err := rs.FileChecksum(ctx, catalogManifest.Path, catalogManifest.PackRelativePath)
		if err != nil {
			ctx.LogErrorf("Error getting checksum for file %v: %v", catalogManifest.PackRelativePath, err)
			response.AddError(err)
			break
		}
		catalogManifest.CompressedChecksum = vmChecksum
		catalogManifest.Size = fileDetails.MachineFileSize
		catalogManifest.PackSize = fileDetails.MachineFileSize
		catalogManifest.MetadataFile = s.getMetaFilename(catalogManifest.Name)

		if err := catalogManifest.Validate(true); err != nil {
			ctx.LogErrorf("Error validating manifest: %v", err)
			response.AddError(err)
			break
		}

		exists, err := db.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogManifest.CatalogId, catalogManifest.Version, catalogManifest.Architecture)
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
			if claim == "" {
				continue
			}
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
			if role == "" {
				continue
			}
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
		catalogManifest.ID = cat.ID
		catalogManifest.Name = cat.Name
		catalogManifest.MetadataFile = s.getMetaFilename(catalogManifest.Name)
		catalogManifest.CreatedAt = cat.CreatedAt
		catalogManifest.UpdatedAt = cat.UpdatedAt

		metadataExists, err := rs.FileExists(ctx, catalogManifest.Path, catalogManifest.MetadataFile)
		if err != nil {
			ctx.LogErrorf("Error checking if meta file %v exists: %v", catalogManifest.MetadataFile, err)
			response.AddError(err)
			_ = db.DeleteCatalogManifest(ctx, cat.ID)
			break
		}

		if metadataExists {
			if err := rs.DeleteFile(ctx, catalogManifest.Path, catalogManifest.MetadataFile); err != nil {
				ctx.LogErrorf("Error deleting file %v: %v", catalogManifest.MetadataFile, err)
				response.AddError(err)
				_ = db.DeleteCatalogManifest(ctx, cat.ID)
				break
			}
		}

		tempManifestContentFilePath := filepath.Join("/tmp", catalogManifest.MetadataFile)
		cleanManifest := catalogManifest
		cleanManifest.Provider = nil
		manifestContent, err := json.MarshalIndent(cleanManifest, "", "  ")
		if err != nil {
			ctx.LogErrorf("Error marshalling manifest %v: %v", cleanManifest, err)
			_ = db.DeleteCatalogManifest(ctx, cat.ID)
			response.AddError(err)
			break
		}

		response.CleanupRequest.AddLocalFileCleanupOperation(tempManifestContentFilePath, false)
		if err := helper.WriteToFile(string(manifestContent), tempManifestContentFilePath); err != nil {
			ctx.LogErrorf("Error writing manifest to temporary file %v: %v", tempManifestContentFilePath, err)
			_ = db.DeleteCatalogManifest(ctx, cat.ID)
			response.AddError(err)
			break
		}
		ctx.LogInfof("Pushing manifest meta file %v", catalogManifest.MetadataFile)
		if err := rs.PushFile(ctx, "/tmp", catalogManifest.Path, catalogManifest.MetadataFile); err != nil {
			ctx.LogErrorf("Error pushing file %v to remote service %v: %v", catalogManifest.MetadataFile, rs.Name(), err)
			_ = db.DeleteCatalogManifest(ctx, cat.ID)
			response.AddError(err)
			break
		}

		db.SaveNow(ctx)
	}

	if !foundProvider {
		err := errors.Newf("provider %v not found", provider.String())
		response.AddError(err)
	}

	// Cleaning up
	s.CleanImportVmRequest(ctx, r, response)

	return response
}

func (s *CatalogManifestService) checkForFiles(ctx basecontext.ApiContext, r *models.ImportVmRequest, rs interfaces.RemoteStorageService) (ImportVmManifestDetails, error) {
	result := ImportVmManifestDetails{
		FilePath:         filepath.Dir(r.MachineRemotePath),
		MetadataFilename: s.getMetaFilename(r.Name()),
		MachineFilename:  filepath.Base(r.MachineRemotePath),
	}
	if r.MachineRemotePath == "" {
		return result, errors.New("MachineRemotePath is required")
	}

	// Checking for pack file
	machineFileExists, err := rs.FileExists(ctx, result.FilePath, result.MachineFilename)
	if err != nil {
		ctx.LogErrorf("Error checking if pack file %v exists: %v", r.CatalogId, err)
		return result, err
	}

	result.HasPackFile = machineFileExists
	fileSize, err := rs.FileSize(ctx, result.FilePath, result.MachineFilename)
	if err != nil {
		ctx.LogErrorf("Error getting file size for %v: %v", result.MachineFilename, err)
		return result, err
	}
	result.MachineFileSize = fileSize

	metaExists, err := rs.FileExists(ctx, result.FilePath, result.MetadataFilename)
	if err != nil {
		ctx.LogErrorf("Error checking if meta file %v exists: %v", r.CatalogId, err)
		return result, err
	}

	result.HasMetaFile = metaExists
	return result, nil
}

func (s *CatalogManifestService) generateMetadata(r *models.ImportVmRequest, fd ImportVmManifestDetails) (*models.VirtualMachineCatalogManifest, error) {
	result := models.NewVirtualMachineCatalogManifest()
	result.Name = r.Name()
	result.CatalogId = r.CatalogId
	result.Path = fd.FilePath
	result.Architecture = r.Architecture
	result.Version = r.Version
	result.IsCompressed = r.IsCompressed

	return result, nil
}

func (s *CatalogManifestService) CleanImportVmRequest(ctx basecontext.ApiContext, r *models.ImportVmRequest, response *models.ImportVmResponse) {
	if cleanErrors := response.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		ctx.LogErrorf("Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	}
}
