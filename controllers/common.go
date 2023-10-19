package controllers

import (
	"Parallels/pd-api-service/models"
	"encoding/json"
	"net/http"
)

func ReturnApiError(w http.ResponseWriter, err models.ApiErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(err)
}
