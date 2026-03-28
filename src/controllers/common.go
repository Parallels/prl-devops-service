package controllers

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

// getEffectiveCallerID returns the user ID to associate with a job.
// For requests authenticated via API key (IsMicroService == true), such as
// orchestrator-to-host calls, there is no user account on the target, so
// "system" is used as a fallback owner. Returns ("", false) when the request
// is not authenticated at all.
func getEffectiveCallerID(ctx *basecontext.BaseContext) (string, bool) {
	if user := ctx.GetUser(); user != nil {
		return user.ID, true
	}
	authCtx := ctx.GetAuthorizationContext()
	if authCtx != nil && authCtx.IsMicroService {
    // Getting the user based on the user we are logged in as
    // in the service using the system service
    sysSvc := system.Get()
    sysUser, err := sysSvc.GetCurrentUser(ctx)
    currentUser := "root"
    if err == nil {
      currentUser = sysUser
    }
    return currentUser, true
  }

	return "", false
}

func GetFilterHeader(r *http.Request) string {
	return r.Header.Get("X-Filter")
}

func GetBaseContext(r *http.Request) *basecontext.BaseContext {
	ctx := basecontext.NewBaseContextFromRequest(r)

	return ctx
}

func Recover(ctx basecontext.ApiContext, r *http.Request, w http.ResponseWriter) {
	if err := recover(); err != nil {
		ctx.LogErrorf("Recovered from panic: %v\n%v", err, string(debug.Stack()))
		sysErr := errors.NewWithCodef(http.StatusInternalServerError, "internal server error")
		sysErr.Stack = make([]errors.StackItem, 0)
		sysErr.AddStackMessage(string(debug.Stack()))
		ReturnApiError(ctx, w, models.NewFromErrorWithCode(sysErr, http.StatusInternalServerError))
	}
}

func ReturnApiError(ctx basecontext.ApiContext, w http.ResponseWriter, err models.ApiErrorResponse) {
	ctx.LogErrorf("Error: %v", err.Message)
	w.WriteHeader(err.Code)

	_ = json.NewEncoder(w).Encode(err)
}

func ReturnApiErrorWithDiagnostics(ctx basecontext.ApiContext, w http.ResponseWriter, err models.ApiErrorDiagnosticsResponse) {
	ctx.LogErrorf("Error: %v", err.Message)
	w.WriteHeader(err.Code)

	_ = json.NewEncoder(w).Encode(err)
}

func ReturnApiCommonResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	data := models.ApiCommonResponse{
		Success: true,
	}
	_ = json.NewEncoder(w).Encode(data)
}

func ReturnApiCommonResponseWithCode(w http.ResponseWriter, code int) {
	w.WriteHeader(http.StatusOK)
	data := models.ApiCommonResponse{
		Success: true,
		Code:    code,
	}
	_ = json.NewEncoder(w).Encode(data)
}

func ReturnApiCommonResponseWithData(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	responseData := models.ApiCommonResponse{
		Success: true,
		Data:    data,
	}
	_ = json.NewEncoder(w).Encode(responseData)
}

// emitAuthEvent fires an EventTypeAuth event in a goroutine. It is a no-op
// when the event emitter is unavailable, so callers never need to guard it.
func emitAuthEvent(message string, body interface{}) {
	emitter := serviceprovider.GetEventEmitter()
	if emitter == nil || !emitter.IsRunning() {
		return
	}
	msg := models.NewEventMessage(constants.EventTypeAuth, message, body)
	go func() { _ = emitter.Broadcast(msg) }()
}

func ReturnApiCommonResponseWithDataAndCode(w http.ResponseWriter, data interface{}, code int) {
	w.WriteHeader(http.StatusOK)
	responseData := models.ApiCommonResponse{
		Success: true,
		Data:    data,
		Code:    code,
	}
	_ = json.NewEncoder(w).Encode(responseData)
}
