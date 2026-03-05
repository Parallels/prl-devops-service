package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/jobs/tracker"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerJobsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Jobs handlers", version)

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/jobs").
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_OWN_CLAIM).
		WithOrClaims().
		WithHandler(GetJobsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/jobs/{id}").
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_OWN_CLAIM).
		WithOrClaims().
		WithHandler(GetJobHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/jobs/cleanup").
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_OWN_CLAIM).
		WithOrClaims().
		WithHandler(CleanupJobsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/jobs/debug").
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_OWN_CLAIM).
		WithOrClaims().
		WithHandler(DebugJobHandler()).
		Register()
}

func GetJobsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "User not found"})
			return
		}

		authCtx := ctx.GetAuthorizationContext()
		canListAll := authCtx != nil && authCtx.UserHasClaim("job_manager_list")

		var jobs []models.JobResponse
		if canListAll {
			dbJobs, err := dbService.GetJobs(ctx)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
				return
			}
			for _, dbJob := range dbJobs {
				jobs = append(jobs, *mappers.MapJobToApiJob(dbJob))
			}
		} else {
			dbJobs, err := dbService.GetJobsByOwner(ctx, userContext.ID)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
				return
			}
			for _, dbJob := range dbJobs {
				jobs = append(jobs, *mappers.MapJobToApiJob(dbJob))
			}
		}

		if jobs == nil {
			jobs = make([]models.JobResponse, 0)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(jobs)
		ctx.LogInfof("Jobs returned successfully")
	}
}

func GetJobHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "User not found"})
			return
		}

		vars := mux.Vars(r)
		jobId := vars["id"]

		dbJob, err := dbService.GetJob(ctx, jobId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusNotFound))
			return
		}

		authCtx := ctx.GetAuthorizationContext()
		canListAll := authCtx != nil && authCtx.UserHasClaim("job_manager_list")
		if !canListAll && !strings.EqualFold(dbJob.Owner, userContext.ID) {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusForbidden, Message: "Forbidden to view this job"})
			return
		}

		response := mappers.MapJobToApiJob(*dbJob)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Job returned successfully")
	}
}

func CleanupJobsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "User not found"})
			return
		}

		authCtx := ctx.GetAuthorizationContext()
		canListAll := authCtx != nil && authCtx.UserHasClaim("job_manager_list")

		if canListAll {
			err = dbService.DeleteJobsByState(ctx, constants.JobStateCompleted, constants.JobStateFailed)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
				return
			}
		} else {
			dbJobs, err := dbService.GetJobsByOwner(ctx, userContext.ID)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
				return
			}

			for _, dbJob := range dbJobs {
				if dbJob.State == constants.JobStateCompleted || dbJob.State == constants.JobStateFailed {
					_ = dbService.DeleteJob(ctx, dbJob.ID)
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		ctx.LogInfof("Jobs cleanup completed successfully")
	}
}

// DebugJobHandler kicks off a background scenario chosen by `profile` in the request body.
//
// Supported profiles:
//   - "simple"        – linear 0→100% over 20 s (default)
//   - "pull_remote"   – full multi-step remote pull with parallel download & decompress
//   - "pull_cache"    – fast cached pull: validate → copy-from-cache → register
//   - "skipped_steps" – some steps are instant-completed (cache-hit scenario)
func DebugJobHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "User not found"})
			return
		}

		var request models.JobCreateRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"})
			return
		}

		jobManager := jobs.Get(ctx)
		if jobManager == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Job Manager is not available"), http.StatusInternalServerError))
			return
		}

		action := request.Action
		if action == "" {
			action = "Debug Task"
		}

		job, err := jobManager.CreateNewJob(userContext.ID, request.JobType, request.JobOperation, action)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		profile := strings.ToLower(strings.TrimSpace(request.Profile))
		switch profile {
		case "pull_remote":
			go runDebugProfilePullRemote(job.ID, jobManager)
		case "pull_cache":
			go runDebugProfilePullCache(job.ID, jobManager)
		case "skipped_steps":
			go runDebugProfileSkippedSteps(job.ID, jobManager)
		default: // "simple" or ""
			go runDebugProfileSimple(job.ID, jobManager)
		}

		response := mappers.MapJobToApiJob(*job)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// ---------------------------------------------------------------------------
// Profile: simple – linear 0→100% in 10 ticks of 2 s each
// ---------------------------------------------------------------------------

func runDebugProfileSimple(jobId string, jobManager *jobs.JobManagerService) {
	ns := tracker.GetProgressService()
	bCtx := basecontext.NewRootBaseContext()
	bCtx.LogInfof("[Debug/simple] Starting job %s", jobId)

	if ns != nil {
		ns.NotifyJobMessage(jobId, "Initializing debug background environment...")
		time.Sleep(1 * time.Second)
		ns.NotifyJobMessage(jobId, "Preparing execution workspace...")
		time.Sleep(1 * time.Second)
	}

	for i := 1; i <= 10; i++ {
		time.Sleep(2 * time.Second)
		_, _ = jobManager.UpdateJobProgress(jobId, i*10, constants.JobStateRunning)
	}

	recordId := "rec_" + helpers.GenerateId()[:8]
	_, _ = jobManager.UpdateJobResultRecord(jobId, recordId, "debug_simple_report")
	_ = jobManager.MarkJobComplete(jobId, "Simple debug task finished")
	bCtx.LogInfof("[Debug/simple] Job %s completed", jobId)
}

// ---------------------------------------------------------------------------
// Profile: pull_remote – full pull workflow mimicking catalog/pull.go
//   Download and decompress run concurrently, just like in production.
// ---------------------------------------------------------------------------

func runDebugProfilePullRemote(jobId string, jobManager *jobs.JobManagerService) {
	ns := tracker.GetProgressService()
	if ns == nil {
		return
	}
	bCtx := basecontext.NewRootBaseContext()
	bCtx.LogInfof("[Debug/pull_remote] Starting job %s", jobId)

	// Register the same step configuration as catalog/pull.go (no caching)
	ns.RegisterJobWorkflow(jobId, []tracker.JobStep{
		{Name: constants.ActionValidatingRequest, Weight: 2, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCheckingLocalCatalog, Weight: 3, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCheckingRemoteCatalog, Weight: 5, Parallel: false, HasPercentage: false},
		{Name: constants.ActionDownloadingManifest, Weight: 5, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCreatingDestinationFolder, Weight: 1, Parallel: false, HasPercentage: false},
		{Name: constants.ActionDownloadingPackFile, Weight: 45, Parallel: true, HasPercentage: true},
		{Name: constants.ActionDecompressingPackFile, Weight: 30, Parallel: true, HasPercentage: true},
		{Name: constants.ActionCleaningStructure, Weight: 2, Parallel: false, HasPercentage: false},
		{Name: constants.ActionRegisteringMachine, Weight: 4, Parallel: false, HasPercentage: false},
		{Name: constants.ActionRenamingMachine, Weight: 2, Parallel: false, HasPercentage: false},
		{Name: constants.ActionStartingMachine, Weight: 1, Parallel: false, HasPercentage: false},
	})

	ns.NotifyJobMessage(jobId, "Connecting to external Parallels Hypervisor API...")
	time.Sleep(1 * time.Second)
	ns.NotifyJobMessage(jobId, "Validating VM network port availability...")
	time.Sleep(1 * time.Second)

	instantStep := func(action, msg string) {
		ns.WithJob(jobId, action).NotifyInfof("%s", msg)
		ns.Notify(tracker.NewJobProgressMessage(jobId, action, 100).
			SetJobId(jobId).SetCurrentAction(action))
		time.Sleep(300 * time.Millisecond)
	}

	// Fast sequential steps
	instantStep(constants.ActionValidatingRequest, "Validating debug pull-remote request")
	instantStep(constants.ActionCheckingLocalCatalog, "debug-vm not found in local catalog — will pull from remote")
	instantStep(constants.ActionCheckingRemoteCatalog, "Connected to minio://demo-provider — manifest found")
	instantStep(constants.ActionDownloadingManifest, "Manifest downloaded (12 KB)")
	instantStep(constants.ActionCreatingDestinationFolder, "Created /tmp/debug-vm.pvm")

	// --- Parallel download + decompression ---
	const totalBytes int64 = 10 * 1024 * 1024 * 1024 // pretend 10 GB
	dlCorrId := "debug-dl-" + jobId
	dcCorrId := "debug-dc-" + jobId

	var wg sync.WaitGroup
	wg.Add(2)

	// Download: 1% per 200 ms → 20 s total
	go func() {
		defer wg.Done()
		startTime := time.Now()
		for pct := 0.0; pct <= 100.0; pct += 1.0 {
			curBytes := int64(pct / 100.0 * float64(totalBytes))
			ns.Notify(tracker.NewJobProgressMessage(dlCorrId, "Downloading debug-vm.pdpack", pct).
				SetCurrentSize(curBytes).
				SetTotalSize(totalBytes).
				SetStartingTime(startTime).
				SetJobId(jobId).
				SetCurrentAction(constants.ActionDownloadingPackFile).
				SetFilename("debug-vm.pdpack"))
			time.Sleep(200 * time.Millisecond)
		}
		ns.WithJob(jobId, constants.ActionDownloadingPackFile).NotifyInfof("Download of debug-vm.pdpack complete")
	}()

	// Decompress: starts 5 s after download begins, slower pace (1% per 300 ms → 30 s)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Second) // wait for some data first
		startTime := time.Now()
		for pct := 0.0; pct <= 100.0; pct += 1.0 {
			curBytes := int64(pct / 100.0 * float64(totalBytes))
			ns.Notify(tracker.NewJobProgressMessage(dcCorrId, "Decompressing debug-vm.pdpack", pct).
				SetCurrentSize(curBytes).
				SetTotalSize(totalBytes).
				SetStartingTime(startTime).
				SetJobId(jobId).
				SetCurrentAction(constants.ActionDecompressingPackFile).
				SetFilename("debug-vm.pdpack"))
			time.Sleep(300 * time.Millisecond)
		}
		ns.WithJob(jobId, constants.ActionDecompressingPackFile).NotifyInfof("Decompression of debug-vm.pdpack complete")
	}()

	wg.Wait()

	// Finishing steps
	instantStep(constants.ActionCleaningStructure, "Flattening directory structure")
	time.Sleep(500 * time.Millisecond)

	ns.WithJob(jobId, constants.ActionRegisteringMachine).NotifyInfof("Registering debug-vm with Parallels Desktop")
	time.Sleep(2 * time.Second)
	instantStep(constants.ActionRegisteringMachine, "debug-vm registered successfully")
	instantStep(constants.ActionRenamingMachine, "Machine renamed to debug-vm")
	instantStep(constants.ActionStartingMachine, "debug-vm started")

	recordId := "rec_remote_" + helpers.GenerateId()[:8]
	_, _ = jobManager.UpdateJobResultRecord(jobId, recordId, "vm_deployment_record")

	_ = jobManager.MarkJobComplete(jobId, "pull_remote debug task finished")
	bCtx.LogInfof("[Debug/pull_remote] Job %s completed", jobId)
}

// ---------------------------------------------------------------------------
// Profile: pull_cache – pull from local cache, skips remote download/decompress
// ---------------------------------------------------------------------------

func runDebugProfilePullCache(jobId string, jobManager *jobs.JobManagerService) {
	ns := tracker.GetProgressService()
	if ns == nil {
		return
	}
	bCtx := basecontext.NewRootBaseContext()
	bCtx.LogInfof("[Debug/pull_cache] Starting job %s", jobId)

	ns.RegisterJobWorkflow(jobId, []tracker.JobStep{
		{Name: constants.ActionValidatingRequest, Weight: 3, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCheckingLocalCatalog, Weight: 3, Parallel: false, HasPercentage: false},
		{Name: constants.ActionDownloadingManifest, Weight: 4, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCreatingDestinationFolder, Weight: 2, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCachingPackFile, Weight: 3, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCopyingFromCache, Weight: 60, Parallel: false, HasPercentage: true},
		{Name: constants.ActionCleaningStructure, Weight: 5, Parallel: false, HasPercentage: false},
		{Name: constants.ActionRegisteringMachine, Weight: 10, Parallel: false, HasPercentage: false},
		{Name: constants.ActionRenamingMachine, Weight: 5, Parallel: false, HasPercentage: false},
		{Name: constants.ActionStartingMachine, Weight: 5, Parallel: false, HasPercentage: false},
	})

	ns.NotifyJobMessage(jobId, "Verifying local Parallels cache integrity...")
	time.Sleep(1 * time.Second)
	ns.NotifyJobMessage(jobId, "Locking local catalog directory...")
	time.Sleep(1 * time.Second)

	instantStep := func(action, msg string) {
		ns.WithJob(jobId, action).NotifyInfof("%s", msg)
		ns.Notify(tracker.NewJobProgressMessage(jobId, action, 100).
			SetJobId(jobId).SetCurrentAction(action))
		time.Sleep(400 * time.Millisecond)
	}

	instantStep(constants.ActionValidatingRequest, "Validating debug pull-cache request")
	instantStep(constants.ActionCheckingLocalCatalog, "debug-vm found in local catalog")
	instantStep(constants.ActionDownloadingManifest, "Manifest already cached")
	instantStep(constants.ActionCreatingDestinationFolder, "Created /tmp/debug-vm.pvm")
	instantStep(constants.ActionCachingPackFile, "debug-vm pack-file is already in cache — using existing")

	// Copy from cache: 1% per 150 ms → ~15 s
	const totalBytes int64 = 10 * 1024 * 1024 * 1024
	cacheId := "debug-cache-" + jobId
	startTime := time.Now()
	ns.WithJob(jobId, constants.ActionCopyingFromCache).NotifyInfof("Copying debug-vm from cache to /tmp/debug-vm.pvm")
	for pct := 0.0; pct <= 100.0; pct += 1.0 {
		ns.Notify(tracker.NewJobProgressMessage(cacheId, "Copying from cache", pct).
			SetCurrentSize(int64(pct / 100.0 * float64(totalBytes))).
			SetTotalSize(totalBytes).
			SetStartingTime(startTime).
			SetJobId(jobId).
			SetCurrentAction(constants.ActionCopyingFromCache).
			SetFilename("debug-vm.pdpack"))
		time.Sleep(150 * time.Millisecond)
	}
	ns.WithJob(jobId, constants.ActionCopyingFromCache).NotifyInfof("Copy complete")

	instantStep(constants.ActionCleaningStructure, "Flattening directory structure")
	time.Sleep(1 * time.Second)
	instantStep(constants.ActionRegisteringMachine, "Registering debug-vm with Parallels Desktop")
	instantStep(constants.ActionRenamingMachine, "Machine renamed to debug-vm")
	instantStep(constants.ActionStartingMachine, "debug-vm started")

	recordId := "rec_cache_" + helpers.GenerateId()[:8]
	_, _ = jobManager.UpdateJobResultRecord(jobId, recordId, "vm_deployment_record")

	_ = jobManager.MarkJobComplete(jobId, "pull_cache debug task finished")
	bCtx.LogInfof("[Debug/pull_cache] Job %s completed", jobId)
}

// ---------------------------------------------------------------------------
// Profile: skipped_steps – download & decompress are instantly marked done (cache-hit)
//   and copy-from-cache is the main operation.  Tests UI step collapse/fast-forward.
// ---------------------------------------------------------------------------

func runDebugProfileSkippedSteps(jobId string, jobManager *jobs.JobManagerService) {
	ns := tracker.GetProgressService()
	if ns == nil {
		return
	}
	bCtx := basecontext.NewRootBaseContext()
	bCtx.LogInfof("[Debug/skipped_steps] Starting job %s", jobId)

	ns.RegisterJobWorkflow(jobId, []tracker.JobStep{
		{Name: constants.ActionValidatingRequest, Weight: 2, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCheckingLocalCatalog, Weight: 3, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCheckingRemoteCatalog, Weight: 5, Parallel: false, HasPercentage: false},
		{Name: constants.ActionDownloadingManifest, Weight: 5, Parallel: false, HasPercentage: false},
		{Name: constants.ActionCreatingDestinationFolder, Weight: 1, Parallel: false, HasPercentage: false},
		// These two are instantly skipped (cache hit)
		{Name: constants.ActionDownloadingPackFile, Weight: 0, Parallel: false, HasPercentage: true},
		{Name: constants.ActionDecompressingPackFile, Weight: 0, Parallel: false, HasPercentage: true},
		// Main work is copying from cache
		{Name: constants.ActionCopyingFromCache, Weight: 70, Parallel: false, HasPercentage: true},
		{Name: constants.ActionCleaningStructure, Weight: 3, Parallel: false, HasPercentage: false},
		{Name: constants.ActionRegisteringMachine, Weight: 5, Parallel: false, HasPercentage: false},
		{Name: constants.ActionRenamingMachine, Weight: 3, Parallel: false, HasPercentage: false},
		{Name: constants.ActionStartingMachine, Weight: 3, Parallel: false, HasPercentage: false},
	})

	ns.NotifyJobMessage(jobId, "Checking Minio minio://demo-provider credentials...")
	time.Sleep(800 * time.Millisecond)
	ns.NotifyJobMessage(jobId, "Scanning fast-path overrides for demo deployment...")
	time.Sleep(800 * time.Millisecond)

	instantStep := func(action, msg string) {
		ns.WithJob(jobId, action).NotifyInfof("%s", msg)
		ns.Notify(tracker.NewJobProgressMessage(jobId, action, 100).
			SetJobId(jobId).SetCurrentAction(action))
		time.Sleep(300 * time.Millisecond)
	}

	skip := func(action, reason string) {
		ns.WithJob(jobId, action).NotifyInfof("SKIPPED — %s", reason)
		ns.Notify(tracker.NewJobProgressMessage(jobId, action, 100).
			SetJobId(jobId).SetCurrentAction(action))
		time.Sleep(50 * time.Millisecond) // nearly instant
	}

	instantStep(constants.ActionValidatingRequest, "Request validated")
	instantStep(constants.ActionCheckingLocalCatalog, "Catalog entry found")
	instantStep(constants.ActionCheckingRemoteCatalog, "Remote provider located")
	instantStep(constants.ActionDownloadingManifest, "Manifest downloaded")
	instantStep(constants.ActionCreatingDestinationFolder, "Destination folder created")

	// Skipped because the pack file is already in cache
	skip(constants.ActionDownloadingPackFile, "pack file already present in cache")
	skip(constants.ActionDecompressingPackFile, "pack file already decompressed in cache")

	// Copy from cache (main operation, 2% per 200 ms → ~20 s)
	const totalBytes int64 = 10 * 1024 * 1024 * 1024
	cacheId := "debug-skip-" + helpers.GenerateId()
	startTime := time.Now()
	ns.WithJob(jobId, constants.ActionCopyingFromCache).NotifyInfof("Copying from cache to /tmp/debug-vm.pvm")
	for pct := 0.0; pct <= 100.0; pct += 2.0 {
		ns.Notify(tracker.NewJobProgressMessage(cacheId, "Copying from cache", pct).
			SetCurrentSize(int64(pct / 100.0 * float64(totalBytes))).
			SetTotalSize(totalBytes).
			SetStartingTime(startTime).
			SetJobId(jobId).
			SetCurrentAction(constants.ActionCopyingFromCache).
			SetFilename("debug-vm.pdpack"))
		time.Sleep(200 * time.Millisecond)
	}
	ns.WithJob(jobId, constants.ActionCopyingFromCache).NotifyInfof("Copy complete")

	instantStep(constants.ActionCleaningStructure, "Flattened directory structure")
	instantStep(constants.ActionRegisteringMachine, "debug-vm registered")
	instantStep(constants.ActionRenamingMachine, "Renamed to debug-vm")
	instantStep(constants.ActionStartingMachine, "debug-vm started")

	recordId := "rec_fast_" + helpers.GenerateId()[:8]
	_, _ = jobManager.UpdateJobResultRecord(jobId, recordId, "vm_deployment_record")

	_ = jobManager.MarkJobComplete(jobId, "skipped_steps debug task finished")
	bCtx.LogInfof("[Debug/skipped_steps] Job %s completed", jobId)
}
