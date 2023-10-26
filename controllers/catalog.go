package controllers

import (
	"Parallels/pd-api-service/catalog"
	catalog_models "Parallels/pd-api-service/catalog/models"
	"Parallels/pd-api-service/mappers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/service_provider"
	"encoding/json"
	"net/http"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

// GetUsers is a public function that returns all users
func GetCatalogManifestsController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		manifests, err := dbService.GetCatalogManifests(r.Context())
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		responseData := make([]models.CatalogVirtualMachineManifest, 0)
		for _, manifest := range manifests {
			responseData = append(responseData, mappers.DtoCatalogManifestToApi(manifest))
		}

		jsonData, err := json.Marshal(responseData)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func GetCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Query the users table for the manifest with the given ID
		manifest, err := dbService.GetCatalogManifest(r.Context(), id)

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}
		if manifest == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Manifest not found",
				Code:    http.StatusNotFound,
			})
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*manifest)

		// Marshal the user struct to JSON
		jsonData, err := json.Marshal(resultData)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func PushCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request catalog_models.PushCatalogManifestRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := catalog.NewManifestService(r.Context())
		resultManifest, err := manifest.Push(&request)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(mappers.DtoCatalogManifestFromBase(*resultManifest))
		resultData.ID = resultManifest.ID

		// Marshal the user struct to JSON
		jsonData, err := json.Marshal(resultData)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func PullCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request catalog_models.PullCatalogManifestRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := catalog.NewManifestService(r.Context())
		resultManifest, err := manifest.Pull(&request)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		resultData := mappers.BasePullCatalogManifestResponseToApi(*resultManifest)
		resultData.ID = resultManifest.ID

		// Marshal the user struct to JSON
		jsonData, err := json.Marshal(resultData)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}
