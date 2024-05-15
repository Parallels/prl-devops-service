package catalog

import (
	"encoding/json"
	"path/filepath"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"

	"github.com/cjlapao/common-go/helper"
)

func (s *CatalogManifestService) PushMetadata(ctx basecontext.ApiContext, r *models.PushCatalogManifestRequest) *models.VirtualMachineCatalogManifest {
	executed := false
	manifest := models.NewVirtualMachineCatalogManifest()
	var err error
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(ctx, r.Connection)
		if checkErr != nil {
			ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			manifest.AddError(checkErr)
			return manifest
		}

		if check {
			executed = true
			manifest.CleanupRequest.RemoteStorageService = rs
			apiClient := apiclient.NewHttpClient(ctx)

			if err := manifest.Provider.Parse(r.Connection); err != nil {
				ctx.LogErrorf("Error parsing provider %v: %v", r.Connection, err)
				manifest.AddError(err)
				break
			}

			if manifest.Provider.IsRemote() {
				ctx.LogDebugf("Testing remote provider %v", manifest.Provider.Host)
				apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
			}

			// Generating the manifest content
			ctx.LogInfof("Pushing manifest metadata %v to provider %s", r.CatalogId, rs.Name())
			err = s.GenerateManifestContent(ctx, r, manifest)
			if err != nil {
				ctx.LogErrorf("Error generating manifest content for %v: %v", r.CatalogId, err)
				manifest.AddError(err)
				break
			}

			if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
				ctx.LogErrorf("Error creating temp dir: %v", err)
			}

			// Checking if the manifest metadata exists in the remote server
			var catalogManifest *models.VirtualMachineCatalogManifest
			manifestPath := filepath.Join(rs.GetProviderRootPath(ctx), manifest.CatalogId)
			exists, _ := rs.FileExists(ctx, manifestPath, s.getMetaFilename(manifest.Name))
			if exists {
				if err := rs.DeleteFile(ctx, manifestPath, s.getMetaFilename(manifest.Name)); err == nil {
					ctx.LogInfof("Error removing remote metadata file %v", s.getMetaFilename(manifest.Name))
				}

				// Pushing the metadata file to the remote server
				if catalogManifest != nil {
					manifest.Path = catalogManifest.Path
					manifest.MetadataFile = s.getMetaFilename(catalogManifest.Name)

					tempManifestContentFilePath := filepath.Join("/tmp", manifest.MetadataFile)
					manifestContent, err := json.MarshalIndent(manifest, "", "  ")
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
				} else {
					// The catalog manifest metadata does not exist creating it
					errNotFound := errors.New("Remote Manifest metadata not found")
					manifest.AddError(errNotFound)
					break
				}
			}
		}
	}

	if !executed {
		manifest.AddError(errors.Newf("no remote service found for connection %v", r.Connection))
	}

	if cleanErrors := manifest.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		ctx.LogErrorf("Error cleaning up manifest %v", r.CatalogId)
		for _, err := range manifest.Errors {
			manifest.AddError(err)
		}
	}

	return manifest
}
