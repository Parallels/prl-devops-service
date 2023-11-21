package catalog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	api_models "github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/Parallels/pd-api-service/serviceprovider/apiclient"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/helper/http_helper"
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
		ctx.LogError("Error creating temp dir: %v", err)
		response.AddError(err)
		return response
	}

	ctx.LogInfo("Checking if the machine %v already exists", r.MachineName)
	exists, err := parallelsDesktopSvc.GetVm(ctx, r.MachineName)
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
	cfg := config.NewConfig()

	if err := provider.Parse(r.Connection); err != nil {
		response.AddError(err)
		return response
	}

	// getting the provider metadata from the database
	if provider.IsRemote() {
		ctx.LogInfo("Checking if the manifest exists in the remote catalog")
		manifest = &models.VirtualMachineCatalogManifest{}
		manifest.Provider = &provider
		apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))

		var catalogManifest api_models.CatalogManifest
		path := http_helper.JoinUrl(constants.DEFAULT_API_PREFIX, "catalog", helpers.NormalizeStringUpper(r.CatalogId), helpers.NormalizeString(r.Version), "download")
		getUrl := fmt.Sprintf("%s%s", manifest.Provider.GetUrl(), path)
		if clientResponse, err := apiClient.Get(getUrl, &catalogManifest); err != nil {
			if clientResponse != nil && clientResponse.ApiError != nil {
				if clientResponse.StatusCode == 403 || clientResponse.StatusCode == 400 {
					ctx.LogError("Error getting catalog manifest %v: %v", path, clientResponse.ApiError.Message)
					response.AddError(errors.Newf(clientResponse.ApiError.Message))
					return response
				}
			}
			ctx.LogError("Error getting catalog manifest %v: %v", path, err)
			response.AddError(err)
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
		ctx.LogDebug("Remote Manifest: %v", manifest)
	} else {
		ctx.LogInfo("Checking if the manifest exists in the local catalog")
		dto, err := db.GetCatalogManifestByName(ctx, r.CatalogId)
		if err != nil {
			manifestErr := errors.Newf("Error getting catalog manifest %v: %v", r.CatalogId, err)
			ctx.LogError(manifestErr.Error())
			response.AddError(manifestErr)
			return response
		}
		m := mappers.DtoCatalogManifestToBase(*dto)
		manifest = &m
		ctx.LogDebug("Local Manifest: %v", manifest)
	}

	// Checking if we have read all of the manifest correctly
	if manifest == nil || manifest.Provider == nil {
		ctx.LogError("Manifest %v not found in the catalog", r.CatalogId)
		manifestErr := errors.Newf("manifest %v not found in the catalog", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	// Checking for tainted or revoked manifests
	if manifest.Tainted {
		ctx.LogError("Manifest %v is tainted", r.CatalogId)
		manifestErr := errors.Newf("manifest %v is tainted", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	if manifest.Revoked {
		ctx.LogError("Manifest %v is revoked", r.CatalogId)
		manifestErr := errors.Newf("manifest %v is revoked", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	if !helper.FileExists(r.Path) {
		ctx.LogError("Path %v does not exist", r.Path)
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
			ctx.LogError("Error checking remote service %v: %v", rs.Name(), checkErr)
			response.AddError(checkErr)
			break
		}

		if check {
			ctx.LogInfo("Found remote service %v", rs.Name())
			foundProvider = true
			r.LocalMachineFolder = fmt.Sprintf("%s.%s", filepath.Join(r.Path, r.MachineName), manifest.Type)
			ctx.LogInfo("Local machine folder: %v", r.LocalMachineFolder)
			count := 1
			for {
				if helper.FileExists(r.LocalMachineFolder) {
					ctx.LogInfo("Local machine folder %v already exists, attempting to create a different one", r.LocalMachineFolder)
					r.LocalMachineFolder = fmt.Sprintf("%s_%v.%s", filepath.Join(r.Path, r.MachineName), count, manifest.Type)
					count += 1
				} else {
					break
				}
			}

			if err := helpers.CreateDirIfNotExist(r.LocalMachineFolder); err != nil {
				ctx.LogError("Error creating local machine folder %v: %v", r.LocalMachineFolder, err)
				response.AddError(err)
				break
			}
			ctx.LogInfo("Created local machine folder %v", r.LocalMachineFolder)

			ctx.LogInfo("Pulling manifest %v", manifest.Name)
			packContent := make([]models.VirtualMachineManifestContentItem, 0)
			if manifest.PackContents == nil {
				ctx.LogDebug("Manifest %v does not have pack contents, adding default files", manifest.Name)
				packContent = append(packContent, models.VirtualMachineManifestContentItem{
					Path: manifest.Path,
					Name: manifest.PackFile,
				})
				packContent = append(packContent, models.VirtualMachineManifestContentItem{
					Path: manifest.Path,
					Name: manifest.MetadataFile,
				})
				ctx.LogDebug("default file content %v", packContent)
			} else {
				ctx.LogDebug("Manifest %v has pack contents, adding them", manifest.Name)
				packContent = append(packContent, manifest.PackContents...)
			}
			ctx.LogDebug("pack content %v", packContent)

			for _, file := range packContent {
				if strings.HasSuffix(file.Name, ".meta") {
					ctx.LogDebug("Skipping meta file %v", file.Name)
					continue
				}

				destinationFolder := r.Path
				fileName := file.Name
				fileChecksum, err := rs.FileChecksum(ctx, file.Path, file.Name)
				if err != nil {
					ctx.LogError("Error getting file %v checksum: %v", fileName, err)
					response.AddError(err)
					break
				}

				cacheFileName := fmt.Sprintf("%s.pdpack", fileChecksum)
				needsPulling := false
				if cfg.IsCatalogCachingEnable() {
					destinationFolder, err = cfg.GetCatalogCacheFolder()
					if err != nil {
						destinationFolder = r.Path
					}
					if helper.FileExists(filepath.Join(destinationFolder, cacheFileName)) {
						ctx.LogInfo("File %v already exists in cache", fileName)
					} else {
						needsPulling = true
					}
				} else {
					needsPulling = true
				}

				if needsPulling {
					if err := rs.PullFile(ctx, file.Path, file.Name, destinationFolder); err != nil {
						ctx.LogError("Error pulling file %v: %v", fileName, err)
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

				ctx.LogInfo("Decompressing file %v", cacheFileName)
				if err := s.decompressMachine(ctx, filepath.Join(destinationFolder, cacheFileName), r.LocalMachineFolder); err != nil {
					ctx.LogError("Error decompressing file %v: %v", fileName, err)
					response.AddError(err)
					break
				}

				systemSrv := serviceProvider.System
				if r.Owner != "" && r.Owner != "root" {
					if err := systemSrv.ChangeFileUserOwner(ctx, r.Owner, r.LocalMachineFolder); err != nil {
						ctx.LogError("Error changing file %v owner to %v: %v", r.LocalMachineFolder, r.Owner, err)
						response.AddError(err)
						break
					}
				}
			}

			if response.HasErrors() {
				response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			}

			ctx.LogInfo("Finished pulling pack file for manifest %v", manifest.Name)
		}
	}

	if !foundProvider {
		response.AddError(errors.New("No remote service was able to pull the manifest"))
	}

	if r.LocalMachineFolder == "" {
		ctx.LogError("No remote service was able to pull the manifest")
		response.AddError(errors.New("No remote service was able to pull the manifest"))
	}

	// Registering
	s.registerMachineWithParallelsDesktop(ctx, r, response)

	// Renaming
	s.renameMachineWithParallelsDesktop(ctx, r, response)

	//starting the machine
	if r.StartAfterPull {
		s.startMachineWithParallelsDesktop(ctx, r, response)
	}

	// Cleaning up
	s.CleanPullRequest(ctx, r, response)

	return response
}

func (s *CatalogManifestService) registerMachineWithParallelsDesktop(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	ctx.LogInfo("Registering machine %v", r.MachineName)
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
			ctx.LogError("Error registering machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
		}
	} else {
		ctx.LogError("Error registering machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) renameMachineWithParallelsDesktop(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	ctx.LogInfo("Renaming machine %v", r.MachineName)
	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		ctx.LogInfo("Renaming machine %v to %v", r.MachineName, r.MachineName)
		filter := fmt.Sprintf("home=%s", r.LocalMachineFolder)
		vms, err := parallelsDesktopSvc.GetVms(ctx, filter)
		if err != nil {
			ctx.LogError("Error getting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		if len(vms) != 1 {
			ctx.LogError("Error getting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		response.ID = vms[0].ID
		renameRequest := api_models.RenameVirtualMachineRequest{
			ID:          vms[0].ID,
			CurrentName: vms[0].Name,
			NewName:     r.MachineName,
		}

		if err := parallelsDesktopSvc.RenameVm(ctx, renameRequest); err != nil {
			ctx.LogError("Error renaming machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}
	} else {
		ctx.LogError("Error renaming machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) startMachineWithParallelsDesktop(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	ctx.LogInfo("Starting machine %v", r.MachineName)
	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		ctx.LogInfo("Starting machine %v to %v", r.MachineName, r.MachineName)
		filter := fmt.Sprintf("home=%s", r.LocalMachineFolder)
		vms, err := parallelsDesktopSvc.GetVms(ctx, filter)
		if err != nil {
			ctx.LogError("Error getting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		if len(vms) != 1 {
			ctx.LogError("Error getting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		if err := parallelsDesktopSvc.StartVm(ctx, vms[0].ID); err != nil {
			ctx.LogError("Error starting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}
	} else {
		ctx.LogError("Error starting machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) CleanPullRequest(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	if cleanErrors := response.CleanupRequest.Clean(ctx); len(cleanErrors) > 0 {
		ctx.LogError("Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	}
}
