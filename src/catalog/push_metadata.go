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

func (s *CatalogManifestService) PushMetadata(ctx basecontext.ApiContext, r *models.VirtualMachineCatalogManifest) *models.VirtualMachineCatalogManifest {
	executed := false
	manifest := r
	var err error
	connection := r.Provider.String()
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(ctx, connection)
		if checkErr != nil {
			ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			manifest.AddError(checkErr)
			return manifest
		}

		if check {
			executed = true
			manifest.CleanupRequest.RemoteStorageService = rs
			apiClient := apiclient.NewHttpClient(ctx)

			if err := manifest.Provider.Parse(connection); err != nil {
				ctx.LogErrorf("Error parsing provider %v: %v", connection, err)
				manifest.AddError(err)
				break
			}

			if manifest.Provider.IsRemote() {
				ctx.LogDebugf("Testing remote provider %v", manifest.Provider.Host)
				apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
			}

			if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
				ctx.LogErrorf("Error creating temp dir: %v", err)
			}

			// Checking if the manifest metadata exists in the remote server
			var catalogManifest *models.VirtualMachineCatalogManifest
			manifestPath := strings.ToLower(filepath.Join(rs.GetProviderRootPath(ctx), manifest.CatalogId))

			exists, _ := rs.FileExists(ctx, manifestPath, s.getMetaFilename(manifest.Name))
			if !exists {
				ctx.LogInfof("Remote metadata does not exist, creating it")
				ctx.LogErrorf("Error Remote metadata does not exist %v", manifest.CatalogId)
				manifest.AddError(err)
				break
			}

			if err := rs.PullFile(ctx, manifestPath, s.getMetaFilename(manifest.Name), "/tmp"); err != nil {
				ctx.LogInfof("Error pulling remote metadata file %v", s.getMetaFilename(manifest.Name))
			}

			currentContent, err := helper.ReadFromFile(filepath.Join("/tmp", s.getMetaFilename(manifest.Name)))
			if err != nil {
				ctx.LogErrorf("Error reading metadata file %v: %v", s.getMetaFilename(manifest.Name), err)
				manifest.AddError(err)
				break
			}

			if err := json.Unmarshal(currentContent, &catalogManifest); err != nil {
				ctx.LogErrorf("Error unmarshalling metadata file %v: %v", s.getMetaFilename(manifest.Name), err)
				manifest.AddError(err)
				break
			}

			if catalogManifest == nil {
				ctx.LogErrorf("Error unmarshalling metadata file %v: %v", s.getMetaFilename(manifest.Name), err)
				manifest.AddError(err)
				break
			}

			if err := helper.DeleteFile(filepath.Join("/tmp", s.getMetaFilename(manifest.Name))); err != nil {
				ctx.LogErrorf("Error deleting metadata file %v: %v", s.getMetaFilename(manifest.Name), err)
				manifest.AddError(err)
				break
			}

			catalogManifest.RequiredClaims = r.RequiredClaims
			catalogManifest.RequiredRoles = r.RequiredRoles
			catalogManifest.Tags = r.Tags

			tempManifestContentFilePath := filepath.Join("/tmp", catalogManifest.MetadataFile)
			manifestContent, err := json.MarshalIndent(catalogManifest, "", "  ")
			if err != nil {
				ctx.LogErrorf("Error marshalling manifest %v: %v", manifest, err)
				manifest.AddError(err)
				break
			}

			manifest.CleanupRequest.AddLocalFileCleanupOperation(tempManifestContentFilePath, false)
			if err := helper.WriteToFile(string(manifestContent), tempManifestContentFilePath); err != nil {
				ctx.LogErrorf("Error writing manifest to temporary file %v: %v", tempManifestContentFilePath, err)
				manifest.AddError(err)
				break
			}

			metadataChecksum, err := helpers.GetFileMD5Checksum(tempManifestContentFilePath)
			if err != nil {
				ctx.LogErrorf("Error getting metadata checksum %v: %v", tempManifestContentFilePath, err)
				manifest.AddError(err)
				break
			}

			remoteMetadataChecksum, err := rs.FileChecksum(ctx, catalogManifest.Path, catalogManifest.MetadataFile)
			if err != nil {
				ctx.LogErrorf("Error getting remote metadata checksum %v: %v", catalogManifest.MetadataFile, err)
				manifest.AddError(err)
				break
			}

			if remoteMetadataChecksum != metadataChecksum {
				ctx.LogInfof("Remote metadata is not up to date, pushing it")
				if err := rs.PushFile(ctx, "/tmp", catalogManifest.Path, manifest.MetadataFile); err != nil {
					ctx.LogErrorf("Error pushing metadata file %v: %v", catalogManifest.MetadataFile, err)
					manifest.AddError(err)
					break
				}
			} else {
				ctx.LogInfof("Remote metadata is up to date")
			}

			if manifest.HasErrors() {
				manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.PackFile), false)
				manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.MetadataFile), false)
				manifest.CleanupRequest.AddRemoteFileCleanupOperation(manifest.Path, true)
			}
		}
	}

	if !executed {
		manifest.AddError(errors.Newf("no remote service found for connection %v", connection))
	}

	if cleanErrors := manifest.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		ctx.LogErrorf("Error cleaning up manifest %v", r.CatalogId)
		for _, err := range manifest.Errors {
			manifest.AddError(err)
		}
	}

	return manifest
}
