package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
)

func registerPerformanceHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s ApiKeys handlers", version)
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).WithPath("/performance/db").
		WithHandler(PerformDbTestHandler()).
		Register()
}

func PerformDbTestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request models.PerformanceRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if request.TestCount == 0 {
			request.TestCount = 1
		}

		if request.ConsecutiveCalls == 0 {
			request.ConsecutiveCalls = 1
		}

		if request.TimeBetweenCalls == 0 {
			request.TimeBetweenCalls = 1
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		for i := 0; i < request.TestCount; i++ {
			ctx.LogInfof("This is a test log")
			for j := 0; j < request.ConsecutiveCalls; j++ {
				go dbService.SaveNow(ctx)
				if request.TimeBetweenConsecutiveCalls > 0 {
					time.Sleep(time.Duration(request.TimeBetweenConsecutiveCalls) * time.Millisecond)
				}
			}

			if request.TimeBetweenCalls > 0 {
				time.Sleep(time.Duration(request.TimeBetweenCalls) * time.Millisecond)
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode("ok")
		ctx.LogInfof("Performance run successfully")
	}
}
