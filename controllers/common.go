package controllers

import (
	"Parallels/pd-api-service/models"
	"encoding/json"
	"net/http"
)

func ReturnApiError(w http.ResponseWriter, err models.ApiErrorResponse) {
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

func ReturnApiCommonResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	data := models.ApiCommonResponse{
		Success: true,
	}
	json.NewEncoder(w).Encode(data)
}

func ReturnApiCommonResponseWithCode(w http.ResponseWriter, code int) {
	w.WriteHeader(http.StatusOK)
	data := models.ApiCommonResponse{
		Success: true,
		Code:    code,
	}
	json.NewEncoder(w).Encode(data)
}

func ReturnApiCommonResponseWithData(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	responseData := models.ApiCommonResponse{
		Success: true,
		Data:    data,
	}
	json.NewEncoder(w).Encode(responseData)
}

func ReturnApiCommonResponseWithDataAndCode(w http.ResponseWriter, data interface{}, code int) {
	w.WriteHeader(http.StatusOK)
	responseData := models.ApiCommonResponse{
		Success: true,
		Data:    data,
		Code:    code,
	}
	json.NewEncoder(w).Encode(responseData)
}
