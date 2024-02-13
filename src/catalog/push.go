package catalog

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	api_models "github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/Parallels/pd-api-service/serviceprovider/apiclient"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/helper/http_helper"
)

func (s *CatalogManifestService) Push(ctx basecontext.ApiContext, r *models.PushCatalogManifestRequest) *models.VirtualMachineCatalogManifest {
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
			ctx.LogInfof("Pushing manifest %v to provider %s", r.CatalogId, rs.Name())
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
				if err := rs.PullFile(ctx, manifestPath, s.getMetaFilename(manifest.Name), "/tmp"); err == nil {
					ctx.LogInfof("Remote Manifest metadata found, retrieving it")
					tmpCatalogManifestFilePath := filepath.Join("/tmp", s.getMetaFilename(manifest.Name))
					manifest.CleanupRequest.AddLocalFileCleanupOperation(tmpCatalogManifestFilePath, false)
					catalogManifest, err = s.readManifestFromFile(tmpCatalogManifestFilePath)

					if err != nil {
						ctx.LogErrorf("Error reading manifest from file %v: %v", tmpCatalogManifestFilePath, err)
						manifest.AddError(err)
						break
					}

					manifest.CreatedAt = catalogManifest.CreatedAt
					manifest.RequiredRoles = catalogManifest.RequiredRoles
					manifest.RequiredClaims = catalogManifest.RequiredClaims
				}
			}

			// Pushing the necessary files to the remote server
			if catalogManifest != nil {
				manifest.Path = catalogManifest.Path
				manifest.MetadataFile = s.getMetaFilename(catalogManifest.Name)
				manifest.PackFile = s.getPackFilename(catalogManifest.Name)
				localPackPath := filepath.Dir(manifest.CompressedPath)

				// The catalog manifest metadata already exists checking if the files are up to date and pushing them if not
				ctx.LogInfof("Found remote catalog manifest, checking if the files are up to date")
				remotePackChecksum, err := rs.FileChecksum(ctx, catalogManifest.Path, catalogManifest.PackFile)
				if err != nil {
					ctx.LogErrorf("Error getting remote pack checksum %v: %v", catalogManifest.PackFile, err)
					manifest.AddError(err)
					break
				}
				if remotePackChecksum != manifest.CompressedChecksum {
					ctx.LogInfof("Remote pack is not up to date, pushing it")
					if err := rs.PushFile(ctx, localPackPath, catalogManifest.Path, catalogManifest.PackFile); err != nil {
						ctx.LogErrorf("Error pushing pack file %v: %v", catalogManifest.PackFile, err)
						manifest.AddError(err)
						break
					}
				} else {
					ctx.LogInfof("Remote pack is up to date")
				}
				manifest.PackContents = append(manifest.PackContents, models.VirtualMachineManifestContentItem{
					Path:      manifest.Path,
					IsDir:     false,
					Name:      filepath.Base(manifest.PackFile),
					Checksum:  manifest.CompressedChecksum,
					CreatedAt: helpers.GetUtcCurrentDateTime(),
					UpdatedAt: helpers.GetUtcCurrentDateTime(),
				})

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

				manifest.PackContents = append(manifest.PackContents, models.VirtualMachineManifestContentItem{
					Path:      manifest.Path,
					IsDir:     false,
					Name:      filepath.Base(manifest.MetadataFile),
					Checksum:  metadataChecksum,
					CreatedAt: helpers.GetUtcCurrentDateTime(),
					UpdatedAt: helpers.GetUtcCurrentDateTime(),
				})

				if manifest.HasErrors() {
					manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.PackFile), false)
					manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.MetadataFile), false)
					manifest.CleanupRequest.AddRemoteFileCleanupOperation(manifest.Path, true)
				}

			} else {
				// The catalog manifest metadata does not exist creating it
				ctx.LogInfof("Remote Manifest metadata not found, creating it")

				manifest.Path = filepath.Join(rs.GetProviderRootPath(ctx), manifest.CatalogId)
				manifest.MetadataFile = s.getMetaFilename(manifest.Name)
				manifest.PackFile = s.getPackFilename(manifest.Name)
				tempManifestContentFilePath := filepath.Join("/tmp", s.getMetaFilename(manifest.Name))
				if manifest.Architecture == "amd64" {
					manifest.Architecture = "x86_64"
				}
				if r.Architecture == "arm" {
					manifest.Architecture = "arm64"
				}
				if manifest.Architecture == "aarch64" {
					manifest.Architecture = "arm64"
				}

				if err := rs.CreateFolder(ctx, "/", manifest.Path); err != nil {
					manifest.AddError(err)
					break
				}

				manifest.PackContents = append(manifest.PackContents, models.VirtualMachineManifestContentItem{
					Path:      manifest.Path,
					IsDir:     false,
					Name:      filepath.Base(manifest.MetadataFile),
					CreatedAt: helpers.GetUtcCurrentDateTime(),
					UpdatedAt: helpers.GetUtcCurrentDateTime(),
				})
				manifest.PackContents = append(manifest.PackContents, models.VirtualMachineManifestContentItem{
					Path:      manifest.Path,
					IsDir:     false,
					Name:      filepath.Base(manifest.PackFile),
					Checksum:  manifest.CompressedChecksum,
					CreatedAt: helpers.GetUtcCurrentDateTime(),
					UpdatedAt: helpers.GetUtcCurrentDateTime(),
				})

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

				ctx.LogInfof("Pushing manifest pack file %v", manifest.PackFile)
				localPackPath := filepath.Dir(manifest.CompressedPath)
				if err := rs.PushFile(ctx, localPackPath, manifest.Path, manifest.PackFile); err != nil {
					manifest.AddError(err)
					break
				}

				ctx.LogInfof("Pushing manifest meta file %v", manifest.MetadataFile)
				if err := rs.PushFile(ctx, "/tmp", manifest.Path, manifest.MetadataFile); err != nil {
					manifest.AddError(err)
					break
				}

				if err != nil {
					ctx.LogErrorf("Error getting metadata checksum %v: %v", tempManifestContentFilePath, err)
					manifest.AddError(err)
					break
				}

				if manifest.HasErrors() {
					manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.PackFile), false)
					manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.MetadataFile), false)
					manifest.CleanupRequest.AddRemoteFileCleanupOperation(manifest.Path, true)
				}
			}

			// Data has been pushed, checking if there is any error here if not let's add the manifest to the database or update it
			if !manifest.HasErrors() {
				if manifest.Provider.IsRemote() {
					ctx.LogInfof("Manifest pushed successfully, adding it to the remote database")
					apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
					path := http_helper.JoinUrl(constants.DEFAULT_API_PREFIX, "catalog")

					var response api_models.CatalogManifest
					postUrl := fmt.Sprintf("%s%s", manifest.Provider.GetUrl(), path)
					if _, err := apiClient.Post(postUrl, manifest, &response); err != nil {
						ctx.LogErrorf("Error posting catalog manifest %v: %v", manifest.Provider.String(), err)
						manifest.AddError(err)
						break
					}
				} else {
					ctx.LogInfof("Manifest pushed successfully, adding it to the database")
					db := serviceprovider.Get().JsonDatabase
					if err := db.Connect(ctx); err != nil {
						manifest.AddError(err)
						break
					}

					exists, _ := db.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, manifest.CatalogId, manifest.Version, manifest.Architecture)
					if exists != nil {
						ctx.LogInfof("Updating manifest %v", manifest.Name)
						dto := mappers.CatalogManifestToDto(*manifest)
						dto.ID = exists.ID
						if _, err := db.UpdateCatalogManifest(ctx, dto); err != nil {
							ctx.LogErrorf("Error updating manifest %v: %v", manifest.Name, err)
							manifest.AddError(err)
							break
						}
					} else {
						ctx.LogInfof("Creating manifest %v", manifest.Name)
						dto := mappers.CatalogManifestToDto(*manifest)
						if _, err := db.CreateCatalogManifest(ctx, dto); err != nil {
							ctx.LogErrorf("Error creating manifest %v: %v", manifest.Name, err)
							manifest.AddError(err)
							break
						}
					}
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
