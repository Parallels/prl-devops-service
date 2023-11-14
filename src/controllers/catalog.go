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

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerCatalogManifestHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfo("Registering version %s Catalog Manifests handlers", version)
	catalogManifestsController := restapi.NewController()
	catalogManifestsController.WithMethod(restapi.GET)
	catalogManifestsController.WithVersion(version)
	catalogManifestsController.WithPath("/catalog")
	catalogManifestsController.WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM)
	catalogManifestsController.WithHandler(GetCatalogManifestsHandler()).Register()

	getCatalogManifestController := restapi.NewController()
	getCatalogManifestController.WithMethod(restapi.GET)
	getCatalogManifestController.WithVersion(version)
	getCatalogManifestController.WithPath("/catalog/{catalogId}")
	getCatalogManifestController.WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM)
	getCatalogManifestController.WithHandler(GetCatalogManifestHandler()).Register()

	getCatalogManifestVersionController := restapi.NewController()
	getCatalogManifestVersionController.WithMethod(restapi.GET)
	getCatalogManifestVersionController.WithVersion(version)
	getCatalogManifestVersionController.WithPath("/catalog/{catalogId}/{version}")
	getCatalogManifestVersionController.WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM)
	getCatalogManifestVersionController.WithHandler(GetCatalogManifestVersionHandler()).Register()

	createCatalogManifestController := restapi.NewController()
	createCatalogManifestController.WithMethod(restapi.POST)
	createCatalogManifestController.WithVersion(version)
	createCatalogManifestController.WithPath("/catalog")
	createCatalogManifestController.WithRequiredClaim(constants.CREATE_CATALOG_MANIFEST_CLAIM)
	createCatalogManifestController.WithHandler(CreateCatalogManifestHandler()).Register()

	deleteCatalogManifestByIdController := restapi.NewController()
	deleteCatalogManifestByIdController.WithMethod(restapi.DELETE)
	deleteCatalogManifestByIdController.WithVersion(version)
	deleteCatalogManifestByIdController.WithPath("/catalog/{catalogId}")
	deleteCatalogManifestByIdController.WithRequiredClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM)
	deleteCatalogManifestByIdController.WithHandler(DeleteCatalogManifestHandler()).Register()

	deleteCatalogManifestVersionController := restapi.NewController()
	deleteCatalogManifestVersionController.WithMethod(restapi.DELETE)
	deleteCatalogManifestVersionController.WithVersion(version)
	deleteCatalogManifestVersionController.WithPath("/catalog/{catalogId}/{version}")
	deleteCatalogManifestVersionController.WithRequiredClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM)
	deleteCatalogManifestVersionController.WithHandler(DeleteCatalogManifestVersionHandler()).Register()

	downloadCatalogManifestController := restapi.NewController()
	downloadCatalogManifestController.WithMethod(restapi.GET)
	downloadCatalogManifestController.WithVersion(version)
	downloadCatalogManifestController.WithPath("/catalog/{catalogId}/{version}/download")
	downloadCatalogManifestController.WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM)
	downloadCatalogManifestController.WithHandler(DownloadCatalogManifestVersionHandler()).Register()

	taintCatalogManifestController := restapi.NewController()
	taintCatalogManifestController.WithMethod(restapi.PATCH)
	taintCatalogManifestController.WithVersion(version)
	taintCatalogManifestController.WithPath("/catalog/{catalogId}/{version}/taint")
	taintCatalogManifestController.WithRequiredRole(constants.SUPER_USER_ROLE)
	taintCatalogManifestController.WithHandler(TaintCatalogManifestVersionHandler()).Register()

	unTaintCatalogManifestController := restapi.NewController()
	unTaintCatalogManifestController.WithMethod(restapi.PATCH)
	unTaintCatalogManifestController.WithVersion(version)
	unTaintCatalogManifestController.WithPath("/catalog/{catalogId}/{version}/untaint")
	unTaintCatalogManifestController.WithRequiredRole(constants.SUPER_USER_ROLE)
	unTaintCatalogManifestController.WithHandler(UnTaintCatalogManifestVersionHandler()).Register()

	revokeCatalogManifestController := restapi.NewController()
	revokeCatalogManifestController.WithMethod(restapi.PATCH)
	revokeCatalogManifestController.WithVersion(version)
	revokeCatalogManifestController.WithPath("/catalog/{catalogId}/{version}/revoke")
	revokeCatalogManifestController.WithRequiredRole(constants.SUPER_USER_ROLE)
	revokeCatalogManifestController.WithHandler(RevokeCatalogManifestVersionHandler()).Register()

	pushCatalogManifestController := restapi.NewController()
	pushCatalogManifestController.WithMethod(restapi.POST)
	pushCatalogManifestController.WithVersion(version)
	pushCatalogManifestController.WithPath("/catalog/push")
	pushCatalogManifestController.WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM)
	pushCatalogManifestController.WithHandler(PushCatalogManifestHandler()).Register()

	pullCatalogManifestController := restapi.NewController()
	pullCatalogManifestController.WithMethod(restapi.PUT)
	pullCatalogManifestController.WithVersion(version)
	pullCatalogManifestController.WithPath("/catalog/pull")
	pullCatalogManifestController.WithRequiredClaim(constants.PULL_CATALOG_MANIFEST_CLAIM)
	pullCatalogManifestController.WithHandler(PullCatalogManifestHandler()).Register()

	importCatalogManifestController := restapi.NewController()
	importCatalogManifestController.WithMethod(restapi.PUT)
	importCatalogManifestController.WithVersion(version)
	importCatalogManifestController.WithPath("/catalog/import")
	importCatalogManifestController.WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM)
	importCatalogManifestController.WithHandler(ImportCatalogManifestHandler()).Register()
}

//	@Summary		Gets all the remote catalogs
//	@Description	This endpoint returns all the remote catalogs
//	@Tags			Catalogs
//	@Produce		json
//	@Success		200	{object}	[]map[string][]models.CatalogManifest
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog [get]
func GetCatalogManifestsHandler() restapi.ControllerHandler {
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

//	@Summary		Gets all the remote catalogs
//	@Description	This endpoint returns all the remote catalogs
//	@Tags			Catalogs
//	@Produce		json
//	@Param			catalogId	path		string	true	"Catalog ID"
//	@Success		200			{object}	[]models.CatalogManifest
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/{catalogId} [get]
func GetCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
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

//	@Summary		Gets a catalog manifest version
//	@Description	This endpoint returns a catalog manifest version
//	@Tags			Catalogs
//	@Produce		json
//	@Param			catalogId	path		string	true	"Catalog ID"
//	@Param			version		path		string	true	"Version"
//	@Success		200			{object}	models.CatalogManifest
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/{catalogId}/{version} [get]
func GetCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
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

//	@Summary		Downloads a catalog manifest version
//	@Description	This endpoint downloads a catalog manifest version
//	@Tags			Catalogs
//	@Produce		json
//	@Param			catalogId	path		string	true	"Catalog ID"
//	@Param			version		path		string	true	"Version"
//	@Success		200			{object}	models.CatalogManifest
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/{catalogId}/{version}/download [get]
func DownloadCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
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

//	@Summary		Taints a catalog manifest version
//	@Description	This endpoint Taints a catalog manifest version
//	@Tags			Catalogs
//	@Produce		json
//	@Param			catalogId	path		string	true	"Catalog ID"
//	@Param			version		path		string	true	"Version"
//	@Success		200			{object}	models.CatalogManifest
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/{catalogId}/{version}/taint [patch]
func TaintCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
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

//	@Summary		UnTaints a catalog manifest version
//	@Description	This endpoint UnTaints a catalog manifest version
//	@Tags			Catalogs
//	@Produce		json
//	@Param			catalogId	path		string	true	"Catalog ID"
//	@Param			version		path		string	true	"Version"
//	@Success		200			{object}	models.CatalogManifest
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/{catalogId}/{version}/untaint [patch]
func UnTaintCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
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

//	@Summary		UnTaints a catalog manifest version
//	@Description	This endpoint UnTaints a catalog manifest version
//	@Tags			Catalogs
//	@Produce		json
//	@Param			catalogId	path		string	true	"Catalog ID"
//	@Param			version		path		string	true	"Version"
//	@Success		200			{object}	models.CatalogManifest
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/{catalogId}/{version}/revoke [patch]
func RevokeCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
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

		dbService, err := GetDatabaseService(ctx)
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

//	@Summary		Deletes a catalog manifest and all its versions
//	@Description	This endpoint deletes a catalog manifest and all its versions
//	@Tags			Catalogs
//	@Produce		json
//	@Param			catalogId	path	string	true	"Catalog ID"
//	@Success		200
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/{catalogId} [delete]
func DeleteCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
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

//	@Summary		Deletes a catalog manifest version
//	@Description	This endpoint deletes a catalog manifest version
//	@Tags			Catalogs
//	@Produce		json
//	@Param			catalogId	path	string	true	"Catalog ID"
//	@Param			version		path	string	true	"Version"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/{catalogId}/{version} [delete]
func DeleteCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
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

//	@Summary		Pushes a catalog manifest to the catalog inventory
//	@Description	This endpoint pushes a catalog manifest to the catalog inventory
//	@Tags			Catalogs
//	@Produce		json
//	@Param			pushRequest	body		catalog_models.PushCatalogManifestRequest	true	"Push request"
//	@Success		200			{object}	models.CatalogManifest
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/push [post]
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

//	@Summary		Pull a remote catalog manifest
//	@Description	This endpoint pulls a remote catalog manifest
//	@Tags			Catalogs
//	@Produce		json
//	@Param			pullRequest	body		catalog_models.PullCatalogManifestRequest	true	"Pull request"
//	@Success		200			{object}	models.PullCatalogManifestResponse
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/pull [put]
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

//	@Summary		Imports a remote catalog manifest metadata into the catalog inventory
//	@Description	This endpoint imports a remote catalog manifest metadata into the catalog inventory
//	@Tags			Catalogs
//	@Produce		json
//	@Param			importRequest	body		catalog_models.ImportCatalogManifestRequest	true	"Pull request"
//	@Success		200				{object}	models.ImportCatalogManifestResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/catalog/import [put]
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
