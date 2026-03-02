package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
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
		WithHandler(GetJobsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/jobs/{id}").
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_OWN_CLAIM).
		WithHandler(GetJobHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/jobs/cleanup").
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_OWN_CLAIM).
		WithHandler(CleanupJobsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/jobs/debug").
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.JOBS_MANAGER_LIST_OWN_CLAIM).
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

		jobType := request.JobType
		jobOperation := request.JobOperation

		job, err := jobManager.CreateNewJob(userContext.ID, jobType, jobOperation, action)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		// Fire off the background task
		go func(jobId string) {
			bCtx := basecontext.NewRootBaseContext()
			bCtx.LogInfof("[Debug Job] Starting background job: %s", jobId)

			for i := 1; i <= 10; i++ {
				time.Sleep(2 * time.Second)
				_, _ = jobManager.UpdateJobProgress(jobId, action, i*10, constants.JobStateRunning)
				bCtx.LogInfof("[Debug Job] Progress updated for job %s: %d%%", jobId, i*10)
			}

			_ = jobManager.MarkJobComplete(jobId, "Debug task finished successfully")
			bCtx.LogInfof("[Debug Job] Job finished: %s", jobId)
		}(job.ID)

		response := mappers.MapJobToApiJob(*job)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
	}
}
