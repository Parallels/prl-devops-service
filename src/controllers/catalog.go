package controllers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	"github.com/Parallels/prl-devops-service/catalog/cacheservice"
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/telemetry"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerCatalogManifestHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Catalog Manifests handlers", version)

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog/cache").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(GetCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/cache").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/cache/{catalogId}").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteCatalogCacheItemHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/cache/{catalogId}/{version}").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteCatalogCacheItemVersionHandler()).
		Register()

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
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}").
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithHandler(GetCatalogManifestVersionArchitectureHandler()).
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
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}").
		WithRequiredClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM).
		WithHandler(DeleteCatalogManifestVersionArchitectureHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/download").
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithHandler(DownloadCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/taint").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(TaintCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/untaint").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(UnTaintCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/revoke").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(RevokeCatalogManifestVersionHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/claims").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(AddClaimsToCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/claims").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(RemoveClaimsToCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/roles").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(AddRolesToCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/roles").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(RemoveRolesToCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/tags").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(AddTagsToCatalogManifestHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/connection").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(UpdateCatalogManifestProviderHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/{catalogId}/{version}/{architecture}/tags").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(RemoveTagsToCatalogManifestHandler()).
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

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog/import-vm").
		WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM).
		WithHandler(ImportVmHandler()).
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
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		manifestsDto, err := dbService.GetCatalogManifests(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(manifestsDto) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.CatalogManifest, 0)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Manifests returned: %v", len(response))
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
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Manifests returned: %v", len(result))
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
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

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
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Manifests returned: %v", len(response))
			return
		}

		resultData := mappers.DtoCatalogManifestsToApi(manifest)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifests returned: %v", len(resultData))
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
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]

		manifests, err := dbService.GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestsToApi(manifests)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifests returned: %v", len(resultData))
	}
}

// @Summary		Gets a catalog manifest version architecture
// @Description	This endpoint returns a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string	true	"Catalog ID"
// @Param			version			path		string	true	"Version"
// @Param			architecture	path		string	true	"Architecture"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture} [get]
func GetCatalogManifestVersionArchitectureHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*manifest)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest returned with id %v", resultData.ID)
	}
}

// @Summary		Downloads a catalog manifest version
// @Description	This endpoint downloads a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string	true	"Catalog ID"
// @Param			version			path		string	true	"Version"
// @Param			architecture	path		string	true	"Architecture"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/download [get]
func DownloadCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
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
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest: %v", resultData.ID)
	}
}

// @Summary		Taints a catalog manifest version
// @Description	This endpoint Taints a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string	true	"Catalog ID"
// @Param			version			path		string	true	"Version"
// @Param			architecture	path		string	true	"Architecture"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/taint [patch]
func TaintCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
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
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest tainted: %v", resultData.ID)
	}
}

// @Summary		UnTaints a catalog manifest version
// @Description	This endpoint UnTaints a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string	true	"Catalog ID"
// @Param			version			path		string	true	"Version"
// @Param			architecture	path		string	true	"Architecture"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/untaint [patch]
func UnTaintCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
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
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest untainted: %v", resultData.ID)
	}
}

// @Summary		UnTaints a catalog manifest version
// @Description	This endpoint UnTaints a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string	true	"Catalog ID"
// @Param			version			path		string	true	"Version"
// @Param			architecture	path		string	true	"Architecture"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/revoke [patch]
func RevokeCatalogManifestVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
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
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest untainted: %v", resultData.ID)
	}
}

// @Summary		Adds claims to a catalog manifest version
// @Description	This endpoint adds claims to a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string										true	"Catalog ID"
// @Param			version			path		string										true	"Version"
// @Param			architecture	path		string										true	"Architecture"
// @Param			request			body		models.VirtualMachineCatalogManifestPatch	true	"Body"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/claims [patch]
func AddClaimsToCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request catalog_models.VirtualMachineCatalogManifestPatch
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
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

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if request.RequiredClaims != nil && len(request.RequiredClaims) > 0 {
			if err := dbService.AddCatalogManifestRequiredClaims(ctx, manifest.ID, request.RequiredClaims...); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		} else {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "No claims provided",
				Code:    http.StatusBadRequest,
			})
			return
		}

		newManifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		catalogSvc := catalog.NewManifestService(ctx)
		catalogRequest := mappers.DtoCatalogManifestToBase(*newManifest)
		catalogRequest.CleanupRequest = cleanupservice.NewCleanupService()
		catalogRequest.Errors = []error{}

		resultOp := catalogSvc.PushMetadata(&catalogRequest)
		if resultOp.HasErrors() {
			errorMessage := "Error pushing manifest: \n"
			for _, err := range resultOp.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(newManifest)
		ctx.LogInfof("Manifest Claims Updated: %v", newManifest.ID)
	}
}

// @Summary		Removes claims from a catalog manifest version
// @Description	This endpoint removes claims from a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string										true	"Catalog ID"
// @Param			version			path		string										true	"Version"
// @Param			architecture	path		string										true	"Architecture"
// @Param			request			body		models.VirtualMachineCatalogManifestPatch	true	"Body"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/claims [delete]
func RemoveClaimsToCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request catalog_models.VirtualMachineCatalogManifestPatch
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
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

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if request.RequiredClaims != nil && len(request.RequiredClaims) > 0 {
			if err := dbService.RemoveCatalogManifestRequiredClaims(ctx, manifest.ID, request.RequiredClaims...); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		} else {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "No claims provided",
				Code:    http.StatusBadRequest,
			})
			return
		}

		newManifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		catalogSvc := catalog.NewManifestService(ctx)
		catalogRequest := mappers.DtoCatalogManifestToBase(*newManifest)
		catalogRequest.CleanupRequest = cleanupservice.NewCleanupService()
		catalogRequest.Errors = []error{}

		resultOp := catalogSvc.PushMetadata(&catalogRequest)
		if resultOp.HasErrors() {
			errorMessage := "Error pushing manifest: \n"
			for _, err := range resultOp.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(newManifest)
		ctx.LogInfof("Manifest Claims Updated: %v", newManifest.ID)
	}
}

// @Summary		Adds roles to a catalog manifest version
// @Description	This endpoint adds roles to a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string										true	"Catalog ID"
// @Param			version			path		string										true	"Version"
// @Param			architecture	path		string										true	"Architecture"
// @Param			request			body		models.VirtualMachineCatalogManifestPatch	true	"Body"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/roles [patch]
func AddRolesToCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request catalog_models.VirtualMachineCatalogManifestPatch
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
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

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if request.RequiredRoles != nil && len(request.RequiredRoles) > 0 {
			if err := dbService.AddCatalogManifestRequiredRoles(ctx, manifest.ID, request.RequiredRoles...); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		} else {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "No roles provided",
				Code:    http.StatusBadRequest,
			})
			return
		}

		newManifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		catalogSvc := catalog.NewManifestService(ctx)
		catalogRequest := mappers.DtoCatalogManifestToBase(*newManifest)
		catalogRequest.CleanupRequest = cleanupservice.NewCleanupService()
		catalogRequest.Errors = []error{}

		resultOp := catalogSvc.PushMetadata(&catalogRequest)
		if resultOp.HasErrors() {
			errorMessage := "Error pushing manifest: \n"
			for _, err := range resultOp.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(newManifest)
		ctx.LogInfof("Manifest Roles Updated: %v", newManifest.ID)
	}
}

// @Summary		Removes roles from a catalog manifest version
// @Description	This endpoint removes roles from a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string										true	"Catalog ID"
// @Param			version			path		string										true	"Version"
// @Param			architecture	path		string										true	"Architecture"
// @Param			request			body		models.VirtualMachineCatalogManifestPatch	true	"Body"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/roles [delete]
func RemoveRolesToCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request catalog_models.VirtualMachineCatalogManifestPatch
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
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

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if request.RequiredRoles != nil && len(request.RequiredRoles) > 0 {
			if err := dbService.RemoveCatalogManifestRequiredRoles(ctx, manifest.ID, request.RequiredRoles...); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		} else {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "No roles provided",
				Code:    http.StatusBadRequest,
			})
			return
		}

		newManifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		catalogSvc := catalog.NewManifestService(ctx)
		catalogRequest := mappers.DtoCatalogManifestToBase(*newManifest)
		catalogRequest.CleanupRequest = cleanupservice.NewCleanupService()
		catalogRequest.Errors = []error{}

		resultOp := catalogSvc.PushMetadata(&catalogRequest)
		if resultOp.HasErrors() {
			errorMessage := "Error pushing manifest: \n"
			for _, err := range resultOp.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(newManifest)
		ctx.LogInfof("Manifest Claims Updated: %v", newManifest.ID)
	}
}

// @Summary		Adds tags to a catalog manifest version
// @Description	This endpoint adds tags to a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string										true	"Catalog ID"
// @Param			version			path		string										true	"Version"
// @Param			architecture	path		string										true	"Architecture"
// @Param			request			body		models.VirtualMachineCatalogManifestPatch	true	"Body"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/tags [patch]
func AddTagsToCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request catalog_models.VirtualMachineCatalogManifestPatch
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
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

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if request.Tags != nil && len(request.Tags) > 0 {
			if err := dbService.AddCatalogManifestTags(ctx, manifest.ID, request.Tags...); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		} else {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "No tags provided",
				Code:    http.StatusBadRequest,
			})
			return
		}

		newManifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		catalogSvc := catalog.NewManifestService(ctx)
		catalogRequest := mappers.DtoCatalogManifestToBase(*newManifest)
		catalogRequest.CleanupRequest = cleanupservice.NewCleanupService()
		catalogRequest.Errors = []error{}

		resultOp := catalogSvc.PushMetadata(&catalogRequest)
		if resultOp.HasErrors() {
			errorMessage := "Error pushing manifest: \n"
			for _, err := range resultOp.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(newManifest)
		ctx.LogInfof("Manifest Tags Updated: %v", newManifest.ID)
	}
}

// @Summary		Removes tags from a catalog manifest version
// @Description	This endpoint removes tags from a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path		string										true	"Catalog ID"
// @Param			version			path		string										true	"Version"
// @Param			architecture	path		string										true	"Architecture"
// @Param			request			body		models.VirtualMachineCatalogManifestPatch	true	"Body"
// @Success		200				{object}	models.CatalogManifest
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/tags [delete]
func RemoveTagsToCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request catalog_models.VirtualMachineCatalogManifestPatch
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
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

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if request.Tags != nil && len(request.Tags) > 0 {
			if err := dbService.RemoveCatalogManifestTags(ctx, manifest.ID, request.Tags...); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
		} else {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "No tags provided",
				Code:    http.StatusBadRequest,
			})
			return
		}

		newManifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		catalogSvc := catalog.NewManifestService(ctx)
		catalogRequest := mappers.DtoCatalogManifestToBase(*newManifest)
		catalogRequest.CleanupRequest = cleanupservice.NewCleanupService()
		catalogRequest.Errors = []error{}

		resultOp := catalogSvc.PushMetadata(&catalogRequest)
		if resultOp.HasErrors() {
			errorMessage := "Error pushing manifest: \n"
			for _, err := range resultOp.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(newManifest)
		ctx.LogInfof("Manifest Claims Updated: %v", newManifest.ID)
	}
}

func CreateCatalogManifestHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request catalog_models.VirtualMachineCatalogManifest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(true); err != nil {
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

		ctx.LogInfof("Creating manifest %v", request.Name)
		dto := mappers.CatalogManifestToDto(request)
		result, err := dbService.CreateCatalogManifest(ctx, dto)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		resultData := mappers.DtoCatalogManifestToApi(*result)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest returned: %v", resultData.ID)
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
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]

		cleanRemote := http_helper.GetHttpRequestStrValue(r, constants.DELETE_REMOTE_MANIFEST_QUERY)
		// by default we will clean the remote manifest
		if cleanRemote == "" {
			cleanRemote = "true"
		}
		var errorDeletingRemote error
		var errorDeletingRecord error

		manifest := catalog.NewManifestService(ctx)
		if cleanRemote == "true" {
			ctx.LogInfof("Deleting remote manifest %v", catalogId)
			err = manifest.Delete(catalogId, "", "")
			if err != nil {
				errorDeletingRemote = err
			}
		}

		err = dbService.DeleteCatalogManifest(ctx, catalogId)
		if err != nil {
			errorDeletingRecord = err
		}

		if errorDeletingRecord != nil || errorDeletingRemote != nil {
			if errorDeletingRecord != nil && errorDeletingRemote != nil {
				ReturnApiError(ctx, w, models.NewFromError(errors.New(fmt.Sprintf("Error deleting record: %v, Error deleting remote: %v", errorDeletingRecord, errorDeletingRemote))))
				return
			}
			if errorDeletingRecord != nil {
				ReturnApiCommonResponseWithDataAndCode(w, models.NewFromError(errorDeletingRemote), http.StatusAccepted)
				return
			}
			if errorDeletingRemote != nil {
				ReturnApiCommonResponseWithDataAndCode(w, models.NewFromError(errorDeletingRemote), http.StatusAccepted)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Catalog manifest deleted successfully")
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
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]

		cleanRemote := http_helper.GetHttpRequestStrValue(r, constants.DELETE_REMOTE_MANIFEST_QUERY)
		// by default we will clean the remote manifest
		if cleanRemote == "" {
			cleanRemote = "true"
		}
		var errorDeletingRemote error
		var errorDeletingRecord error

		manifest := catalog.NewManifestService(ctx)
		if cleanRemote == "true" {
			ctx.LogInfof("Deleting remote manifest %v", catalogId)
			err = manifest.Delete(catalogId, version, "")
			if err != nil {
				errorDeletingRemote = err
			}
		}

		err = dbService.DeleteCatalogManifestVersion(ctx, catalogId, version)
		if err != nil {
			errorDeletingRecord = err
		}

		if errorDeletingRecord != nil || errorDeletingRemote != nil {
			if errorDeletingRecord != nil && errorDeletingRemote != nil {
				ReturnApiError(ctx, w, models.NewFromError(errors.New(fmt.Sprintf("Error deleting record: %v, Error deleting remote: %v", errorDeletingRecord, errorDeletingRemote))))
				return
			}
			if errorDeletingRecord != nil {
				ReturnApiCommonResponseWithDataAndCode(w, models.NewFromError(errorDeletingRemote), http.StatusAccepted)
				return
			}
			if errorDeletingRemote != nil {
				ReturnApiCommonResponseWithDataAndCode(w, models.NewFromError(errorDeletingRemote), http.StatusAccepted)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Catalog manifest deleted successfully")
	}
}

// @Summary		Deletes a catalog manifest version architecture
// @Description	This endpoint deletes a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId		path	string	true	"Catalog ID"
// @Param			version			path	string	true	"Version"
// @Param			architecture	path	string	true	"Architecture"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture} [delete]
func DeleteCatalogManifestVersionArchitectureHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		cleanRemote := http_helper.GetHttpRequestStrValue(r, constants.DELETE_REMOTE_MANIFEST_QUERY)
		// by default we will clean the remote manifest
		if cleanRemote == "" {
			cleanRemote = "true"
		}
		var errorDeletingRemote error
		var errorDeletingRecord error

		manifest := catalog.NewManifestService(ctx)
		if cleanRemote == "true" {
			ctx.LogInfof("Deleting remote manifest %v", catalogId)
			err = manifest.Delete(catalogId, version, architecture)
			if err != nil {
				errorDeletingRemote = err
			}
		}

		err = dbService.DeleteCatalogManifestVersionArch(ctx, catalogId, version, architecture)
		if err != nil {
			errorDeletingRecord = err
		}

		if errorDeletingRecord != nil || errorDeletingRemote != nil {
			if errorDeletingRecord != nil && errorDeletingRemote != nil {
				ReturnApiError(ctx, w, models.NewFromError(errors.New(fmt.Sprintf("Error deleting record: %v, Error deleting remote: %v", errorDeletingRecord, errorDeletingRemote))))
				return
			}
			if errorDeletingRecord != nil {
				ReturnApiCommonResponseWithDataAndCode(w, models.NewFromError(errorDeletingRemote), http.StatusAccepted)
				return
			}
			if errorDeletingRemote != nil {
				ReturnApiCommonResponseWithDataAndCode(w, models.NewFromError(errorDeletingRemote), http.StatusAccepted)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Catalog manifest deleted successfully")
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
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request catalog_models.PushCatalogManifestRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := catalog.NewManifestService(ctx)
		resultManifest := manifest.Push(&request)
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
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest pushed: %v", resultData.ID)
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
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request catalog_models.PullCatalogManifestRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		sendTelemetry := false
		var amplitudeEvent models.AmplitudeEvent
		telemetryItem := telemetry.TelemetryItem{}
		if request.AmplitudeEvent != "" {
			decodedBytes, err := base64.StdEncoding.DecodeString(request.AmplitudeEvent)
			test := string(decodedBytes)
			fmt.Println(test)
			if err == nil {
				err := json.Unmarshal(decodedBytes, &amplitudeEvent)
				if err == nil {
					telemetryItem.Type = amplitudeEvent.EventType
					if telemetryItem.Type == "" {
						telemetryItem.Type = "DEVOPS::PULL_MANIFEST"
					}
					telemetryItem.Properties = amplitudeEvent.EventProperties
					telemetryItem.Options = amplitudeEvent.UserProperties
					telemetryItem.UserID = amplitudeEvent.AppId
					telemetryItem.DeviceId = amplitudeEvent.DeviceId
					if amplitudeEvent.Origin != "" {
						telemetryItem.Properties["origin"] = amplitudeEvent.Origin
					}
					sendTelemetry = true
				} else {
					ctx.LogErrorf("Error unmarshalling amplitude event", err)
				}
			} else {
				ctx.LogErrorf("Error decoding amplitude event", err)
			}
		}

		manifest := catalog.NewManifestService(ctx)
		resultManifest := manifest.Pull(&request)

		if resultManifest.HasErrors() {
			if sendTelemetry && amplitudeEvent.EventProperties != nil {
				telemetryItem.Properties["success"] = "false"
				telemetry.TrackEvent(telemetryItem)
			}

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

		if sendTelemetry && amplitudeEvent.EventProperties != nil {
			telemetryItem.Properties["success"] = "true"
			telemetry.TrackEvent(telemetryItem)
		}

		resultData := mappers.BasePullCatalogManifestResponseToApi(*resultManifest)
		resultData.ID = resultManifest.ID

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest pulled: %v", resultData.ID)
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
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request catalog_models.ImportCatalogManifestRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := catalog.NewManifestService(ctx)
		resultManifest := manifest.Import(&request)
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
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest imported: %v", resultData.ID)
	}
}

// @Summary		Imports a vm into the catalog inventory generating the metadata for it
// @Description	This endpoint imports a virtual machine in pvm or macvm format into the catalog inventory generating the metadata for it
// @Tags			Catalogs
// @Produce		json
// @Param			importRequest	body		catalog_models.ImportVmRequest	true	"Vm Impoty request"
// @Success		200				{object}	models.ImportVmResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/import-vm [put]
func ImportVmHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request catalog_models.ImportVmRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := catalog.NewManifestService(ctx)
		resultManifest := manifest.ImportVm(&request)
		if resultManifest.HasErrors() {
			errorMessage := "Error importing vm: \n"
			for _, err := range resultManifest.Errors {
				errorMessage += "\n" + err.Error() + " "
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}

		resultData := mappers.BaseImportVmResponseToApi(*resultManifest)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resultData)
		ctx.LogInfof("Manifest imported: %v", resultData.ID)
	}
}

// @Summary		Updates a catalog
// @Description	This endpoint adds claims to a catalog manifest version
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path		string										true	"Catalog ID"
// @Param			request		body		models.VirtualMachineCatalogManifestPatch	true	"Body"
// @Success		200			{object}	models.CatalogManifest
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/{catalogId}/{version}/{architecture}/claims [patch]
func UpdateCatalogManifestProviderHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request catalog_models.VirtualMachineCatalogManifestPatch
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if request.Connection == "" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "No connection provided",
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		architecture := vars["architecture"]

		manifest, err := dbService.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		dboProvider := data_models.CatalogManifestProvider{
			Type: request.Provider.Type,
			Meta: request.Provider.Meta,
		}
		manifest.Provider = &dboProvider

		catalogSvc := catalog.NewManifestService(ctx)
		catalogRequest := mappers.DtoCatalogManifestToBase(*manifest)
		catalogRequest.CleanupRequest = cleanupservice.NewCleanupService()
		catalogRequest.Errors = []error{}

		resultOp := catalogSvc.PushMetadata(&catalogRequest)
		if resultOp.HasErrors() {
			errorMessage := "Error pushing manifest: \n"
			for _, err := range resultOp.Errors {
				if err == nil {
					errorMessage += "\n" + "error connecting to the provider" + " "
				} else {
					errorMessage += "\n" + err.Error() + " "
				}
			}
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: errorMessage,
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := dbService.UpdateCatalogManifestProvider(ctx, manifest.ID, dboProvider); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(manifest)
		ctx.LogInfof("Manifest Claims Updated: %v", manifest.ID)
	}
}

// @Summary		Gets catalog cache
// @Description	This endpoint returns all the remote catalog cache if any
// @Tags			Catalogs
// @Produce		json
// @Success		200	{object}	[]models.CatalogManifest
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/cache [get]
func GetCatalogCacheHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		catalogCacheSvc, err := cacheservice.NewCacheService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
		}
		items, err := catalogCacheSvc.GetAllCacheItems()
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		responseManifests := mappers.BaseVirtualMachineCatalogManifestListToApi(items)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(responseManifests)
		ctx.LogInfof("Manifests cached items returned: %v", len(items.Manifests))
	}
}

// @Summary		Deletes all catalog cache
// @Description	This endpoint returns all the remote catalog cache if any
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path	string	true	"Catalog ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/cache [delete]
func DeleteCatalogCacheHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		catalogCacheSvc, err := cacheservice.NewCacheService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
		}

		err = catalogCacheSvc.RemoveAllCacheItems()
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		// responseManifests := mappers.DtoCatalogManifestsToApi(manifestsDto)

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Removed all cached items")
	}
}

// @Summary		Deletes catalog cache item and all its versions
// @Description	This endpoint returns all the remote catalog cache if any and all its versions
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path	string	true	"Catalog ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/cache/{catalogId} [delete]
func DeleteCatalogCacheItemHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		if catalogId == "" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Catalog ID is required",
				Code:    http.StatusBadRequest,
			})
			return
		}

		catalogCacheSvc, err := cacheservice.NewCacheService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
		}

		err = catalogCacheSvc.RemoveCacheItem(catalogId, "")
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		// responseManifests := mappers.DtoCatalogManifestsToApi(manifestsDto)

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Manifests cached item %v removed", len(catalogId))
	}
}

// @Summary		Deletes catalog cache version item
// @Description	This endpoint deletes a version of a cache ite,
// @Tags			Catalogs
// @Produce		json
// @Param			catalogId	path	string	true	"Catalog ID"
// @Param			version		path	string	true	"Version"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog/cache/{catalogId}/{version} [delete]
func DeleteCatalogCacheItemVersionHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		catalogId := vars["catalogId"]
		version := vars["version"]
		if catalogId == "" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Catalog ID is required",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if version == "" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Version is required",
				Code:    http.StatusBadRequest,
			})
			return
		}

		catalogCacheSvc, err := cacheservice.NewCacheService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
		}

		err = catalogCacheSvc.RemoveCacheItem(catalogId, version)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		// responseManifests := mappers.DtoCatalogManifestsToApi(manifestsDto)

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Manifests cached item %v removed", len(catalogId))
	}
}
