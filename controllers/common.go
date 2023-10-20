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
