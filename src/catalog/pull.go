package catalog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/helper/http_helper"
)

type CatalogCacheType int

const (
	CatalogCacheTypeNone CatalogCacheType = iota
	CatalogCacheTypeFile
	CatalogCacheTypeFolder
)

func (s *CatalogManifestService) Pull(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest) *models.PullCatalogManifestResponse {
	foundProvider := false
	response := models.NewPullCatalogManifestResponse()
	response.MachineName = r.MachineName
	apiClient := apiclient.NewHttpClient(ctx)
	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService
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

	ctx.LogInfof("Checking if the machine %v already exists", r.MachineName)
	s.sendPullStepInfo(r, "Checking if the machine already exists")
	exists, err := parallelsDesktopSvc.GetVmSync(ctx, r.MachineName)
	if err != nil {
		if errors.GetSystemErrorCode(err) != 404 {
			response.AddError(err)
			return response
		}
	}

	if exists != nil {
		response.AddError(errors.Newf("machine %v already exists", r.MachineName))
		return response
	}

	var manifest *models.VirtualMachineCatalogManifest
	provider := models.CatalogManifestProvider{}
	cfg := config.Get()

	if err := provider.Parse(r.Connection); err != nil {
		response.AddError(err)
		return response
	}

	// getting the provider metadata from the database
	if provider.IsRemote() {
		ctx.LogInfof("Checking if the manifest exists in the remote catalog")
		manifest = &models.VirtualMachineCatalogManifest{}
		manifest.Provider = &provider
		apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
		srvCtl := system.Get()
		arch, err := srvCtl.GetArchitecture(ctx)
		if err != nil {
			response.AddError(errors.New("unable to determine architecture"))
			return response
		}

		var catalogManifest api_models.CatalogManifest
		path := http_helper.JoinUrl(constants.DEFAULT_API_PREFIX, "catalog", helpers.NormalizeStringUpper(r.CatalogId), helpers.NormalizeString(r.Version), arch, "download")
		getUrl := fmt.Sprintf("%s%s", manifest.Provider.GetUrl(), path)
		if clientResponse, err := apiClient.Get(getUrl, &catalogManifest); err != nil {
			if clientResponse != nil && clientResponse.ApiError != nil {
				if clientResponse.StatusCode == 403 || clientResponse.StatusCode == 400 {
					ctx.LogErrorf("Error getting catalog manifest %v: %v", path, clientResponse.ApiError.Message)
					response.AddError(errors.Newf(clientResponse.ApiError.Message))
					return response
				}
			}
			ctx.LogErrorf("Error getting catalog manifest %v: %v", path, err)
			response.AddError(errors.Newf("Could not find a catalog manifest %s version %s for architecture %s", r.CatalogId, r.Version, arch))
			return response
		}
		m := mappers.ApiCatalogManifestToCatalogManifest(catalogManifest)
		if manifest.Provider.Host != "" {
			m.Provider.Host = manifest.Provider.Host
		}
		if manifest.Provider.Port != "" {
			m.Provider.Port = manifest.Provider.Port
		}
		if manifest.Provider.Username != "" {
			m.Provider.Username = manifest.Provider.Username
		}
		if manifest.Provider.Password != "" {
			m.Provider.Password = manifest.Provider.Password
		}
		if manifest.Provider.ApiKey != "" {
			m.Provider.ApiKey = manifest.Provider.ApiKey
		}
		if len(manifest.Provider.Meta) > 0 {
			for key, value := range manifest.Provider.Meta {
				m.Provider.Meta[key] = value
			}
		}

		manifest = &m
		ctx.LogDebugf("Remote Manifest: %v", manifest)
	} else {
		s.sendPullStepInfo(r, "Checking if the manifest exists in the local catalog")
		ctx.LogInfof("Checking if the manifest exists in the local catalog")
		dto, err := db.GetCatalogManifestByName(ctx, r.CatalogId)
		if err != nil {
			manifestErr := errors.Newf("Error getting catalog manifest %v: %v", r.CatalogId, err)
			ctx.LogErrorf(manifestErr.Error())
			response.AddError(manifestErr)
			return response
		}
		m := mappers.DtoCatalogManifestToBase(*dto)
		manifest = &m
		ctx.LogDebugf("Local Manifest: %v", manifest)
	}

	// Checking if we have read all of the manifest correctly
	if manifest == nil || manifest.Provider == nil {
		ctx.LogErrorf("Manifest %v not found in the catalog", r.CatalogId)
		manifestErr := errors.Newf("manifest %v not found in the catalog", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	// Checking for tainted or revoked manifests
	if manifest.Tainted {
		ctx.LogErrorf("Manifest %v is tainted", r.CatalogId)
		manifestErr := errors.Newf("manifest %v is tainted", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	if manifest.Revoked {
		ctx.LogErrorf("Manifest %v is revoked", r.CatalogId)
		manifestErr := errors.Newf("manifest %v is revoked", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	if !helper.FileExists(r.Path) {
		ctx.LogErrorf("Path %v does not exist", r.Path)
		manifestErr := errors.Newf("path %v does not exist", r.Path)
		response.AddError(manifestErr)
		return response
	}

	response.ID = manifest.ID
	response.CatalogId = manifest.CatalogId
	response.Version = manifest.Version

	response.Manifest = manifest
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(ctx, manifest.Provider.String())
		if checkErr != nil {
			ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			response.AddError(checkErr)
			break
		}

		if check {
			ctx.LogInfof("Found remote service %v", rs.Name())
			rs.SetProgressChannel(r.FileNameChannel, r.ProgressChannel)
			foundProvider = true
			r.LocalMachineFolder = fmt.Sprintf("%s.%s", filepath.Join(r.Path, r.MachineName), manifest.Type)
			ctx.LogInfof("Local machine folder: %v", r.LocalMachineFolder)
			count := 1
			for {
				if helper.FileExists(r.LocalMachineFolder) {
					ctx.LogInfof("Local machine folder %v already exists, attempting to create a different one", r.LocalMachineFolder)
					r.LocalMachineFolder = fmt.Sprintf("%s_%v.%s", filepath.Join(r.Path, r.MachineName), count, manifest.Type)
					count += 1
				} else {
					break
				}
			}

			if err := helpers.CreateDirIfNotExist(r.LocalMachineFolder); err != nil {
				ctx.LogErrorf("Error creating local machine folder %v: %v", r.LocalMachineFolder, err)
				response.AddError(err)
				break
			}

			ctx.LogInfof("Created local machine folder %v", r.LocalMachineFolder)

			ctx.LogInfof("Pulling manifest %v", manifest.Name)
			packContent := make([]models.VirtualMachineManifestContentItem, 0)
			if manifest.PackContents == nil {
				ctx.LogDebugf("Manifest %v does not have pack contents, adding default files", manifest.Name)
				packContent = append(packContent, models.VirtualMachineManifestContentItem{
					Path: manifest.Path,
					Name: manifest.PackFile,
				})
				packContent = append(packContent, models.VirtualMachineManifestContentItem{
					Path: manifest.Path,
					Name: manifest.MetadataFile,
				})
				ctx.LogDebugf("default file content %v", packContent)
			} else {
				ctx.LogDebugf("Manifest %v has pack contents, adding them", manifest.Name)
				packContent = append(packContent, manifest.PackContents...)
			}
			ctx.LogDebugf("pack content %v", packContent)

			for _, file := range packContent {
				if strings.HasSuffix(file.Name, ".meta") {
					ctx.LogDebugf("Skipping meta file %v", file.Name)
					continue
				}

				destinationFolder := r.Path
				fileName := file.Name
				fileChecksum, err := rs.FileChecksum(ctx, file.Path, file.Name)
				if err != nil {
					ctx.LogErrorf("Error getting file %v checksum: %v", fileName, err)
					response.AddError(err)
					break
				}

				cacheFileName := fmt.Sprintf("%s.pdpack", fileChecksum)
				cacheMachineName := fmt.Sprintf("%s.%s", fileChecksum, manifest.Type)
				cacheType := CatalogCacheTypeNone
				needsPulling := false
				// checking for the caching system to see if we need to pull the file
				if cfg.IsCatalogCachingEnable() {
					destinationFolder, err = cfg.CatalogCacheFolder()
					if err != nil {
						destinationFolder = r.Path
					}

					// checking if the compressed file is already in the cache
					if helper.FileExists(filepath.Join(destinationFolder, cacheFileName)) {
						ctx.LogInfof("Compressed File %v already exists in cache", fileName)
						cacheType = CatalogCacheTypeFile
					} else if helper.FileExists(filepath.Join(destinationFolder, cacheMachineName)) {
						ctx.LogInfof("Machine Folder %v already exists in cache", cacheMachineName)
						cacheType = CatalogCacheTypeFolder
					} else {
						cacheType = CatalogCacheTypeFile
						needsPulling = true
					}
				} else {
					cacheType = CatalogCacheTypeFile
					needsPulling = true
				}

				if needsPulling {
					s.sendPullStepInfo(r, "Pulling file")
					if err := rs.PullFile(ctx, file.Path, file.Name, destinationFolder); err != nil {
						ctx.LogErrorf("Error pulling file %v: %v", fileName, err)
						response.AddError(err)
						break
					}
					if cfg.IsCatalogCachingEnable() {
						err := os.Rename(filepath.Join(destinationFolder, file.Name), filepath.Join(destinationFolder, cacheFileName))
						if err != nil {
							log.Fatal(err)
						}
					}
				}

				if !cfg.IsCatalogCachingEnable() {
					cacheFileName = file.Name
					response.CleanupRequest.AddLocalFileCleanupOperation(filepath.Join(destinationFolder, file.Name), false)
				}

				if cacheType == CatalogCacheTypeFile {
					s.sendPullStepInfo(r, "Decompressing file")
					ctx.LogInfof("Decompressing file %v", cacheFileName)
					if err := s.decompressMachine(ctx, filepath.Join(destinationFolder, cacheFileName), filepath.Join(destinationFolder, cacheMachineName)); err != nil {
						ctx.LogErrorf("Error decompressing file %v: %v", fileName, err)
						response.AddError(err)
						break
					}

					if err := helper.DeleteFile(filepath.Join(destinationFolder, cacheFileName)); err != nil {
						ctx.LogErrorf("Error deleting file %v: %v", cacheFileName, err)
						response.AddError(err)
						break
					}

					cacheType = CatalogCacheTypeFolder
				}

				if cacheType == CatalogCacheTypeFolder {
					s.sendPullStepInfo(r, "Copying machine to destination")
					ctx.LogInfof("Copying machine folder %v to %v", cacheMachineName, r.LocalMachineFolder)
					if err := helpers.CopyDir(filepath.Join(destinationFolder, cacheMachineName), r.LocalMachineFolder); err != nil {
						ctx.LogErrorf("Error copying machine folder %v to %v: %v", cacheMachineName, r.LocalMachineFolder, err)
						response.AddError(err)
						break
					}
				}

				systemSrv := serviceProvider.System
				if r.Owner != "" && r.Owner != "root" {
					if err := systemSrv.ChangeFileUserOwner(ctx, r.Owner, r.LocalMachineFolder); err != nil {
						ctx.LogErrorf("Error changing file %v owner to %v: %v", r.LocalMachineFolder, r.Owner, err)
						response.AddError(err)
						break
					}
				}
			}

			if response.HasErrors() {
				response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			}

			ctx.LogInfof("Finished pulling pack file for manifest %v", manifest.Name)
		}
	}

	if !foundProvider {
		response.AddError(errors.New("No remote service was able to pull the manifest"))
	}

	if r.LocalMachineFolder == "" {
		ctx.LogErrorf("No remote service was able to pull the manifest")
		response.AddError(errors.New("No remote service was able to pull the manifest"))
	}

	// Registering
	s.registerMachineWithParallelsDesktop(ctx, r, response)

	// Renaming
	s.renameMachineWithParallelsDesktop(ctx, r, response)

	// starting the machine
	if r.StartAfterPull {
		s.startMachineWithParallelsDesktop(ctx, r, response)
	}

	// Cleaning up
	s.CleanPullRequest(ctx, r, response)

	return response
}

func (s *CatalogManifestService) registerMachineWithParallelsDesktop(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	ctx.LogInfof("Registering machine %v", r.MachineName)
	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		machineRegisterRequest := api_models.RegisterVirtualMachineRequest{
			Path:                 r.LocalMachineFolder,
			Owner:                r.Owner,
			MachineName:          r.MachineName,
			RegenerateSourceUuid: true,
		}

		if err := parallelsDesktopSvc.RegisterVm(ctx, machineRegisterRequest); err != nil {
			ctx.LogErrorf("Error registering machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
		}
	} else {
		ctx.LogErrorf("Error registering machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) renameMachineWithParallelsDesktop(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	ctx.LogInfof("Renaming machine %v", r.MachineName)
	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		ctx.LogInfof("Renaming machine %v to %v", r.MachineName, r.MachineName)
		filter := fmt.Sprintf("name=%s", r.MachineName)
		vms, err := parallelsDesktopSvc.GetVmsSync(ctx, filter)
		if err != nil {
			ctx.LogErrorf("Error getting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		var vm *api_models.ParallelsVM
		if len(vms) > 1 {
			for _, searchVM := range vms {
				if searchVM.Name == r.MachineName {
					vm = &searchVM
					break
				}
			}
		} else if len(vms) == 1 {
			vm = &vms[0]
		}

		if vm == nil {
			notFoundError := errors.Newf("Machine %v not found", r.MachineName)
			ctx.LogErrorf("Error getting machine %v: %v", r.MachineName, notFoundError)
			response.AddError(notFoundError)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		// Renaming only if the name is different
		if vm.Name != r.MachineName {
			response.ID = vm.ID
			renameRequest := api_models.RenameVirtualMachineRequest{
				ID:          vm.ID,
				CurrentName: vm.Name,
				NewName:     r.MachineName,
			}

			if err := parallelsDesktopSvc.RenameVm(ctx, renameRequest); err != nil {
				ctx.LogErrorf("Error renaming machine %v: %v", r.MachineName, err)
				response.AddError(err)
				response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
				return
			}
		}

		response.MachineID = vms[0].ID
	} else {
		ctx.LogErrorf("Error renaming machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) startMachineWithParallelsDesktop(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	ctx.LogInfof("Starting machine %v for %v", r.MachineName, r.CatalogId)
	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		filter := fmt.Sprintf("name=%s", r.MachineName)
		vms, err := parallelsDesktopSvc.GetVmsSync(ctx, filter)
		if err != nil {
			ctx.LogErrorf("Error getting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		var vm *api_models.ParallelsVM
		if len(vms) > 1 {
			for _, searchVM := range vms {
				if searchVM.Name == r.MachineName {
					vm = &searchVM
					break
				}
			}
		} else if len(vms) == 1 {
			vm = &vms[0]
		}

		if vm == nil {
			notFoundError := errors.Newf("Machine %v not found", r.MachineName)
			ctx.LogErrorf("Error getting machine %v: %v", r.MachineName, notFoundError)
			response.AddError(notFoundError)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		if err := parallelsDesktopSvc.StartVm(ctx, vm.ID); err != nil {
			ctx.LogErrorf("Error starting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}
	} else {
		ctx.LogErrorf("Error starting machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) CleanPullRequest(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	if cleanErrors := response.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		ctx.LogErrorf("Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	}
}
