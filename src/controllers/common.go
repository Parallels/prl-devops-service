package controllers

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/models"
)

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

func ReturnApiCommonResponseWithDataAndCode(w http.ResponseWriter, data interface{}, code int) {
	w.WriteHeader(http.StatusOK)
	responseData := models.ApiCommonResponse{
		Success: true,
		Data:    data,
		Code:    code,
	}
	_ = json.NewEncoder(w).Encode(responseData)
}
