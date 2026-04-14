package catalog

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/interfaces"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/jobs/tracker"
	"github.com/Parallels/prl-devops-service/mappers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/helper/http_helper"
)

func (s *CatalogManifestService) AsyncPush(jobId string, r *models.PushCatalogManifestRequest) {
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
			s.ns.NotifyErrorf("AsyncPush panic recovered for job %v: %v", jobId, rec)
			jobManager.MarkJobError(jobId, fmt.Errorf("internal error: %v", rec))
		}
	}()

	response := s.PushWithExistingJob(jobId, r)
	if response.HasErrors() {
		errorMessage := "Error pushing manifest:"
		for _, err := range response.Errors {
			errorMessage += fmt.Sprintf("\n%v", err)
		}
		jobManager.MarkJobError(jobId, errors.New(errorMessage))
	} else {
		jobManager.MarkJobCompleteWithRecord(jobId, "Catalog Manifest Pushed", response.ID, response.Name, "catalog_manifest", "")
	}
}

func (s *CatalogManifestService) PushWithExistingJob(jobId string, r *models.PushCatalogManifestRequest) *models.VirtualMachineCatalogManifest {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}

	r.JobId = jobId
	if r.JobId != "" {
		s.ns.NotifyInfof("JobId attached to Push request for manifest %v: %v", r.CatalogId, r.JobId)
	}

	return s.Push(r)
}

func (s *CatalogManifestService) Push(r *models.PushCatalogManifestRequest) *models.VirtualMachineCatalogManifest {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}
	s.ns.InitJob(r.JobId)

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
		if r.JobId != "" {
			s.ns.NotifyDebugf("Setting job id for remote service %v", rs.Name())
			rs.SetJobId(r.JobId)
		}

		// Register the job workflow now that we found the proper provider
		s.ns.RegisterJobWorkflow(r.JobId, []tracker.JobStep{
			{Name: constants.ActionPushValidateStage, Weight: 5, Parallel: false, HasPercentage: false, DisplayName: "Validating Request"},
			{Name: constants.ActionPushCompressStage, Weight: 35, Parallel: false, HasPercentage: true, DisplayName: "Compressing Pack File"},
			{Name: constants.ActionPushCheckRemoteStage, Weight: 5, Parallel: false, HasPercentage: false, DisplayName: "Checking Remote Catalog"},
			{Name: constants.ActionPushUploadPackStage, Weight: 45, Parallel: false, HasPercentage: true, DisplayName: "Uploading Pack File"},
			{Name: constants.ActionPushUploadMetaStage, Weight: 5, Parallel: false, HasPercentage: false, DisplayName: "Uploading Metadata"},
			{Name: constants.ActionPushRegisterStage, Weight: 3, Parallel: false, HasPercentage: false, DisplayName: "Registering Manifest"},
			{Name: constants.ActionCleaningUp, Weight: 2, Parallel: false, HasPercentage: false, DisplayName: "Cleaning up"},
		})

		manifest.CleanupRequest.RemoteStorageService = rs
		apiClient := apiclient.NewHttpClient(s.ctx)

		// Validate stage
		s.ns.StartStepf(r.JobId, constants.ActionPushValidateStage, "Validating push request for %v", r.CatalogId)

		if err := manifest.Provider.Parse(r.Connection); err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPushValidateStage, "Error parsing provider %v: %v", r.Connection, err)
			manifest.AddError(err)
			break
		}

		if manifest.Provider.IsRemote() {
			s.ns.NotifyDebugf("Testing remote provider %v", manifest.Provider.Host)
			apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
		}

		// // We now need to ask the provider if we already have this manifest
		// manifestPath := filepath.Join(rs.GetProviderRootPath(s.ctx), manifest.CatalogId)
		// exists, _ := rs.FileExists(s.ctx, manifestPath, s.getMetaFilename(manifest.Name))
		// if exists && !r.OverrideExisting {
		// 	s.ns.FailStepf(r.JobId, constants.ActionPushValidateStage, "Manifest %v already exists", r.CatalogId)
		// 	manifest.AddError(errors.NewWithCode("Manifest already exists", 409))
		// 	break
		// } else {
		// 	s.ns.NotifyInfof("Manifest %v already exists, overriding it", r.CatalogId)
		// 	if err := rs.DeleteFile(s.ctx, manifestPath, s.getMetaFilename(manifest.Name)); err != nil {
		// 		s.ns.FailStepf(r.JobId, constants.ActionPushValidateStage, "Error deleting manifest %v: %v", r.CatalogId, err)
		// 		manifest.AddError(err)
		// 		break
		// 	}
		// 	if err := rs.DeleteFile(s.ctx, manifestPath, s.getPackFilename(manifest.Name)); err != nil {
		// 		s.ns.FailStepf(r.JobId, constants.ActionPushValidateStage, "Error deleting manifest %v: %v", r.CatalogId, err)
		// 		manifest.AddError(err)
		// 		break
		// 	}
		// }

		s.ns.CompleteStepf(r.JobId, constants.ActionPushValidateStage, "Validation complete for %v", r.CatalogId)

		// Compress stage - generating manifest content
		s.ns.StartStepf(r.JobId, constants.ActionPushCompressStage, "Compressing manifest files for %v", r.CatalogId)
		s.ns.NotifyInfof("Pushing manifest %v to provider %s", r.CatalogId, rs.Name())
		err = s.GenerateManifestContent(r, manifest)
		if err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPushCompressStage, "Error generating manifest content for %v: %v", r.CatalogId, err)
			manifest.AddError(err)
			break
		}
		s.ns.CompleteStepf(r.JobId, constants.ActionPushCompressStage, "Compression complete for %v", r.CatalogId)

		if err := helpers.CreateDirIfNotExist("/tmp"); err != nil {
			s.ns.NotifyErrorf("Error creating temp dir: %v", err)
		}

		// Check remote stage - checking if the manifest metadata exists in the remote server
		s.ns.StartStepf(r.JobId, constants.ActionPushCheckRemoteStage, "Checking remote catalog for %v", r.CatalogId)
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
					s.ns.FailStepf(r.JobId, constants.ActionPushCheckRemoteStage, "Error reading manifest from file %v: %v", tmpCatalogManifestFilePath, err)
					manifest.AddError(err)
					break
				}
				manifest.CreatedAt = catalogManifest.CreatedAt
			}
		}
		s.ns.CompleteStepf(r.JobId, constants.ActionPushCheckRemoteStage, "Remote check complete for %v", r.CatalogId)

		if catalogManifest != nil {
			if err := s.pushUpdateExistingManifest(r, manifest, catalogManifest, rs); err != nil {
				break
			}
		} else {
			if err := s.pushCreateNewManifest(r, manifest, rs); err != nil {
				break
			}
		}

		// Register stage - add the manifest to the database or update it
		if !manifest.HasErrors() {
			s.ns.StartStepf(r.JobId, constants.ActionPushRegisterStage, "Registering manifest %v", r.CatalogId)
			if err := s.registerManifest(r, manifest, apiClient); err != nil {
				s.ns.FailStepf(r.JobId, constants.ActionPushRegisterStage, "Error registering manifest %v: %v", r.CatalogId, err)
				manifest.AddError(err)
				break
			}
			s.ns.CompleteStepf(r.JobId, constants.ActionPushRegisterStage, "Registration complete for %v", r.CatalogId)
		}
		break
	}

	if !executed {
		manifest.AddError(errors.Newf("no remote service found for connection %v", r.Connection))
	}

	// Cleanup stage (best-effort: errors are logged but do not fail the job)
	s.ns.StartStepf(r.JobId, constants.ActionCleaningUp, "Cleaning up for %v", r.CatalogId)
	if cleanErrors := manifest.CleanupRequest.Clean(s.ctx); len(cleanErrors) > 0 {
		for _, err := range cleanErrors {
			s.ns.NotifyWarningf("Cleanup warning for %v: %v", r.CatalogId, err)
		}
		s.ns.CompleteStepf(r.JobId, constants.ActionCleaningUp, "Cleanup completed with warnings for %v", r.CatalogId)
	} else {
		s.ns.CompleteStepf(r.JobId, constants.ActionCleaningUp, "Cleanup complete for %v", r.CatalogId)
	}

	return manifest
}

func (s *CatalogManifestService) pushUpdateExistingManifest(r *models.PushCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest, catalogManifest *models.VirtualMachineCatalogManifest, rs interfaces.RemoteStorageService) error {
	manifest.Path = catalogManifest.Path
	manifest.MetadataFile = s.getMetaFilename(catalogManifest.Name)
	manifest.PackFile = s.getPackFilename(catalogManifest.Name)
	s.applyMinimumSpecRequirements(r, manifest)
	localPackPath := filepath.Dir(manifest.CompressedPath)

	s.ns.NotifyInfof("Found remote catalog manifest, checking if the files are up to date")
	s.ns.StartStepf(r.JobId, constants.ActionPushUploadPackStage, "Checking pack file for %v", r.CatalogId)
	remotePackChecksum, err := rs.FileChecksum(s.ctx, catalogManifest.Path, catalogManifest.PackFile)
	if err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPushUploadPackStage, "Error getting remote pack checksum %v: %v", catalogManifest.PackFile, err)
		manifest.AddError(err)
		return err
	}
	if remotePackChecksum != manifest.CompressedChecksum {
		s.ns.NotifyInfof("Remote pack is not up to date, pushing it")
		rs.SetCurrentAction(constants.ActionPushUploadPackStage)
		if err := rs.PushFile(s.ctx, localPackPath, catalogManifest.Path, catalogManifest.PackFile); err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPushUploadPackStage, "Error pushing pack file %v: %v", catalogManifest.PackFile, err)
			manifest.AddError(err)
			return err
		}
	} else {
		s.ns.NotifyInfof("Remote pack is up to date")
	}
	s.ns.CompleteStepf(r.JobId, constants.ActionPushUploadPackStage, "Pack upload complete for %v", r.CatalogId)

	manifest.PackContents = append(manifest.PackContents, models.VirtualMachineManifestContentItem{
		Path:      manifest.Path,
		IsDir:     false,
		Name:      filepath.Base(manifest.PackFile),
		Checksum:  manifest.CompressedChecksum,
		CreatedAt: helpers.GetUtcCurrentDateTime(),
		UpdatedAt: helpers.GetUtcCurrentDateTime(),
	})

	s.ns.StartStepf(r.JobId, constants.ActionPushUploadMetaStage, "Uploading metadata for %v", r.CatalogId)
	tempManifestContentFilePath := filepath.Join("/tmp", manifest.MetadataFile)
	cleanManifest := *manifest
	cleanManifest.Provider = nil
	manifestContent, err := json.MarshalIndent(cleanManifest, "", "  ")
	if err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPushUploadMetaStage, "Error marshalling manifest %v: %v", cleanManifest, err)
		manifest.AddError(err)
		return err
	}

	manifest.CleanupRequest.AddLocalFileCleanupOperation(tempManifestContentFilePath, false)
	if err := helper.WriteToFile(string(manifestContent), tempManifestContentFilePath); err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPushUploadMetaStage, "Error writing manifest to temporary file %v: %v", tempManifestContentFilePath, err)
		manifest.AddError(err)
		return err
	}

	metadataChecksum, err := helpers.GetFileMD5Checksum(tempManifestContentFilePath)
	if err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPushUploadMetaStage, "Error getting metadata checksum %v: %v", tempManifestContentFilePath, err)
		manifest.AddError(err)
		return err
	}

	remoteMetadataChecksum, err := rs.FileChecksum(s.ctx, catalogManifest.Path, catalogManifest.MetadataFile)
	if err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPushUploadMetaStage, "Error getting remote metadata checksum %v: %v", catalogManifest.MetadataFile, err)
		manifest.AddError(err)
		return err
	}

	if remoteMetadataChecksum != metadataChecksum {
		s.ns.NotifyInfof("Remote metadata is not up to date, pushing it")
		if err := rs.PushFile(s.ctx, "/tmp", catalogManifest.Path, manifest.MetadataFile); err != nil {
			s.ns.FailStepf(r.JobId, constants.ActionPushUploadMetaStage, "Error pushing metadata file %v: %v", catalogManifest.MetadataFile, err)
			manifest.AddError(err)
			return err
		}
	} else {
		s.ns.NotifyInfof("Remote metadata is up to date")
	}
	s.ns.CompleteStepf(r.JobId, constants.ActionPushUploadMetaStage, "Metadata upload complete for %v", r.CatalogId)

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

	return nil
}

func (s *CatalogManifestService) pushCreateNewManifest(r *models.PushCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest, rs interfaces.RemoteStorageService) error {
	s.ns.NotifyInfof("Remote Manifest metadata not found, creating it")

	manifest.Path = filepath.Join(rs.GetProviderRootPath(s.ctx), manifest.CatalogId)
	manifest.MetadataFile = s.getMetaFilename(manifest.Name)
	manifest.PackFile = s.getPackFilename(manifest.Name)
	s.applyMinimumSpecRequirements(r, manifest)
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
		return err
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
		return err
	}

	manifest.CleanupRequest.AddLocalFileCleanupOperation(tempManifestContentFilePath, false)
	if err := helper.WriteToFile(string(manifestContent), tempManifestContentFilePath); err != nil {
		s.ns.NotifyErrorf("Error writing manifest to temporary file %v: %v", tempManifestContentFilePath, err)
		manifest.AddError(err)
		return err
	}

	s.ns.StartStepf(r.JobId, constants.ActionPushUploadPackStage, "Uploading pack file for %v", r.CatalogId)
	s.ns.NotifyInfof("Pushing manifest pack file %v", manifest.PackFile)
	localPackPath := filepath.Dir(manifest.CompressedPath)
	rs.SetCurrentAction(constants.ActionPushUploadPackStage)
	if err := rs.PushFile(s.ctx, localPackPath, manifest.Path, manifest.PackFile); err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPushUploadPackStage, "Error pushing pack file %v: %v", manifest.PackFile, err)
		manifest.AddError(err)
		return err
	}
	s.ns.CompleteStepf(r.JobId, constants.ActionPushUploadPackStage, "Pack upload complete for %v", r.CatalogId)

	s.ns.StartStepf(r.JobId, constants.ActionPushUploadMetaStage, "Uploading metadata for %v", r.CatalogId)
	s.ns.NotifyInfof("Pushing manifest meta file %v", manifest.MetadataFile)
	if err := rs.PushFile(s.ctx, "/tmp", manifest.Path, manifest.MetadataFile); err != nil {
		s.ns.FailStepf(r.JobId, constants.ActionPushUploadMetaStage, "Error pushing metadata file %v: %v", manifest.MetadataFile, err)
		manifest.AddError(err)
		return err
	}
	s.ns.CompleteStepf(r.JobId, constants.ActionPushUploadMetaStage, "Metadata upload complete for %v", r.CatalogId)

	if manifest.HasErrors() {
		manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.PackFile), false)
		manifest.CleanupRequest.AddRemoteFileCleanupOperation(filepath.Join(manifest.Path, manifest.MetadataFile), false)
		manifest.CleanupRequest.AddRemoteFileCleanupOperation(manifest.Path, true)
	}

	return nil
}

func (s *CatalogManifestService) registerManifest(r *models.PushCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest, apiClient *apiclient.HttpClientService) error {
	if manifest.Provider.IsRemote() {
		s.ns.NotifyInfof("Manifest pushed successfully, adding it to the remote database")
		apiClient.SetAuthorization(GetAuthenticator(manifest.Provider))
		path := http_helper.JoinUrl(constants.DEFAULT_API_PREFIX, "catalog")

		var response api_models.CatalogManifest
		postUrl := fmt.Sprintf("%s%s", manifest.Provider.GetUrl(), path)
		if _, err := apiClient.Post(postUrl, manifest, &response); err != nil {
			s.ns.NotifyErrorf("Error posting catalog manifest %v: %v", manifest.Provider.String(), err)
			return err
		}

		manifest.ID = response.ID
		manifest.Name = response.Name
		manifest.CatalogId = response.CatalogId
	} else {
		s.ns.NotifyInfof("Manifest pushed successfully, adding it to the database")
		db := serviceprovider.Get().JsonDatabase
		if err := db.Connect(s.ctx); err != nil {
			return err
		}

		exists, _ := db.GetCatalogManifestsByCatalogIdVersionAndArch(s.ctx, manifest.CatalogId, manifest.Version, manifest.Architecture)
		if exists != nil {
			s.ns.NotifyInfof("Updating manifest %v", manifest.Name)
			dto := mappers.CatalogManifestToDto(*manifest)
			dto.ID = exists.ID
			if _, err := db.UpdateCatalogManifest(s.ctx, dto); err != nil {
				s.ns.NotifyErrorf("Error updating manifest %v: %v", manifest.Name, err)
				return err
			}
		} else {
			s.ns.NotifyInfof("Creating manifest %v", manifest.Name)
			dto := mappers.CatalogManifestToDto(*manifest)
			if _, err := db.CreateCatalogManifest(s.ctx, dto); err != nil {
				s.ns.NotifyErrorf("Error creating manifest %v: %v", manifest.Name, err)
				return err
			}
		}
	}
	return nil
}

func (s *CatalogManifestService) applyMinimumSpecRequirements(r *models.PushCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest) {
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
}
