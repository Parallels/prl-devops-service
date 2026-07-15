package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
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

func (s *CatalogManifestService) ImportVm(r *models.ImportVmRequest) *models.ImportVmResponse {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}

	foundProvider := false
	response := models.NewImportVmRequestResponse()
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
		catalogManifest, err := s.getCatalogManifest(r, rs, &provider)
		if err != nil {
			response.AddError(err)
			break
		}

		exists, err := db.GetCatalogManifestsByCatalogIdVersionAndArch(s.ctx, catalogManifest.CatalogId, catalogManifest.Version, catalogManifest.Architecture)
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

		if err := s.importClaims(dto, db); err != nil {
			response.AddError(err)
			break
		}

		if err := s.importRoles(dto, db); err != nil {
			response.AddError(err)
			break
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

		response.ID = cat.ID
		catalogManifest.ID = cat.ID
		catalogManifest.Name = cat.Name
		catalogManifest.MetadataFile = s.getMetaFilename(catalogManifest.Name)
		catalogManifest.CreatedAt = cat.CreatedAt
		catalogManifest.UpdatedAt = cat.UpdatedAt

		if err := s.pushNewCatalogManifest(catalogManifest, rs); err != nil {
			response.AddError(err)
			_ = db.DeleteCatalogManifest(s.ctx, cat.ID)
			break
		}

		db.SaveNow(s.ctx)
	}

	if !foundProvider {
		err := errors.Newf("provider %v not found", provider.String())
		response.AddError(err)
	}

	// Cleaning up
	s.cleanImportVmRequest(s.ctx, response)

	return response
}

func (s *CatalogManifestService) checkForFiles(r *models.ImportVmRequest, rs interfaces.RemoteStorageService) (models.ImportVmManifestDetails, error) {
	result := models.ImportVmManifestDetails{
		FilePath:         filepath.Dir(r.MachineRemotePath),
		MetadataFilename: s.getMetaFilename(r.Name()),
		MachineFilename:  filepath.Base(r.MachineRemotePath),
	}
	if r.MachineRemotePath == "" {
		return result, errors.New("MachineRemotePath is required")
	}

	// Checking for pack file
	machineFileExists, err := rs.FileExists(s.ctx, result.FilePath, result.MachineFilename)
	if err != nil {
		s.ns.NotifyErrorf("Error checking if pack file %v exists: %v", r.CatalogId, err)
		return result, err
	}

	result.HasPackFile = machineFileExists
	fileSize, err := rs.FileSize(s.ctx, result.FilePath, result.MachineFilename)
	if err != nil {
		s.ns.NotifyErrorf("Error getting file size for %v: %v", result.MachineFilename, err)
		return result, err
	}
	result.MachineFileSize = fileSize

	metaExists, err := rs.FileExists(s.ctx, result.FilePath, result.MetadataFilename)
	if err != nil {
		s.ns.NotifyErrorf("Error checking if meta file %v exists: %v", r.CatalogId, err)
		return result, err
	}

	result.HasMetaFile = metaExists
	return result, nil
}

func (s *CatalogManifestService) generateMetadata(r *models.ImportVmRequest, fd models.ImportVmManifestDetails) (*models.VirtualMachineCatalogManifest, error) {
	result := models.NewVirtualMachineCatalogManifest()
	result.Name = r.Name()
	result.CatalogId = r.CatalogId
	result.Path = fd.FilePath
	result.Architecture = r.Architecture
	result.Version = r.Version
	result.IsCompressed = r.IsCompressed

	return result, nil
}

func (s *CatalogManifestService) getCatalogManifest(r *models.ImportVmRequest, rs interfaces.RemoteStorageService, provider *models.CatalogManifestProvider) (*models.VirtualMachineCatalogManifest, error) {
	var catalogManifest *models.VirtualMachineCatalogManifest
	fileDetails, err := s.checkForFiles(r, rs)
	if err != nil {
		return nil, err
	}
	if !fileDetails.HasPackFile {
		err := errors.Newf("vm file %v does not exist", r.CatalogId)
		return nil, err
	}

	if fileDetails.HasMetaFile {
		if r.Force {
			s.ns.NotifyInfof("Force flag is set, removing existing manifest")
			if err := rs.DeleteFile(s.ctx, fileDetails.FilePath, fileDetails.MetadataFilename); err != nil {
				s.ns.NotifyErrorf("Error deleting file %v: %v", fileDetails.MetadataFilename, err)
				return nil, err
			}
			catalogManifest = models.NewVirtualMachineCatalogManifest()

			catalogManifest.Name = r.Name()
			catalogManifest.Type = r.Type
			catalogManifest.Description = r.Description
			catalogManifest.RequiredClaims = r.RequiredClaims
			catalogManifest.RequiredRoles = r.RequiredRoles
			catalogManifest.Size = r.Size
			catalogManifest.Tags = r.Tags
		} else {
			s.ns.NotifyInfof("Loading manifest from file %v", r.CatalogId)
			content, err := rs.PullFileToMemory(s.ctx, fileDetails.FilePath, fileDetails.MetadataFilename)
			if err != nil {
				s.ns.NotifyErrorf("Error pulling file %v from remote service %v: %v", fileDetails.MetadataFilename, rs.Name(), err)
				return nil, err
			}
			catalogManifest, err = s.readManifestFromBytes(content)
			if err != nil {
				s.ns.NotifyErrorf("Error reading manifest from bytes: %v", err)
				return nil, err
			}
			catalogManifest.Size = r.Size
		}
	} else {
		catalogManifest = models.NewVirtualMachineCatalogManifest()
		catalogManifest.Name = r.Name()
		catalogManifest.Type = r.Type
		catalogManifest.Description = r.Description
		catalogManifest.RequiredClaims = r.RequiredClaims
		catalogManifest.RequiredRoles = r.RequiredRoles
		catalogManifest.Tags = r.Tags
		catalogManifest.Size = r.Size
	}

	s.ns.NotifyInfof("Getting manifest from remote service %v", rs.Name())
	catalogManifest.Version = r.Version
	catalogManifest.CatalogId = r.CatalogId
	catalogManifest.Architecture = r.Architecture
	catalogManifest.Path = fileDetails.FilePath
	catalogManifest.PackRelativePath = fileDetails.MachineFilename
	catalogManifest.PackFile = fileDetails.MachineFilename

	if !strings.HasPrefix(catalogManifest.Path, "/") {
		catalogManifest.Path = "/" + catalogManifest.Path
	}

	catalogManifest.IsCompressed = r.IsCompressed
	vmChecksum, err := rs.FileChecksum(s.ctx, catalogManifest.Path, catalogManifest.PackRelativePath)
	if err != nil {
		s.ns.NotifyErrorf("Error getting checksum for file %v: %v", catalogManifest.PackRelativePath, err)
		return nil, err
	}

	catalogManifest.CompressedChecksum = vmChecksum
	if catalogManifest.Size == 0 {
		catalogManifest.Size = fileDetails.MachineFileSize / (1024 * 1024) // in MB
	}
	catalogManifest.PackSize = fileDetails.MachineFileSize / (1024 * 1024) // in MB
	catalogManifest.MetadataFile = s.getMetaFilename(catalogManifest.Name)

	if err := catalogManifest.Validate(true); err != nil {
		s.ns.NotifyErrorf("Error validating manifest: %v", err)
		return nil, err
	}

	if provider != nil {
		catalogManifest.Provider = provider
	}

	return catalogManifest, nil
}

// pushNewCatalogManifest pushes a new catalog manifest to the remote storage service.
// It first checks if the metadata file already exists in the remote storage and deletes it if it does.
// Then, it creates a temporary file with the manifest content, writes the content to the file,
// and pushes the file to the remote storage service.
//
// Parameters:
//   - catalogManifest: A pointer to the VirtualMachineCatalogManifest model containing the manifest data.
//   - rs: An implementation of the RemoteStorageService interface for interacting with remote storage.
//
// Returns:
//   - error: An error if any operation fails, otherwise nil.
func (s *CatalogManifestService) pushNewCatalogManifest(catalogManifest *models.VirtualMachineCatalogManifest, rs interfaces.RemoteStorageService) error {
	cleanupSvc := cleanupservice.NewCleanupService()
	metadataExists, err := rs.FileExists(s.ctx, catalogManifest.Path, catalogManifest.MetadataFile)
	if err != nil {
		s.ns.NotifyErrorf("Error checking if meta file %v exists: %v", catalogManifest.MetadataFile, err)
		return err
	}

	if metadataExists {
		if err := rs.DeleteFile(s.ctx, catalogManifest.Path, catalogManifest.MetadataFile); err != nil {
			s.ns.NotifyErrorf("Error deleting file %v: %v", catalogManifest.MetadataFile, err)
			return err
		}
	}

	tempFolder := os.TempDir()
	tempManifestContentFilePath := filepath.Join(tempFolder, catalogManifest.MetadataFile)
	cleanManifest := catalogManifest
	cleanManifest.Provider = nil
	manifestContent, err := json.MarshalIndent(cleanManifest, "", "  ")
	if err != nil {
		s.ns.NotifyErrorf("Error marshalling manifest %v: %v", cleanManifest, err)
		return err
	}

	cleanupSvc.AddLocalFileCleanupOperation(tempManifestContentFilePath, false)
	if err := helper.WriteToFile(string(manifestContent), tempManifestContentFilePath); err != nil {
		s.ns.NotifyErrorf("Error writing manifest to temporary file %v: %v", tempManifestContentFilePath, err)
		return err
	}

	s.ns.NotifyInfof("Pushing manifest meta file %v", catalogManifest.MetadataFile)
	if err := rs.PushFile(s.ctx, tempFolder, catalogManifest.Path, catalogManifest.MetadataFile); err != nil {
		s.ns.NotifyErrorf("Error pushing file %v to remote service %v: %v", catalogManifest.MetadataFile, rs.Name(), err)
		return err
	}

	cleanupSvc.Clean(s.ctx)
	return nil
}

func (s *CatalogManifestService) importClaims(dto data_models.CatalogManifest, db *data.JsonDatabase) error {
	// Importing claims and roles
	for _, claim := range dto.RequiredClaims {
		if claim == "" {
			continue
		}
		exists, err := db.GetClaim(s.ctx, claim)
		if err != nil {
			if errors.GetSystemErrorCode(err) != 404 {
				s.ns.NotifyErrorf("Error getting claim %v: %v", claim, err)
				return err
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
				return err
			}
		}
	}

	return nil
}

func (s *CatalogManifestService) importRoles(dto data_models.CatalogManifest, db *data.JsonDatabase) error {
	for _, role := range dto.RequiredRoles {
		if role == "" {
			continue
		}
		exists, err := db.GetRole(s.ctx, role)
		if err != nil {
			if errors.GetSystemErrorCode(err) != 404 {
				s.ns.NotifyErrorf("Error getting role %v: %v", role, err)
				return err
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
				return err
			}
		}
	}

	return nil
}

func (s *CatalogManifestService) cleanImportVmRequest(ctx basecontext.ApiContext, response *models.ImportVmResponse) {
	if cleanErrors := response.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		s.ns.NotifyErrorf("Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	}
}
