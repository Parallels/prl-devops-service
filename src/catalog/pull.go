package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/cacheservice"
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	"github.com/Parallels/prl-devops-service/catalog/common"
	"github.com/Parallels/prl-devops-service/catalog/interfaces"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/compressor"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/mappers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/helper/http_helper"
)

func (s *CatalogManifestService) AsyncPull(jobId string, r *models.PullCatalogManifestRequest) {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}

	jobManager := jobs.Get(s.ctx)
	if jobManager == nil {
		s.ns.NotifyErrorf("Job Manager is not available")
		return
	}

	response := s.PullWithExistingJob(jobId, r)
	if response.HasErrors() {
		errorMessage := "Error pulling manifest:"
		for _, err := range response.Errors {
			errorMessage += fmt.Sprintf("\n%v", err)
		}
		jobManager.MarkJobError(jobId, errors.New(errorMessage))
	} else {
		jobManager.MarkJobComplete(jobId, "Virtual Machine Pulled and Registered")
	}
}

// PullWithExistingJob runs pull while streaming provider channel progress into an existing job.
// It does not mark the job as complete/failed; caller owns final job state.
func (s *CatalogManifestService) PullWithExistingJob(jobId string, r *models.PullCatalogManifestRequest) *models.PullCatalogManifestResponse {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}

	if r == nil {
		response := models.NewPullCatalogManifestResponse()
		response.AddError(errors.New("missing pull request"))
		return response
	}

	jobManager := jobs.Get(s.ctx)
	r.JobId = jobId
	if r.ProgressChannel == nil {
		r.ProgressChannel = make(chan int)
	}
	if r.FileNameChannel == nil {
		r.FileNameChannel = make(chan string)
	}
	if r.StepChannel == nil {
		r.StepChannel = make(chan string)
	}

	done := make(chan struct{})
	var wg sync.WaitGroup
	if jobManager != nil {
		wg.Add(1)
		go func(progressChannel chan int, fileNameChannel chan string, stepChannel chan string) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				case progress, ok := <-progressChannel:
					if !ok {
						progressChannel = nil
					} else {
						jobManager.UpdateJobProgress(jobId, "", progress, "running")
					}
				case fileName, ok := <-fileNameChannel:
					if !ok {
						fileNameChannel = nil
					} else {
						jobManager.UpdateJobProgress(jobId, fmt.Sprintf("Downloading %s", fileName), -1, "running")
					}
				case step, ok := <-stepChannel:
					if !ok {
						stepChannel = nil
					} else if step != "" {
						jobManager.UpdateJobProgress(jobId, step, -1, "running")
					}
				}

				if progressChannel == nil && fileNameChannel == nil && stepChannel == nil {
					return
				}
			}
		}(r.ProgressChannel, r.FileNameChannel, r.StepChannel)
	}

	response := s.Pull(r)
	close(done)
	wg.Wait()

	return response
}

func (s *CatalogManifestService) Pull(r *models.PullCatalogManifestRequest) *models.PullCatalogManifestResponse {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}

	foundProvider := false
	response := models.NewPullCatalogManifestResponse()
	response.MachineName = r.MachineName
	apiClient := apiclient.NewHttpClient(s.ctx)
	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService
	db := serviceProvider.JsonDatabase

	disableLocalCatalog := false

	if db == nil {
		disableLocalCatalog = true
	}

	// Not testing the db connection if we have the local catalog disabled
	if !disableLocalCatalog {
		if err := db.Connect(s.ctx); err != nil {
			response.AddError(err)
			return response
		}
	}

	if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
		s.ns.NotifyErrorf("Error creating temp dir: %v", err)
		response.AddError(err)
		return response
	}

	s.ns.NotifyInfof("Checking if the machine %v already exists", r.MachineName)
	jobManager := jobs.Get(s.ctx)
	if r.JobId != "" && jobManager != nil {
		jobManager.UpdateJobProgress(r.JobId, "Checking if the machine already exists", 5, "running")
	}

	filter := fmt.Sprintf("name=%s", r.MachineName)
	vms, err := parallelsDesktopSvc.GetVms(s.ctx, filter)
	if err != nil {
		if errors.GetSystemErrorCode(err) != 404 {
			response.AddError(err)
			return response
		}
	}

	if len(vms) > 0 {
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
		s.ns.NotifyInfof("Checking if the manifest exists in the remote catalog")
		if r.JobId != "" && jobManager != nil {
			jobManager.UpdateJobProgress(r.JobId, "Checking if the manifest exists in the remote catalog", 10, "running")
		}
		manifest = &models.VirtualMachineCatalogManifest{}
		manifest.Provider = &provider
		apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
		srvCtl := system.Get()
		arch, err := srvCtl.GetArchitecture(s.ctx)
		if err != nil {
			response.AddError(errors.New("unable to determine architecture"))
			return response
		}

		var catalogManifest api_models.CatalogManifest
		path := http_helper.JoinUrl(constants.DEFAULT_API_PREFIX, "catalog", helpers.NormalizeStringUpper(r.CatalogId), helpers.NormalizeString(r.Version), arch, "download")
		getUrl := fmt.Sprintf("%s%s", manifest.Provider.GetUrl(), path)
		if clientResponse, err := apiClient.Get(getUrl, &catalogManifest); err != nil {
			if clientResponse != nil && clientResponse.ApiError != nil {
				if clientResponse.StatusCode == 401 || clientResponse.StatusCode == 403 || clientResponse.StatusCode == 400 {
					s.ns.NotifyErrorf("Error getting catalog manifest %v: %v", path, clientResponse.ApiError.Message)
					response.AddError(errors.New(clientResponse.ApiError.Message))
					return response
				}
			}
			if clientResponse.StatusCode == 401 || clientResponse.StatusCode == 403 {
				s.ns.NotifyErrorf("Error getting catalog manifest %v: Unauthorized access", path)
				response.AddError(errors.New("Unauthorized access to the catalog manifest"))
				return response
			}
			if clientResponse.StatusCode == 400 {
				s.ns.NotifyErrorf("Error getting catalog manifest %v: Bad request", path)
				response.AddError(errors.New("Bad request to the catalog manifest"))
				return response
			}
			s.ns.NotifyErrorf("Error getting catalog manifest %v: %v", path, err)
			response.AddError(errors.Newf("Could not find a catalog manifest %s version %s for architecture %s", r.CatalogId, r.Version, arch))
			return response
		}
		m := mappers.ApiCatalogManifestToCatalogManifest(catalogManifest)
		if manifest.Provider != nil {
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
		}

		manifest = &m
		s.ns.NotifyDebugf("Remote Manifest: %v", manifest)
	} else {
		if disableLocalCatalog {
			response.AddError(errors.New("local catalog is disabled"))
			return response
		}
		s.ns.NotifyInfof("Checking if the manifest exists in the local catalog")
		if r.JobId != "" && jobManager != nil {
			jobManager.UpdateJobProgress(r.JobId, "Checking if the manifest exists in the local catalog", 10, "running")
		}
		dto, err := db.GetCatalogManifestByName(s.ctx, r.CatalogId)
		if err != nil {
			manifestErr := errors.Newf("Error getting catalog manifest %v: %v", r.CatalogId, err)
			s.ns.NotifyErrorf("Error getting catalog manifest %v: %v", r.CatalogId, err)
			response.AddError(manifestErr)
			return response
		}
		m := mappers.DtoCatalogManifestToBase(*dto)
		manifest = &m
		s.ns.NotifyDebugf("Local Manifest: %v", manifest)
	}

	// Checking if we have read all of the manifest correctly
	if manifest.CatalogId == "" {
		s.ns.NotifyErrorf("Manifest %v not found in the catalog", r.CatalogId)
		manifestErr := errors.Newf("manifest %v not found in the catalog", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	if manifest.Provider == nil {
		response.AddError(errors.Newf("Manifest %v does not contain a valid provider", r.CatalogId))
		return response
	}

	// Checking for tainted or revoked manifests
	if manifest.Tainted {
		s.ns.NotifyErrorf("Manifest %v is tainted", r.CatalogId)
		manifestErr := errors.Newf("manifest %v is tainted", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	// Check if the manifest is revoked
	if manifest.Revoked {
		s.ns.NotifyErrorf("Manifest %v is revoked", r.CatalogId)
		manifestErr := errors.Newf("manifest %v is revoked", r.CatalogId)
		response.AddError(manifestErr)
		return response
	}

	// Check if the path for the machine exists
	if !helper.FileExists(r.Path) {
		s.ns.NotifyErrorf("Path %v does not exist", r.Path)
		manifestErr := errors.Newf("path %v does not exist", r.Path)
		response.AddError(manifestErr)
		return response
	}

	response.ID = manifest.ID
	response.CatalogId = manifest.CatalogId
	response.Version = manifest.Version

	response.Manifest = manifest
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(s.ctx, manifest.Provider.String())
		if checkErr != nil {
			s.ns.NotifyErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			response.AddError(checkErr)
			break
		}

		if !check {
			continue
		}

		s.ns.NotifyInfof("Found remote service %v", rs.Name())
		rs.SetProgressChannel(r.FileNameChannel, r.ProgressChannel, r.StepChannel)
		foundProvider = true
		r.LocalMachineFolder = fmt.Sprintf("%s.%s", filepath.Join(r.Path, r.MachineName), manifest.Type)

		// Creating the destination folder for the local machine
		if err := s.createDestinationFolder(r, manifest); err != nil {
			response.AddError(err)
			break
		}

		// checking if the manifest is correctly generated
		if manifest.PackFile == "" || manifest.MetadataFile == "" || manifest.Path == "" {
			s.ns.NotifyErrorf("Manifest %v is not correctly generated", manifest.Name)
			response.AddError(errors.Newf("Manifest %v is not correctly generated", manifest.Name))
			break
		}

		// checking if we have the caching enabled, if so we will cache the files using the
		// caching service and then pull the files from the cache
		if cfg.IsCatalogCachingEnable() {
			s.ns.NotifyInfof("Manifest %v caching is enabled, pulling the pack file", manifest.Name)
			if r.JobId != "" && jobManager != nil {
				jobManager.UpdateJobProgress(r.JobId, "Pulling pack file from cache", 15, "running")
			}
			if err := s.pullFromCache(r, manifest, rs); err != nil {
				response.AddError(err)
				break
			}
		} else {
			s.ns.NotifyInfof("Manifest %v caching is disabled, pulling the pack file", manifest.Name)
			if r.JobId != "" && jobManager != nil {
				jobManager.UpdateJobProgress(r.JobId, "Downloading pack file", 15, "running")
			}
			if err := s.pullAndDecompressPackFile(r, manifest, rs); err != nil {
				response.AddError(err)
				break
			}
		}

		systemSrv := serviceProvider.System
		if r.Owner != "" && r.Owner != "root" {
			if err := systemSrv.ChangeFileUserOwner(s.ctx, r.Owner, r.LocalMachineFolder); err != nil {
				s.ns.NotifyErrorf("Error changing file %v owner to %v: %v", r.LocalMachineFolder, r.Owner, err)
				response.AddError(err)
				break
			}
		}

		if response.HasErrors() {
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			break
		}

		s.ns.NotifyInfof("Finished pulling pack file for manifest %v", manifest.Name)
		break
	}

	if !foundProvider {
		response.AddError(errors.New("No remote service was able to pull the manifest"))
	}

	if response.HasErrors() {
		return response
	}

	if r.LocalMachineFolder == "" {
		s.ns.NotifyErrorf("No remote service was able to pull the manifest")
		response.AddError(errors.New("No remote service was able to pull the manifest"))
	}

	// Registering
	s.registerMachineWithParallelsDesktop(r, response)

	// Renaming
	s.renameMachineWithParallelsDesktop(r, response)

	// starting the machine
	if r.StartAfterPull {
		s.startMachineWithParallelsDesktop(r, response)
	}

	// Cleaning up
	s.CleanPullRequest(r, response)

	return response
}

func (s *CatalogManifestService) registerMachineWithParallelsDesktop(r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	s.ns.NotifyInfof("Registering machine %v", r.MachineName)
	jobManager := jobs.Get(s.ctx)
	if r.JobId != "" && jobManager != nil {
		jobManager.UpdateJobProgress(r.JobId, "Registering the machine with Parallels Desktop", 90, "running")
	}

	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		machineRegisterRequest := api_models.RegisterVirtualMachineRequest{
			Path:                 r.LocalMachineFolder,
			Owner:                r.Owner,
			MachineName:          r.MachineName,
			RegenerateSourceUuid: true,
		}

		if err := parallelsDesktopSvc.RegisterVm(s.ctx, machineRegisterRequest); err != nil {
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
		}
	} else {
		s.ns.NotifyErrorf("Error registering machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) renameMachineWithParallelsDesktop(r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	s.ns.NotifyInfof("Renaming machine %v", r.MachineName)
	jobManager := jobs.Get(s.ctx)
	if r.JobId != "" && jobManager != nil {
		jobManager.UpdateJobProgress(r.JobId, "Renaming machine", 95, "running")
	}

	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		s.ns.NotifyInfof("Renaming machine %v to %v", r.MachineName, r.MachineName)
		filter := fmt.Sprintf("name=%s", r.MachineName)
		vms, err := parallelsDesktopSvc.GetVms(s.ctx, filter)
		if err != nil {
			s.ns.NotifyErrorf("Error getting machine %v: %v", r.MachineName, err)
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
			s.ns.NotifyErrorf("Error getting machine %v: %v", r.MachineName, notFoundError)
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

			if err := parallelsDesktopSvc.RenameVm(s.ctx, renameRequest); err != nil {
				s.ns.NotifyErrorf("Error renaming machine %v: %v", r.MachineName, err)
				response.AddError(err)
				response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
				return
			}
		}

		response.MachineID = vms[0].ID
	} else {
		s.ns.NotifyErrorf("Error renaming machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) startMachineWithParallelsDesktop(r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	s.ns.NotifyInfof("Starting machine %v for %v", r.MachineName, r.CatalogId)
	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		filter := fmt.Sprintf("name=%s", r.MachineName)
		vms, err := parallelsDesktopSvc.GetVms(s.ctx, filter)
		if err != nil {
			s.ns.NotifyErrorf("Error getting machine %v: %v", r.MachineName, err)
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
			s.ns.NotifyErrorf("Error getting machine %v: %v", r.MachineName, notFoundError)
			response.AddError(notFoundError)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		if err := parallelsDesktopSvc.StartVm(s.ctx, vm.ID); err != nil {
			s.ns.NotifyErrorf("Error starting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}
	} else {
		s.ns.NotifyErrorf("Error starting machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) CleanPullRequest(r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	if cleanErrors := response.CleanupRequest.Clean(s.ctx); len(cleanErrors) > 0 {
		s.ns.NotifyErrorf("Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	}
}

func (s *CatalogManifestService) createDestinationFolder(r *models.PullCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest) error {
	jobManager := jobs.Get(s.ctx)
	if r.JobId != "" && jobManager != nil {
		jobManager.UpdateJobProgress(r.JobId, "Creating destination folder", 12, "running")
	}

	r.LocalMachineFolder = fmt.Sprintf("%s.%s", filepath.Join(r.Path, r.MachineName), manifest.Type)
	s.ns.NotifyInfof("Local machine folder: %v", r.LocalMachineFolder)
	count := 1
	max_attempts := 30
	created := false
	for {
		if helper.FileExists(r.LocalMachineFolder) {
			s.ns.NotifyInfof("Local machine folder %v already exists, attempting to create a different one", r.LocalMachineFolder)
			r.LocalMachineFolder = fmt.Sprintf("%s_%v.%s", filepath.Join(r.Path, r.MachineName), count, manifest.Type)
			count += 1
			if count > max_attempts {
				s.ns.NotifyInfof("Max attempts reached to find a new local machine folder name, breaking")
				break
			}
		} else {
			created = true
			break
		}
	}
	if !created {
		s.ns.NotifyErrorf("Error creating local machine folder %v", r.LocalMachineFolder)
		return errors.Newf("Error creating local machine folder %v", r.LocalMachineFolder)
	}

	if err := helpers.CreateDirIfNotExist(r.LocalMachineFolder); err != nil {
		s.ns.NotifyErrorf("Error creating local machine folder %v: %v", r.LocalMachineFolder, err)
		return err
	}

	s.ns.NotifyInfof("Created local machine folder %v", r.LocalMachineFolder)
	return nil
}

func (s *CatalogManifestService) pullFromCache(r *models.PullCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest, rss interfaces.RemoteStorageService) error {
	cacheService, err := cacheservice.NewCacheService(s.ctx)
	if err != nil {
		s.ns.NotifyErrorf("Error creating cache service: %v", err)
		return err
	}

	cacheRequest := cacheservice.NewCacheRequest(s.ctx, manifest, rss)
	cacheService.WithRequest(cacheRequest)

	if !cacheService.IsCached() {
		s.ns.NotifyInfof("Manifest %v is not cached, caching it", manifest.Name)
		if err := cacheService.Cache(); err != nil {
			s.ns.NotifyErrorf("Error caching manifest %v: %v", manifest.Name, err)
			return err
		}
	}

	// We now need to copy the cached folder to the local machine folder
	cacheResponse, err := cacheService.Get()
	if err != nil {
		s.ns.NotifyErrorf("Error getting cache response: %v", err)
		return err
	}
	if err := helpers.CopyDir(cacheResponse.PackFilePath, r.LocalMachineFolder); err != nil {
		s.ns.NotifyErrorf("Error copying cached folder %v to %v: %v", cacheResponse.PackFilePath, r.LocalMachineFolder, err)
		return err
	}
	s.ctx.LogInfof("Finished copying cached folder %v to %v", cacheResponse.PackFilePath, r.LocalMachineFolder)

	return nil
}

func (s *CatalogManifestService) pullAndDecompressPackFile(r *models.PullCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest, rss interfaces.RemoteStorageService) error {
	if rss == nil {
		return errors.NewWithCode("Remote storage service is nil", 500)
	}
	s.ctx.LogInfof("Pulling and decompressing pack file for manifest ID %v, Name %v", manifest.ID, manifest.Name)
	cfg := config.Get()
	cleanupSvc := cleanupservice.NewCleanupService()
	if rss.CanStream() && cfg.IsRemoteProviderStreamEnabled() {
		if err := s.processFileWithStream(r, rss, manifest, cleanupSvc); err != nil {
			return err
		}
	} else {
		if err := s.processFileWithoutStream(r, rss, manifest, cleanupSvc); err != nil {
			return err
		}
	}

	jobManager := jobs.Get(s.ctx)
	if r.JobId != "" && jobManager != nil {
		jobManager.UpdateJobProgress(r.JobId, "Cleaning and flattening pulled structure", 85, "running")
	}

	if err := common.CleanAndFlatten(r.LocalMachineFolder); err != nil {
		s.ctx.LogErrorf("Error cleaning and flattening local machine folder %v: %v", r.LocalMachineFolder, err)
		cleanupSvc.Clean(s.ctx)
		return err
	}
	s.ctx.LogInfof("Operation completed successfully for manifest ID %v, Name %v, cleaning up", manifest.ID, manifest.Name)
	cleanupSvc.Clean(s.ctx)
	return nil
}

func (s *CatalogManifestService) processFileWithStream(r *models.PullCatalogManifestRequest, rss interfaces.RemoteStorageService, manifest *models.VirtualMachineCatalogManifest, cleanupSvc *cleanupservice.CleanupService) error {
	destinationFolder := r.LocalMachineFolder
	if err := rss.PullFileAndDecompress(s.ctx, manifest.Path, manifest.PackFile, destinationFolder); err != nil {
		s.ctx.LogErrorf("Error pulling and decompressing pack file for manifest ID %v, Name %v: %v adding folder: %v to cleanup", manifest.ID, manifest.Name, err, destinationFolder)
		cleanupSvc.AddLocalFileCleanupOperation(destinationFolder, true)
		return err
	}
	return nil
}

func (s *CatalogManifestService) processFileWithoutStream(r *models.PullCatalogManifestRequest, rss interfaces.RemoteStorageService, manifest *models.VirtualMachineCatalogManifest, cleanupSvc *cleanupservice.CleanupService) error {
	destinationFolder := r.LocalMachineFolder
	// Creating the path for temporary file
	tempDir := os.TempDir()
	tempFilename := manifest.CompressedChecksum
	if tempFilename == "" {
		tempFilename = helpers.GenerateId()
	}

	tempDestinationFolder := filepath.Join(tempDir, tempFilename)
	if err := os.MkdirAll(tempDestinationFolder, os.ModePerm); err != nil {
		return err
	}

	// Adding the cleanup operation for the temporary folder
	cleanupSvc.AddLocalFileCleanupOperation(tempDestinationFolder, true)
	s.ctx.LogInfof("Added temporary folder %v to cleanup operations", tempDestinationFolder)

	// Pulling the file to the temporary folder
	if err := rss.PullFile(s.ctx, manifest.Path, manifest.PackFile, tempDestinationFolder); err != nil {
		s.ctx.LogErrorf("Error pulling file for manifest ID %v, Name %v: %v adding folder: %v to cleanup", manifest.ID, manifest.Name, err, tempDestinationFolder)
		cleanupSvc.Clean(s.ctx)
		return err
	}

	// checking if the pack file is compressed or not if it is we will decompress it to the destination folder
	// and remove the pack file from the cache folder if not we will just rename the pack file to the checksum
	if manifest.IsCompressed || strings.HasSuffix(manifest.PackFile, ".pdpack") {
		compressedFilePath := filepath.Join(tempDestinationFolder, manifest.PackFile)

		jobManager := jobs.Get(s.ctx)
		if r.JobId != "" && jobManager != nil {
			jobManager.UpdateJobProgress(r.JobId, "Decompressing Pack File", 80, "running")
		}

		if err := compressor.DecompressFileWithStepChannel(s.ctx, compressedFilePath, destinationFolder, r.StepChannel); err != nil {
			cleanupSvc.AddLocalFileCleanupOperation(destinationFolder, true)
			s.ctx.LogErrorf("Error decompressing file for manifest ID %v, Name %v: %v adding folder: %v to cleanup", manifest.ID, manifest.Name, err, destinationFolder)
			cleanupSvc.Clean(s.ctx)
			return err
		}
	} else {
		tempFilePath := filepath.Join(tempDestinationFolder, manifest.PackFile)
		if info, err := os.Stat(tempFilePath); err == nil && info.IsDir() {
			if err := helpers.CopyDir(tempFilePath, destinationFolder); err != nil {
				cleanupSvc.Clean(s.ctx)
				s.ctx.LogErrorf("Error copying directory for manifest ID %v, Name %v: %v adding folder: %v to cleanup", manifest.ID, manifest.Name, err, destinationFolder)
				return err
			}
		} else {
			if err := helpers.CopyFile(tempFilePath, destinationFolder); err != nil {
				s.ctx.LogErrorf("Error copying file for manifest ID %v, Name %v: %v adding folder: %v to cleanup", manifest.ID, manifest.Name, err, destinationFolder)
				cleanupSvc.Clean(s.ctx)
				return err
			}
		}
	}
	s.ctx.LogInfof("Finished pulling and decompressing pack file for manifest ID %v, Name %v", manifest.ID, manifest.Name)
	return nil
}
