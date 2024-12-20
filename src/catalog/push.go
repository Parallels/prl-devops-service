package catalog

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/helper/http_helper"
)

func (s *CatalogManifestService) Push(r *models.PushCatalogManifestRequest) *models.VirtualMachineCatalogManifest {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}
	executed := false
	manifest := models.NewVirtualMachineCatalogManifest()
	var err error
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(s.ctx, r.Connection)
		if checkErr != nil {
			s.ns.NotifyErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			manifest.AddError(checkErr)
			return manifest
		}

		if !check {
			continue
		}
		executed = true
		if r.ProgressChannel != nil {
			s.ns.NotifyDebugf("Setting progress channel for remote service %v", rs.Name())
			rs.SetProgressChannel(r.FileNameChannel, r.ProgressChannel)
		}
		manifest.CleanupRequest.RemoteStorageService = rs
		apiClient := apiclient.NewHttpClient(s.ctx)

		if err := manifest.Provider.Parse(r.Connection); err != nil {
			s.ns.NotifyErrorf("Error parsing provider %v: %v", r.Connection, err)
			manifest.AddError(err)
			break
		}

		if manifest.Provider.IsRemote() {
			s.ns.NotifyDebugf("Testing remote provider %v", manifest.Provider.Host)
			apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
		}

		// Generating the manifest content
		s.ns.NotifyInfof("Pushing manifest %v to provider %s", r.CatalogId, rs.Name())
		err = s.GenerateManifestContent(r, manifest)
		if err != nil {
			s.ns.NotifyErrorf("Error generating manifest content for %v: %v", r.CatalogId, err)
			manifest.AddError(err)
			break
		}

		if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
			s.ns.NotifyErrorf("Error creating temp dir: %v", err)
		}

		// Checking if the manifest metadata exists in the remote server
		var catalogManifest *models.VirtualMachineCatalogManifest
		manifestPath := filepath.Join(rs.GetProviderRootPath(s.ctx), manifest.CatalogId)
		exists, _ := rs.FileExists(s.ctx, manifestPath, s.getMetaFilename(manifest.Name))
		if exists {
			if err := rs.PullFile(s.ctx, manifestPath, s.getMetaFilename(manifest.Name), "/tmp"); err == nil {
				s.ns.NotifyInfof("Remote Manifest metadata found, retrieving it")
				tmpCatalogManifestFilePath := filepath.Join("/tmp", s.getMetaFilename(manifest.Name))
				manifest.CleanupRequest.AddLocalFileCleanupOperation(tmpCatalogManifestFilePath, false)
				catalogManifest, err = s.readManifestFromFile(tmpCatalogManifestFilePath)
				if err != nil {
					s.ns.NotifyErrorf("Error reading manifest from file %v: %v", tmpCatalogManifestFilePath, err)
					manifest.AddError(err)
					break
				}

				manifest.CreatedAt = catalogManifest.CreatedAt
			}
		}

		// Pushing the necessary files to the remote server
		if catalogManifest != nil {
			manifest.Path = catalogManifest.Path
			manifest.MetadataFile = s.getMetaFilename(catalogManifest.Name)
			manifest.PackFile = s.getPackFilename(catalogManifest.Name)
			if r.MinimumSpecRequirements.Cpu != 0 {
				if manifest.MinimumSpecRequirements == nil {
					manifest.MinimumSpecRequirements = &models.MinimumSpecRequirement{}
				}
				manifest.MinimumSpecRequirements.Cpu = r.MinimumSpecRequirements.Cpu
			}
			if r.MinimumSpecRequirements.Memory != 0 {
				if manifest.MinimumSpecRequirements == nil {
					manifest.MinimumSpecRequirements = &models.MinimumSpecRequirement{}
				}
				manifest.MinimumSpecRequirements.Memory = r.MinimumSpecRequirements.Memory
			}
			if r.MinimumSpecRequirements.Disk != 0 {
				if manifest.MinimumSpecRequirements == nil {
					manifest.MinimumSpecRequirements = &models.MinimumSpecRequirement{}
				}
				manifest.MinimumSpecRequirements.Disk = r.MinimumSpecRequirements.Disk
			}
			localPackPath := filepath.Dir(manifest.CompressedPath)

			// The catalog manifest metadata already exists checking if the files are up to date and pushing them if not
			s.ns.NotifyInfof("Found remote catalog manifest, checking if the files are up to date")
			remotePackChecksum, err := rs.FileChecksum(s.ctx, catalogManifest.Path, catalogManifest.PackFile)
			if err != nil {
				s.ns.NotifyErrorf("Error getting remote pack checksum %v: %v", catalogManifest.PackFile, err)
				manifest.AddError(err)
				break
			}
			if remotePackChecksum != manifest.CompressedChecksum {
				s.ns.NotifyInfof("Remote pack is not up to date, pushing it")
				if err := rs.PushFile(s.ctx, localPackPath, catalogManifest.Path, catalogManifest.PackFile); err != nil {
					s.ns.NotifyErrorf("Error pushing pack file %v: %v", catalogManifest.PackFile, err)
					manifest.AddError(err)
					break
				}
			} else {
				s.ns.NotifyInfof("Remote pack is up to date")
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
			cleanManifest := manifest
			cleanManifest.Provider = nil
			manifestContent, err := json.MarshalIndent(cleanManifest, "", "  ")
			if err != nil {
				s.ns.NotifyErrorf("Error marshalling manifest %v: %v", cleanManifest, err)
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
			s.ns.NotifyInfof("Remote Manifest metadata not found, creating it")

			manifest.Path = filepath.Join(rs.GetProviderRootPath(s.ctx), manifest.CatalogId)
			manifest.MetadataFile = s.getMetaFilename(manifest.Name)
			manifest.PackFile = s.getPackFilename(manifest.Name)
			if r.MinimumSpecRequirements.Cpu != 0 {
				if manifest.MinimumSpecRequirements == nil {
					manifest.MinimumSpecRequirements = &models.MinimumSpecRequirement{}
				}
				manifest.MinimumSpecRequirements.Cpu = r.MinimumSpecRequirements.Cpu
			}
			if r.MinimumSpecRequirements.Memory != 0 {
				if manifest.MinimumSpecRequirements == nil {
					manifest.MinimumSpecRequirements = &models.MinimumSpecRequirement{}
				}
				manifest.MinimumSpecRequirements.Memory = r.MinimumSpecRequirements.Memory
			}
			if r.MinimumSpecRequirements.Disk != 0 {
				if manifest.MinimumSpecRequirements == nil {
					manifest.MinimumSpecRequirements = &models.MinimumSpecRequirement{}
				}
				manifest.MinimumSpecRequirements.Disk = r.MinimumSpecRequirements.Disk
			}
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

			if err := rs.CreateFolder(s.ctx, "/", manifest.Path); err != nil {
				manifest.AddError(err)
				break
			}

			manifest.PackContents = append(manifest.PackContents,
				models.VirtualMachineManifestContentItem{
					Path:      manifest.Path,
					IsDir:     false,
					Name:      filepath.Base(manifest.MetadataFile),
					CreatedAt: helpers.GetUtcCurrentDateTime(),
					UpdatedAt: helpers.GetUtcCurrentDateTime(),
				},
				models.VirtualMachineManifestContentItem{
					Path:      manifest.Path,
					IsDir:     false,
					Name:      filepath.Base(manifest.PackFile),
					Checksum:  manifest.CompressedChecksum,
					CreatedAt: helpers.GetUtcCurrentDateTime(),
					UpdatedAt: helpers.GetUtcCurrentDateTime(),
				})

			cleanManifest := *manifest
			cleanManifest.Provider = nil
			manifestContent, err := json.MarshalIndent(cleanManifest, "", "  ")
			if err != nil {
				s.ns.NotifyErrorf("Error marshalling manifest %v: %v", cleanManifest, err)
				manifest.AddError(err)
				break
			}

			manifest.CleanupRequest.AddLocalFileCleanupOperation(tempManifestContentFilePath, false)
			if err := helper.WriteToFile(string(manifestContent), tempManifestContentFilePath); err != nil {
				s.ns.NotifyErrorf("Error writing manifest to temporary file %v: %v", tempManifestContentFilePath, err)
				manifest.AddError(err)
				break
			}

			s.ns.NotifyInfof("Pushing manifest pack file %v", manifest.PackFile)
			localPackPath := filepath.Dir(manifest.CompressedPath)
			s.sendPushStepInfo(r, "Pushing manifest pack file")
			if err := rs.PushFile(s.ctx, localPackPath, manifest.Path, manifest.PackFile); err != nil {
				manifest.AddError(err)
				break
			}

			s.ns.NotifyInfof("Pushing manifest meta file %v", manifest.MetadataFile)
			if err := rs.PushFile(s.ctx, "/tmp", manifest.Path, manifest.MetadataFile); err != nil {
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
				s.ns.NotifyInfof("Manifest pushed successfully, adding it to the remote database")
				apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
				path := http_helper.JoinUrl(constants.DEFAULT_API_PREFIX, "catalog")

				var response api_models.CatalogManifest
				postUrl := fmt.Sprintf("%s%s", manifest.Provider.GetUrl(), path)
				if _, err := apiClient.Post(postUrl, manifest, &response); err != nil {
					s.ns.NotifyErrorf("Error posting catalog manifest %v: %v", manifest.Provider.String(), err)
					manifest.AddError(err)
					break
				}

				manifest.ID = response.ID
				manifest.Name = response.Name
				manifest.CatalogId = response.CatalogId
			} else {
				s.ns.NotifyInfof("Manifest pushed successfully, adding it to the database")
				db := serviceprovider.Get().JsonDatabase
				if err := db.Connect(s.ctx); err != nil {
					manifest.AddError(err)
					break
				}

				exists, _ := db.GetCatalogManifestsByCatalogIdVersionAndArch(s.ctx, manifest.CatalogId, manifest.Version, manifest.Architecture)
				if exists != nil {
					s.ns.NotifyInfof("Updating manifest %v", manifest.Name)
					dto := mappers.CatalogManifestToDto(*manifest)
					dto.ID = exists.ID
					if _, err := db.UpdateCatalogManifest(s.ctx, dto); err != nil {
						s.ns.NotifyErrorf("Error updating manifest %v: %v", manifest.Name, err)
						manifest.AddError(err)
						break
					}
				} else {
					s.ns.NotifyInfof("Creating manifest %v", manifest.Name)
					dto := mappers.CatalogManifestToDto(*manifest)
					if _, err := db.CreateCatalogManifest(s.ctx, dto); err != nil {
						s.ns.NotifyErrorf("Error creating manifest %v: %v", manifest.Name, err)
						manifest.AddError(err)
						break
					}
				}
			}
		}
	}

	if !executed {
		manifest.AddError(errors.Newf("no remote service found for connection %v", r.Connection))
	}

	if cleanErrors := manifest.CleanupRequest.Clean(s.ctx); len(cleanErrors) > 0 {
		s.ns.NotifyErrorf("Error cleaning up manifest %v", r.CatalogId)
		for _, err := range manifest.Errors {
			manifest.AddError(err)
		}
	}

	return manifest
}
