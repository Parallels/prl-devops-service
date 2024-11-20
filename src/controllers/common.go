package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
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
		ctx.LogErrorf("Recovered from panic: %v", err)
		ReturnApiError(ctx, w, models.NewFromErrorWithCode(fmt.Errorf("internal server error"), http.StatusInternalServerError))

		fmt.Printf("Recovered from panic: %v", err)
	}
}

func ReturnApiError(ctx basecontext.ApiContext, w http.ResponseWriter, err models.ApiErrorResponse) {
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
