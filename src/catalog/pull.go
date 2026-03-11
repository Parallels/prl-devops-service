package catalog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	"github.com/Parallels/prl-devops-service/jobs/tracker"
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

	defer func() {
		if rec := recover(); rec != nil {
			s.ns.NotifyErrorf("AsyncPull panic recovered for job %v: %v", jobId, rec)
			jobManager.MarkJobError(jobId, fmt.Errorf("internal error: %v", rec))
		}
	}()

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

func getPullWorkflowSteps(isCache bool, startAfterPull bool) []tracker.JobStep {
	steps := []tracker.JobStep{
		{Name: constants.ActionValidatingRequest, Weight: 5.0},
		{Name: constants.ActionCheckingLocalCatalog, Weight: 5.0},
		{Name: constants.ActionCheckingRemoteCatalog, Weight: 5.0},
		{Name: constants.ActionDownloadingManifest, Weight: 5.0},
	}

	endStepsLength := 3 // Cleaning, Registering, Renaming
	if startAfterPull {
		endStepsLength = 4 // + Starting
	}

	fixedWeight := 20.0 + (float64(endStepsLength) * 5.0)
	heavyWeight := (100.0 - fixedWeight) / 2.0

	if isCache {
		steps = append(steps, tracker.JobStep{Name: constants.ActionDownloadingPackFile, Weight: heavyWeight / 2.0})
		steps = append(steps, tracker.JobStep{Name: constants.ActionCachingPackFile, Weight: heavyWeight / 2.0})
		steps = append(steps, tracker.JobStep{Name: constants.ActionCopyingFromCache, Weight: heavyWeight})
	} else {
		steps = append(steps, tracker.JobStep{Name: constants.ActionDownloadingPackFile, Weight: heavyWeight})
		steps = append(steps, tracker.JobStep{Name: constants.ActionDecompressingPackFile, Weight: heavyWeight})
	}

	steps = append(steps, tracker.JobStep{Name: constants.ActionCleaningStructure, Weight: 5.0})
	steps = append(steps, tracker.JobStep{Name: constants.ActionRegisteringMachine, Weight: 5.0})
	steps = append(steps, tracker.JobStep{Name: constants.ActionRenamingMachine, Weight: 5.0})
	if startAfterPull {
		steps = append(steps, tracker.JobStep{Name: constants.ActionStartingMachine, Weight: 5.0})
	}

	return steps
}

func (s *CatalogManifestService) PullWithExistingJob(jobId string, r *models.PullCatalogManifestRequest) *models.PullCatalogManifestResponse {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}

	if r == nil {
		response := models.NewPullCatalogManifestResponse()
		response.AddError(errors.New("missing pull request"))
		return response
	}

	r.JobId = jobId
	if r.JobId != "" {
		s.ns.NotifyInfof("JobId attached to Pull request for manifest %v: %v", r.CatalogId, r.JobId)
	}

	response := s.Pull(r)

	return response
}

func (s *CatalogManifestService) Pull(r *models.PullCatalogManifestRequest) *models.PullCatalogManifestResponse {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}
	s.ns.InitJob(r.JobId)

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

	s.ns.NotifyJobMessage(r.JobId, "Configuring the pull job...")

	if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
		s.ns.WithJob(r.JobId, constants.ActionValidatingRequest).NotifyErrorf("Error creating temp dir: %v", err)
		response.AddError(err)
		return response
	}

	// Just for test, setting everything to take way more time
	time.Sleep(2 * time.Second)

	var manifest *models.VirtualMachineCatalogManifest
	provider := models.CatalogManifestProvider{}
	cfg := config.Get()

	if err := provider.Parse(r.Connection); err != nil {
		response.AddError(err)
		return response
	}

	// getting the provider metadata from the database
	if provider.IsRemote() {
		s.ns.NotifyJobMessage(r.JobId, "Checking remote catalog...")
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
					s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: %s", path, clientResponse.ApiError.Message)
					response.AddError(errors.New(clientResponse.ApiError.Message))
					return response
				}
			}
			if clientResponse.StatusCode == 401 || clientResponse.StatusCode == 403 {
				s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: Unauthorized access", path)
				response.AddError(errors.New("Unauthorized access to the catalog manifest"))
				return response
			}
			if clientResponse.StatusCode == 400 {
				s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: Bad request", path)
				response.AddError(errors.New("Bad request to the catalog manifest"))
				return response
			}
			s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: %s", path, err)
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
		s.ns.NotifyJobMessage(r.JobId, "Manifest %s version %s for architecture %s has been downloaded", r.CatalogId, r.Version, arch)
		s.ns.NotifyDebugf("Remote Manifest: %v", manifest)
	} else {
		if disableLocalCatalog {
			response.AddError(errors.New("local catalog is disabled"))
			return response
		}
		s.ns.NotifyJobMessage(r.JobId, "Checking if the manifest exists in the local catalog")
		dto, err := db.GetCatalogManifestByName(s.ctx, r.CatalogId)
		if err != nil {
			manifestErr := errors.Newf("Error getting catalog manifest %v: %v", r.CatalogId, err)
			s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: %s", r.CatalogId, err)
			response.AddError(manifestErr)
			return response
		}
		m := mappers.DtoCatalogManifestToBase(*dto)
		manifest = &m
		s.ns.NotifyJobMessage(r.JobId, "Manifest %s version %s for architecture %s has been downloaded", r.CatalogId, r.Version, m.Architecture)
		s.ns.NotifyDebugf("Local Manifest: %v", manifest)
	}

	// Checking if we have read all of the manifest correctly
	if manifest.CatalogId == "" {
		manifestErr := errors.Newf("manifest %v not found in the catalog", r.CatalogId)
		s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: %s", r.CatalogId, manifestErr)
		response.AddError(manifestErr)
		return response
	}

	if manifest.Provider == nil {
		manifestErr := errors.Newf("Manifest %v does not contain a valid provider", r.CatalogId)
		s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: %s", r.CatalogId, manifestErr)
		response.AddError(manifestErr)
		return response
	}

	// Checking for tainted or revoked manifests
	if manifest.Tainted {
		manifestErr := errors.Newf("manifest %v is tainted", r.CatalogId)
		s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: %s", r.CatalogId, manifestErr)
		response.AddError(manifestErr)
		return response
	}

	// Check if the manifest is revoked
	if manifest.Revoked {
		manifestErr := errors.Newf("manifest %v is revoked", r.CatalogId)
		s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: %s", r.CatalogId, manifestErr)
		response.AddError(manifestErr)
		return response
	}

	// Check if the path for the machine exists
	if !helper.FileExists(r.Path) {
		manifestErr := errors.Newf("path %v does not exist", r.Path)
		s.ns.NotifyJobMessage(r.JobId, "Error getting catalog manifest %s: %s", r.CatalogId, manifestErr)
		response.AddError(manifestErr)
		return response
	}

	response.ID = manifest.ID
	response.CatalogId = manifest.CatalogId
	response.Version = manifest.Version
	response.Manifest = manifest

	s.ns.NotifyJobMessage(r.JobId, "Finding the remote storage provider for %s", r.MachineName)

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

		s.ns.NotifyJobMessage(r.JobId, "Found remote service %v", rs.Name())
		rs.SetJobId(r.JobId)
		foundProvider = true
		r.LocalMachineFolder = fmt.Sprintf("%s.%s", filepath.Join(r.Path, r.MachineName), manifest.Type)

		// Registering the job workflow now that we found the proper provider
		s.ns.RegisterJobWorkflow(r.JobId, []tracker.JobStep{
			{Name: constants.ActionPullValidateStage, Weight: 5, Parallel: false, HasPercentage: true, DisplayName: "Validating Requirements"},
			{Name: constants.ActionPullCheckCacheStage, Weight: 5, Parallel: false, HasPercentage: false, DisplayName: "Checking Cache"},
			{Name: constants.ActionDownloader, Weight: 20, Parallel: rs.CanStream(), HasPercentage: true, DisplayName: "Downloading Pack File"},
			{Name: constants.ActionDecompressor, Weight: 20, Parallel: rs.CanStream(), HasPercentage: true, DisplayName: "Decompressing Pack File"},
			{Name: constants.ActionPullCacheStage, Weight: 20, Parallel: false, HasPercentage: false, DisplayName: "Copy Manifest from cache"},
			{Name: constants.ActionPullRegisterVm, Weight: 10, Parallel: false, HasPercentage: false, DisplayName: "Registering Machine"},
			{Name: constants.ActionPullRenameVm, Weight: 10, Parallel: false, HasPercentage: false, DisplayName: "Renaming Machine"},
			{Name: constants.ActionPullStartVm, Weight: 5, Parallel: false, HasPercentage: false, DisplayName: "Starting Machine"},
			{Name: constants.ActionCleaningUp, Weight: 5, Parallel: false, HasPercentage: false, DisplayName: "Cleaning up"},
		})

		// Checking if the machine already exists before starting the pull process
		s.ns.UpdateStepProgress(r.JobId, constants.ActionPullValidateStage, 10)
		s.ns.UpdateStepMessagef(r.JobId, constants.ActionPullValidateStage, "Checking if the machine %s already exists", r.MachineName)

		filter := fmt.Sprintf("name=%s", r.MachineName)
		vms, err := parallelsDesktopSvc.GetVms(s.ctx, filter)
		if err != nil {
			if errors.GetSystemErrorCode(err) != 404 {
				s.ns.FailStepf(r.JobId, constants.ActionPullValidateStage, "Failed to validate manifest: %v", err)
				response.AddError(err)
				return response
			}
		}

		if len(vms) > 0 {
			err := errors.Newf("machine %v already exists", r.MachineName)
			s.ns.FailStepf(r.JobId, constants.ActionPullValidateStage, "Failed to validate manifest: %v", err)
			response.AddError(err)
			return response
		}

		s.ns.UpdateStepProgress(r.JobId, constants.ActionPullValidateStage, 60)

		time.Sleep(2 * time.Second)

		// Creating the destination folder for the local machine
		if err := s.createDestinationFolder(r, manifest); err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPullValidateStage, "Failed to create destination folder: %v", err)
			response.AddError(err)
			break
		}

		// checking if the manifest is correctly generated
		if manifest.PackFile == "" || manifest.MetadataFile == "" || manifest.Path == "" {
			errorMsg := errors.Newf("Manifest %v is not correctly generated", manifest.Name)
			s.ns.FailStepf(r.JobId, constants.ActionPullValidateStage, "Failed to validate manifest: %v", errorMsg)
			response.AddError(errorMsg)
			break
		}

		s.ns.CompleteStepf(r.JobId, constants.ActionPullValidateStage, "Completed checks")

		time.Sleep(2 * time.Second)

		// checking if we have the caching enabled, if so we will cache the files using the
		// caching service and then pull the files from the cache
		if cfg.IsCatalogCachingEnable() {
			// starting the cache service, the job tracking will now be coordinated by the cache service
			// this means all download/extract/copy from cache will be handled by the cache service
			// and the job tracking will be updated accordingly
			if err := s.pullFromCache(r, manifest, rs); err != nil {
				response.AddError(err)
				break
			}
		} else {
			// Skipping the check cache stage and copy from cache stage so we can update the UI they are not going to happen
			s.ns.SkipStep(r.JobId, constants.ActionPullCheckCacheStage, "Skipping cache check because caching is disabled")
			time.Sleep(2 * time.Second)
			if err := s.pullAndDecompressPackFile(r, manifest, rs); err != nil {
				response.AddError(err)
				break
			}
			s.ns.SkipStep(r.JobId, constants.ActionPullCacheStage, "Skipping cache stage because caching is disabled")
			time.Sleep(2 * time.Second)
		}

		systemSrv := serviceProvider.System
		if r.Owner != "" && r.Owner != "root" {
			if err := systemSrv.ChangeFileUserOwner(s.ctx, r.Owner, r.LocalMachineFolder); err != nil {
				s.ns.WithJob(r.JobId, constants.ActionDownloadingPackFile).NotifyErrorf("Error changing file %v owner to %v: %v", r.LocalMachineFolder, r.Owner, err)
				response.AddError(err)
				break
			}
		}

		if response.HasErrors() {
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			break
		}

		s.ns.WithJob(r.JobId, constants.ActionDownloadingPackFile).NotifyInfof("Finished pulling pack file for manifest %v", manifest.Name)
		break
	}

	if !foundProvider {
		response.AddError(errors.New("No remote service was able to pull the manifest"))
	}

	if response.HasErrors() {
		return response
	}

	if r.LocalMachineFolder == "" {
		s.ns.WithJob(r.JobId, constants.ActionValidatingRequest).NotifyErrorf("No remote service was able to pull the manifest")
		response.AddError(errors.New("No remote service was able to pull the manifest"))
	}

	// Registering
	s.registerMachineWithParallelsDesktop(r, response)

	// Renaming
	s.renameMachineWithParallelsDesktop(r, response)

	// starting the machine
	if r.StartAfterPull {
		s.startMachineWithParallelsDesktop(r, response)
	} else {
		s.ns.SkipStep(r.JobId, constants.ActionPullStartVm, "Skipping starting machine because start after pull is disabled")
	}

	// Cleaning up
	s.CleanPullRequest(r, response)

	return response
}

func (s *CatalogManifestService) registerMachineWithParallelsDesktop(r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	s.ns.StartStepf(r.JobId, constants.ActionPullRegisterVm, "Registering machine %v", r.MachineName)

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
			s.ns.FailStepf(r.JobId, constants.ActionPullRegisterVm, "Error registering machine %v: %v", r.MachineName, err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
		} else {
			s.ns.CompleteStepf(r.JobId, constants.ActionPullRegisterVm, "Completed registering machine %v", r.MachineName)
		}

	} else {
		s.ns.FailStepf(r.JobId, constants.ActionPullRegisterVm, "Error registering machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) renameMachineWithParallelsDesktop(r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	s.ns.StartStepf(r.JobId, constants.ActionPullRenameVm, "Renaming machine %v", r.MachineName)

	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {

		filter := fmt.Sprintf("name=%s", r.MachineName)
		vms, err := parallelsDesktopSvc.GetVms(s.ctx, filter)
		if err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPullRenameVm, "Error getting machine %v: %v", r.MachineName, err)
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
			s.ns.FailStepf(r.JobId, constants.ActionPullRenameVm, "Error getting machine %v: %v", r.MachineName, notFoundError)
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
				s.ns.FailStepf(r.JobId, constants.ActionPullRenameVm, "Error renaming machine %v: %v", r.MachineName, err)
				response.AddError(err)
				response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
				return
			}
		}

		response.MachineID = vms[0].ID
		s.ns.CompleteStepf(r.JobId, constants.ActionPullRenameVm, "Completed renaming machine %v", r.MachineName)
	} else {
		s.ns.FailStepf(r.JobId, constants.ActionPullRenameVm, "Error renaming machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) startMachineWithParallelsDesktop(r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	s.ns.StartStepf(r.JobId, constants.ActionPullStartVm, "Starting machine %v for %v", r.MachineName, r.CatalogId)

	serviceProvider := serviceprovider.Get()
	parallelsDesktopSvc := serviceProvider.ParallelsDesktopService

	if !response.HasErrors() {
		filter := fmt.Sprintf("name=%s", r.MachineName)
		vms, err := parallelsDesktopSvc.GetVms(s.ctx, filter)
		if err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPullStartVm, "Error getting machine %v: %v", r.MachineName, err)
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
			s.ns.FailStepf(r.JobId, constants.ActionPullStartVm, "Error getting machine %v: %v", r.MachineName, notFoundError)
			response.AddError(notFoundError)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}

		if err := parallelsDesktopSvc.StartVm(s.ctx, vm.ID); err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPullStartVm, "Error starting machine %v: %v", r.MachineName, err)
			response.AddError(err)
			response.CleanupRequest.AddLocalFileCleanupOperation(r.LocalMachineFolder, true)
			return
		}
		s.ns.CompleteStepf(r.JobId, constants.ActionPullStartVm, "Completed starting machine %v", r.MachineName)
	} else {
		s.ns.FailStepf(r.JobId, constants.ActionPullStartVm, "Error starting machine %v: %v", r.MachineName, response.Errors)
	}
}

func (s *CatalogManifestService) CleanPullRequest(r *models.PullCatalogManifestRequest, response *models.PullCatalogManifestResponse) {
	s.ns.StartStepf(r.JobId, constants.ActionCleaningUp, "Cleaning up")
	if cleanErrors := response.CleanupRequest.Clean(s.ctx); len(cleanErrors) > 0 {
		s.ns.FailStepf(r.JobId, constants.ActionCleaningUp, "Error cleaning up: %v", cleanErrors)
		for _, err := range cleanErrors {
			response.AddError(err)
		}
	} else {
		s.ns.CompleteStepf(r.JobId, constants.ActionCleaningUp, "Completed cleaning up")
	}
}

func (s *CatalogManifestService) createDestinationFolder(r *models.PullCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest) error {
	s.ns.UpdateStepMessage(r.JobId, constants.ActionPullValidateStage, "Creating folder")
	r.LocalMachineFolder = fmt.Sprintf("%s.%s", filepath.Join(r.Path, r.MachineName), manifest.Type)
	s.ns.UpdateStepMessagef(r.JobId, constants.ActionPullValidateStage, "Local machine folder: %v", r.LocalMachineFolder)
	count := 1
	max_attempts := 30
	created := false
	for {
		if helper.FileExists(r.LocalMachineFolder) {
			s.ns.UpdateStepMessagef(r.JobId, constants.ActionPullValidateStage, "Local machine folder %v already exists, attempting to create a different one", r.LocalMachineFolder)
			r.LocalMachineFolder = fmt.Sprintf("%s_%v.%s", filepath.Join(r.Path, r.MachineName), count, manifest.Type)
			count += 1
			if count > max_attempts {
				s.ns.UpdateStepMessagef(r.JobId, constants.ActionPullValidateStage, "Max attempts reached to find a new local machine folder name, breaking")
				break
			}
		} else {
			created = true
			break
		}
	}
	if !created {
		s.ns.FailStepf(r.JobId, constants.ActionPullValidateStage, "Error creating local machine folder %v", r.LocalMachineFolder)
		return errors.Newf("Error creating local machine folder %v", r.LocalMachineFolder)
	}

	if err := helpers.CreateDirIfNotExist(r.LocalMachineFolder); err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPullValidateStage, "Error creating local machine folder %v: %v", r.LocalMachineFolder, err)
		return err
	}

	s.ns.UpdateStepMessagef(r.JobId, constants.ActionPullValidateStage, "Created local machine folder %v", r.LocalMachineFolder)
	return nil
}

func (s *CatalogManifestService) pullFromCache(r *models.PullCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest, rss interfaces.RemoteStorageService) error {
	cacheService, err := cacheservice.NewCacheService(s.ctx)
	if err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPullCheckCacheStage, "Error creating cache service: %v", err)
		return err
	}

	// Creating the cache request for the service
	cacheRequest := cacheservice.NewCacheRequest(s.ctx, manifest, rss, r.JobId)
	cacheService.WithRequest(cacheRequest)

	// Caching the service if it is not cached, this is were we will be pulling the manifest pack
	if !cacheService.IsCached() {
		s.ns.CompleteStep(r.JobId, constants.ActionPullCheckCacheStage, "Finished checking cache")
		if err := cacheService.Cache(); err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPullCheckCacheStage, "Error caching manifest %v: %v", manifest.Name, err)
			return err
		}
	}

	// if it is cached we need to skip the download and decompress steps
	if cacheService.IsCached() {
		s.ns.SkipStep(r.JobId, constants.ActionDecompressor, "Skipping decompress step")
		s.ns.SkipStep(r.JobId, constants.ActionDownloader, "Skipping download step")
	}

	// We now need to copy the cached folder to the local machine folder
	cacheResponse, err := cacheService.Get()
	if err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPullCacheStage, "Error getting cache response: %v", err)
		return err
	}

	s.ns.CompleteStep(r.JobId, constants.ActionPullCheckCacheStage, "Finished checking cache")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch a goroutine to poll copy progress and update the job manager
	srcSize, err := helpers.DirSize(cacheResponse.PackFilePath)
	if err == nil && srcSize > 0 {
		copyStart := time.Now()
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					dstSize, err := helpers.DirSize(r.LocalMachineFolder)
					if err == nil {
						percentage := int((float64(dstSize) / float64(srcSize)) * 100)
						if percentage > 100 {
							percentage = 100
						}

						if s.ns != nil {
							s.ns.UpdateStepMessagef(r.JobId, constants.ActionPullCacheStage,
								"Copying from cache")
							msg := tracker.NewJobProgressMessage(
								r.JobId,
								constants.ActionPullCacheStage,
								float64(percentage),
							).
								SetCurrentSize(dstSize).
								SetTotalSize(srcSize).
								SetStartingTime(copyStart).
								SetJobId(r.JobId).
								SetCurrentAction(constants.ActionPullCacheStage)
							s.ns.Notify(msg)
						}
					}
				}
			}
		}()
	}

	if err := helpers.CopyDir(cacheResponse.PackFilePath, r.LocalMachineFolder); err != nil {
		s.ns.WithJob(r.JobId, constants.ActionPullCacheStage).NotifyErrorf("Error copying cached folder %v to %v: %v", cacheResponse.PackFilePath, r.LocalMachineFolder, err)
		return err
	}
	s.ns.CompleteStepf(r.JobId, constants.ActionPullCacheStage, "Finished copying cached folder %v to %v", cacheResponse.PackFilePath, r.LocalMachineFolder)
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
		s.ns.Notify(tracker.NewJobProgressMessage(r.JobId, constants.ActionCleaningStructure, 100).
			SetJobId(r.JobId).
			SetCurrentAction(constants.ActionCleaningStructure).
			SetFilename(manifest.Name))
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
			// Job management is now handled exclusively through NotificationService
		}

		if err := compressor.DecompressFileWithStepChannel(s.ctx, compressedFilePath, destinationFolder, nil, r.JobId, constants.ActionDecompressingPackFile); err != nil {
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
