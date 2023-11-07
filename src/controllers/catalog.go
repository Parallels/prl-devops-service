package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/catalog"
	catalog_models "github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

// GetUsers is a public function that returns all users
func GetCatalogManifestsController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		manifestsDto, err := dbService.GetCatalogManifests(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(manifestsDto) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.CatalogManifest, 0)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Manifests returned: %v", len(response))
			return
		}

		responseManifests := mappers.DtoCatalogManifestsToApi(manifestsDto)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseManifests)
		ctx.LogInfo("Manifests returned: %v", len(responseManifests))
	}
}

func GetCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		manifest, err := dbService.GetCatalogManifest(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*manifest)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest returned: %v", resultData.ID)
	}
}

func CreateCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request catalog_models.VirtualMachineCatalogManifest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		exists, _ := dbService.GetCatalogManifest(ctx, id)
		if exists != nil {
			dto := mappers.CatalogManifestToDto(request)
			if err := dbService.UpdateCatalogManifest(ctx, dto); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		} else {
			ctx.LogInfo("Creating manifest %v", request.Name)
			dto := mappers.CatalogManifestToDto(request)
			if err := dbService.CreateCatalogManifest(ctx, dto); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		}

		resultData := mappers.DtoCatalogManifestToApi(mappers.CatalogManifestToDto(request))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest returned: %v", resultData.ID)
	}
}

func DeleteCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		cleanRemote := http_helper.GetHttpRequestStrValue(r, constants.DELETE_REMOTE_MANIFEST_QUERY)

		manifest := catalog.NewManifestService(ctx)
		if cleanRemote == "true" {
			ctx.LogInfo("Deleting remote manifest %v", id)
			err = manifest.Delete(ctx, id)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		}

		err = dbService.DeleteCatalogManifest(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Catalog manifest deleted successfully")
	}
}

func PushCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request catalog_models.PushCatalogManifestRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := catalog.NewManifestService(ctx)
		resultManifest := manifest.Push(ctx, &request)
		if resultManifest.HasErrors() {
			errorMessage := "Error pushing manifest: \n"
			for _, err := range resultManifest.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(mappers.CatalogManifestToDto(*resultManifest))
		resultData.ID = resultManifest.ID

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest pushed: %v", resultData.ID)
	}
}

func PullCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request catalog_models.PullCatalogManifestRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := catalog.NewManifestService(ctx)
		resultManifest := manifest.Pull(ctx, &request)
		if resultManifest.HasErrors() {
			errorMessage := "Error pulling manifest: \n"
			for _, err := range resultManifest.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}

		resultData := mappers.BasePullCatalogManifestResponseToApi(*resultManifest)
		resultData.ID = resultManifest.ID

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest pulled: %v", resultData.ID)
	}
}

func ImportCatalogManifestController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request catalog_models.ImportCatalogManifestRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := catalog.NewManifestService(ctx)
		resultManifest := manifest.Import(ctx, &request)
		if resultManifest.HasErrors() {
			errorMessage := "Error importing manifest: \n"
			for _, err := range resultManifest.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}

		resultData := mappers.BaseImportCatalogManifestResponseToApi(*resultManifest)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest imported: %v", resultData.ID)
	}
}
