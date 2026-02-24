package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/cacheservice"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/gorilla/mux"
)

func registerCacheHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Cache handlers", version)

	// Old endpoints
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

	// New endpoints
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/cache").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(GetCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/cache").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/cache/{catalogId}").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteCatalogCacheItemHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/cache/{catalogId}/{version}").
		WithRequiredClaim(constants.SUPER_USER_ROLE).
		WithHandler(DeleteCatalogCacheItemVersionHandler()).
		Register()
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
// @Router			/v1/cache [get]
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
		// obfuscate provider credentials for external calls
		cfg := config.Get()
		enableObfuscation := cfg.EnableCredentialsObfuscation()
		internalClientHeader := r.Header.Get(constants.INTERNAL_API_CLIENT)
		if internalClientHeader != "true" && enableObfuscation {
			for k, items := range responseManifests.Manifests {
				if items.Provider != nil {
					newProvider := &models.RemoteVirtualMachineProvider{}
					newProvider.Type = items.Provider.Type
					newProvider.Host = items.Provider.Host
					newProvider.Port = items.Provider.Port
					newProvider.Username = helpers.ObfuscateString(items.Provider.Username)
					newProvider.Password = helpers.ObfuscateString(items.Provider.Password)
					newProvider.ApiKey = helpers.ObfuscateString(items.Provider.ApiKey)
					if items.Provider.Meta != nil {
						newProvider.Meta = make(map[string]string)
						for k, v := range items.Provider.Meta {
							newProvider.Meta[k] = helpers.ObfuscateString(v)
						}
					}
					items.Provider = newProvider
				}
				responseManifests.Manifests[k] = items
			}
		}

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
// @Router			/v1/cache [delete]
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
// @Router			/v1/cache/{catalogId} [delete]
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
// @Router			/v1/cache/{catalogId}/{version} [delete]
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
