package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog"
	catalog_models "github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerCatalogManifestHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfo("Registering version %s Catalog Manifests handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog").
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithHandler(GetCatalogManifestsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog/{catalogId}").
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithHandler(GetCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}").
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithHandler(GetCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/catalog").
		WithRequiredClaim(constants.CREATE_CATALOG_MANIFEST_CLAIM).
		WithHandler(CreateCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/{catalogId}").
		WithRequiredClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM).
		WithHandler(DeleteCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}").
		WithRequiredClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM).
		WithHandler(DeleteCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/download").
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithHandler(DownloadCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/taint").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(TaintCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/untaint").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(UnTaintCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/revoke").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(RevokeCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/catalog/push").
		WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM).
		WithHandler(PushCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog/pull").
		WithRequiredClaim(constants.PULL_CATALOG_MANIFEST_CLAIM).
		WithHandler(PullCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog/import").
		WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM).
		WithHandler(ImportCatalogManifestHandler()).
		Register()
}

// @Summary		Gets all the remote catalogs
// @Description	This endpoint returns all the remote catalogs
// @Tags			Catalogs
// @Produce		json
// @Success		200	{object}	[]map[string][]models.CatalogManifest
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog [get]
func GetCatalogManifestsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
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

		result := make([]map[string][]models.CatalogManifest, 0)
		for _, manifest := range manifestsDto {
			var resultManifest map[string][]models.CatalogManifest
			for _, r := range result {
				if _, ok := r[manifest.CatalogId]; ok {
					resultManifest = r
					break
				}
			}
			if resultManifest == nil {
				resultManifest := make(map[string][]models.CatalogManifest)
				resultManifest[manifest.CatalogId] = append(resultManifest[manifest.CatalogId], mappers.DtoCatalogManifestToApi(manifest))
				result = append(result, resultManifest)
			} else {
				if _, ok := resultManifest[manifest.CatalogId]; !ok {
					resultManifest[manifest.CatalogId] = []models.CatalogManifest{}
				}
				resultManifest[manifest.CatalogId] = append(resultManifest[manifest.CatalogId], mappers.DtoCatalogManifestToApi(manifest))
			}
		}

		// responseManifests := mappers.DtoCatalogManifestsToApi(manifestsDto)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Manifests returned: %v", len(result))
	}
}

// @Summary		Gets all the remote catalogs
// @Description	This endpoint returns all the remote catalogs
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path		string	true	"Catalog ID"
// @Success		200			{object}	[]models.CatalogManifest
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId} [get]
func GetCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]

		manifest, err := dbService.GetCatalogManifestsByCatalogId(ctx, catalogId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}
		if len(manifest) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.CatalogManifest, 0)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Manifests returned: %v", len(response))
			return
		}

		resultData := mappers.DtoCatalogManifestsToApi(manifest)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifests returned: %v", len(resultData))
	}
}

// @Summary		Gets a catalog manifest version
// @Description	This endpoint returns a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path		string	true	"Catalog ID"
// @Param			version		path		string	true	"Version"
// @Success		200			{object}	models.CatalogManifest
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version} [get]
func GetCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*manifest)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest: %v", resultData.ID)
	}
}

// @Summary		Downloads a catalog manifest version
// @Description	This endpoint downloads a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path		string	true	"Catalog ID"
// @Param			version		path		string	true	"Version"
// @Success		200			{object}	models.CatalogManifest
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/download [get]
func DownloadCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if manifest.Tainted {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Manifest is tainted",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if manifest.Revoked {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Manifest is revoked",
				Code:    http.StatusForbidden,
			})
			return
		}

		if err := dbService.UpdateCatalogManifestDownloadCount(ctx, manifest.CatalogId, manifest.Version); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*manifest)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest: %v", resultData.ID)
	}
}

// @Summary		Taints a catalog manifest version
// @Description	This endpoint Taints a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path		string	true	"Catalog ID"
// @Param			version		path		string	true	"Version"
// @Success		200			{object}	models.CatalogManifest
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/taint [patch]
func TaintCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if manifest.Tainted {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Manifest is already tainted",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if manifest.Revoked {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Manifest is revoked",
				Code:    http.StatusForbidden,
			})
			return
		}

		result, err := dbService.TaintCatalogManifestVersion(ctx, manifest.CatalogId, manifest.Version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*result)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest tainted: %v", resultData.ID)
	}
}

// @Summary		UnTaints a catalog manifest version
// @Description	This endpoint UnTaints a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path		string	true	"Catalog ID"
// @Param			version		path		string	true	"Version"
// @Success		200			{object}	models.CatalogManifest
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/untaint [patch]
func UnTaintCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if !manifest.Tainted {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Manifest is not tainted",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if manifest.Revoked {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Manifest is revoked",
				Code:    http.StatusForbidden,
			})
			return
		}

		result, err := dbService.UnTaintCatalogManifestVersion(ctx, manifest.CatalogId, manifest.Version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*result)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest untainted: %v", resultData.ID)
	}
}

// @Summary		UnTaints a catalog manifest version
// @Description	This endpoint UnTaints a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path		string	true	"Catalog ID"
// @Param			version		path		string	true	"Version"
// @Success		200			{object}	models.CatalogManifest
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/revoke [patch]
func RevokeCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if manifest.Revoked {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Manifest is already revoked",
				Code:    http.StatusForbidden,
			})
			return
		}

		result, err := dbService.RevokeCatalogManifestVersion(ctx, manifest.CatalogId, manifest.Version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*result)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest untainted: %v", resultData.ID)
	}
}

func CreateCatalogManifestHandler() restapi.ControllerHandler {
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

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		ctx.LogInfo("Creating manifest %v", request.Name)
		dto := mappers.CatalogManifestToDto(request)
		result, err := dbService.CreateCatalogManifest(ctx, dto)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*result)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resultData)
		ctx.LogInfo("Manifest returned: %v", resultData.ID)
	}
}

// @Summary		Deletes a catalog manifest and all its versions
// @Description	This endpoint deletes a catalog manifest and all its versions
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path	string	true	"Catalog ID"
// @Success		200
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId} [delete]
func DeleteCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]

		cleanRemote := http_helper.GetHttpRequestStrValue(r, constants.DELETE_REMOTE_MANIFEST_QUERY)

		manifest := catalog.NewManifestService(ctx)
		if cleanRemote == "true" {
			ctx.LogInfo("Deleting remote manifest %v", catalogId)
			err = manifest.Delete(ctx, catalogId, "")
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		}

		err = dbService.DeleteCatalogManifest(ctx, catalogId)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Catalog manifest deleted successfully")
	}
}

// @Summary		Deletes a catalog manifest version
// @Description	This endpoint deletes a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path	string	true	"Catalog ID"
// @Param			version		path	string	true	"Version"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version} [delete]
func DeleteCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]

		cleanRemote := http_helper.GetHttpRequestStrValue(r, constants.DELETE_REMOTE_MANIFEST_QUERY)

		manifest := catalog.NewManifestService(ctx)
		if cleanRemote == "true" {
			ctx.LogInfo("Deleting remote manifest %v", catalogId)
			err = manifest.Delete(ctx, catalogId, version)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		}

		err = dbService.DeleteCatalogManifestVersion(ctx, catalogId, version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Catalog manifest deleted successfully")
	}
}

// @Summary		Pushes a catalog manifest to the catalog inventory
// @Description	This endpoint pushes a catalog manifest to the catalog inventory
// @Tags			Catalogs
// @Produce		json
// @Param			pushRequest	body		catalog_models.PushCatalogManifestRequest	true	"Push request"
// @Success		200			{object}	models.CatalogManifest
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/push [post]
func PushCatalogManifestHandler() restapi.ControllerHandler {
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

// @Summary		Pull a remote catalog manifest
// @Description	This endpoint pulls a remote catalog manifest
// @Tags			Catalogs
// @Produce		json
// @Param			pullRequest	body		catalog_models.PullCatalogManifestRequest	true	"Pull request"
// @Success		200			{object}	models.PullCatalogManifestResponse
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/pull [put]
func PullCatalogManifestHandler() restapi.ControllerHandler {
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

// @Summary		Imports a remote catalog manifest metadata into the catalog inventory
// @Description	This endpoint imports a remote catalog manifest metadata into the catalog inventory
// @Tags			Catalogs
// @Produce		json
// @Param			importRequest	body		catalog_models.ImportCatalogManifestRequest	true	"Pull request"
// @Success		200				{object}	models.ImportCatalogManifestResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/import [put]
func ImportCatalogManifestHandler() restapi.ControllerHandler {
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
