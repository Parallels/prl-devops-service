package catalog

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"

	"github.com/cjlapao/common-go/helper"
)

func (s *CatalogManifestService) PushMetadata(r *models.VirtualMachineCatalogManifest) *models.VirtualMachineCatalogManifest {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}
	executed := false
	manifest := r
	var err error
	connection := r.Provider.String()
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(s.ctx, connection)
		if checkErr != nil {
			s.ns.NotifyErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			manifest.AddError(checkErr)
			return manifest
		}

		if !check {
			continue
		}
		executed = true
		manifest.CleanupRequest.RemoteStorageService = rs
		apiClient := apiclient.NewHttpClient(s.ctx)

		if err := manifest.Provider.Parse(connection); err != nil {
			s.ns.NotifyErrorf("Error parsing provider %v: %v", connection, err)
			manifest.AddError(err)
			break
		}

		if manifest.Provider.IsRemote() {
			s.ns.NotifyDebugf("Testing remote provider %v", manifest.Provider.Host)
			apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
		}

		if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
			s.ns.NotifyErrorf("Error creating temp dir: %v", err)
		}

		// Checking if the manifest metadata exists in the remote server
		var catalogManifest *models.VirtualMachineCatalogManifest
		manifestPath := strings.ToLower(filepath.Join(rs.GetProviderRootPath(s.ctx), manifest.CatalogId))

		exists, _ := rs.FileExists(s.ctx, manifestPath, s.getMetaFilename(manifest.Name))
		if !exists {
			s.ns.NotifyInfof("Remote metadata does not exist, creating it")
			s.ns.NotifyErrorf("Error Remote metadata does not exist %v", manifest.CatalogId)
			manifest.AddError(err)
			break
		}

		if err := rs.PullFile(s.ctx, manifestPath, s.getMetaFilename(manifest.Name), "/tmp"); err != nil {
			s.ns.NotifyInfof("Error pulling remote metadata file %v", s.getMetaFilename(manifest.Name))
		}

		currentContent, err := helper.ReadFromFile(filepath.Join("/tmp", s.getMetaFilename(manifest.Name)))
		if err != nil {
			s.ns.NotifyErrorf("Error reading metadata file %v: %v", s.getMetaFilename(manifest.Name), err)
			manifest.AddError(err)
			break
		}

		if err := json.Unmarshal(currentContent, &catalogManifest); err != nil {
			s.ns.NotifyErrorf("Error unmarshalling metadata file %v: %v", s.getMetaFilename(manifest.Name), err)
			manifest.AddError(err)
			break
		}

		if catalogManifest == nil {
			s.ns.NotifyErrorf("Error unmarshalling metadata file %v: %v", s.getMetaFilename(manifest.Name), err)
			manifest.AddError(err)
			break
		}

		if err := helper.DeleteFile(filepath.Join("/tmp", s.getMetaFilename(manifest.Name))); err != nil {
			s.ns.NotifyErrorf("Error deleting metadata file %v: %v", s.getMetaFilename(manifest.Name), err)
			manifest.AddError(err)
			break
		}

		catalogManifest.RequiredClaims = r.RequiredClaims
		catalogManifest.RequiredRoles = r.RequiredRoles
		catalogManifest.Tags = r.Tags
		catalogManifest.Provider = nil

		tempManifestContentFilePath := filepath.Join("/tmp", catalogManifest.MetadataFile)
		manifestContent, err := json.MarshalIndent(catalogManifest, "", "  ")
		if err != nil {
			s.ns.NotifyErrorf("Error marshalling manifest %v: %v", manifest, err)
			manifest.AddError(err)
			break
		}

		manifest.CleanupRequest.AddLocalFileCleanupOperation(tempManifestContentFilePath, false)
		if err := helper.WriteToFile(string(manifestContent), tempManifestContentFilePath); err != nil {
			s.ns.NotifyErrorf("Error writing manifest to temporary file %v: %v", tempManifestContentFilePath, err)
			manifest.AddError(err)
			break
		}

		metadataChecksum, err := helpers.GetFileMD5Checksum(tempManifestContentFilePath)
		if err != nil {
			s.ns.NotifyErrorf("Error getting metadata checksum %v: %v", tempManifestContentFilePath, err)
			manifest.AddError(err)
			break
		}

		remoteMetadataChecksum, err := rs.FileChecksum(s.ctx, catalogManifest.Path, catalogManifest.MetadataFile)
		if err != nil {
			s.ns.NotifyErrorf("Error getting remote metadata checksum %v: %v", catalogManifest.MetadataFile, err)
			manifest.AddError(err)
			break
		}

		if remoteMetadataChecksum != metadataChecksum {
			s.ns.NotifyInfof("Remote metadata is not up to date, pushing it")
			if err := rs.PushFile(s.ctx, "/tmp", catalogManifest.Path, manifest.MetadataFile); err != nil {
				s.ns.NotifyErrorf("Error pushing metadata file %v: %v", catalogManifest.MetadataFile, err)
				manifest.AddError(err)
				break
			}
		} else {
			s.ns.NotifyInfof("Remote metadata is up to date")
		}

		if manifest.HasErrors() {
			manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.PackFile), false)
			manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.MetadataFile), false)
			manifest.CleanupRequest.AddRemoteFileCleanupOperation(manifest.Path, true)
		}
	}

	if !executed {
		manifest.AddError(errors.Newf("no remote service found for connection %v", connection))
	}

	if cleanErrors := manifest.CleanupRequest.Clean(s.ctx); len(cleanErrors) > 0 {
		s.ns.NotifyErrorf("Error cleaning up manifest %v", r.CatalogId)
		for _, err := range manifest.Errors {
			manifest.AddError(err)
		}
	}

	return manifest
}
