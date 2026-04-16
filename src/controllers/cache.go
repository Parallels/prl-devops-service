package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/cacheservice"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	diskspace "github.com/Parallels/prl-devops-service/serviceprovider/diskSpace"
	"github.com/gorilla/mux"
)

func registerCacheHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Cache handlers", version)

	// Old endpoints
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog/cache").
		WithRequiredClaim(constants.LIST_CACHE_CLAIM).
		WithHandler(GetCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/cache").
		WithRequiredClaim(constants.DELETE_ALL_CACHE_CLAIM).
		WithHandler(DeleteCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/cache/{catalogId}").
		WithRequiredClaim(constants.DELETE_CACHE_ITEM_CLAIM).
		WithHandler(DeleteCatalogCacheItemHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog/cache/{catalogId}/{version}").
		WithRequiredClaim(constants.DELETE_CACHE_ITEM_CLAIM).
		WithHandler(DeleteCatalogCacheItemVersionHandler()).
		Register()

	// New endpoints
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/cache").
		WithRequiredClaim(constants.LIST_CACHE_CLAIM).
		WithHandler(GetCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/cache").
		WithRequiredClaim(constants.DELETE_ALL_CACHE_CLAIM).
		WithHandler(DeleteCatalogCacheHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/cache/{catalogId}").
		WithRequiredClaim(constants.DELETE_CACHE_ITEM_CLAIM).
		WithHandler(DeleteCatalogCacheItemHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/cache/{catalogId}/{version}").
		WithRequiredClaim(constants.DELETE_CACHE_ITEM_CLAIM).
		WithHandler(DeleteCatalogCacheItemVersionHandler()).
		Register()
}

// @Summary		Gets catalog cache
// @Description	This endpoint returns all the remote catalog cache if any
// @Tags			Catalogs
// @Produce		json
// @Success		200	{object}	[]models.CatalogManifest
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/cache [get]
func GetCatalogCacheHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getCatalogCacheDiag := errors.NewDiagnostics("/cache")
		catalogCacheSvc, err := cacheservice.NewCacheService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getCatalogCacheDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "NewCacheService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getCatalogCacheDiag, rsp.Code))
			return
		}
		items, err := catalogCacheSvc.GetAllCacheItems()
		if err != nil {
			rsp := models.NewFromError(err)
			getCatalogCacheDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetAllCacheItems")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getCatalogCacheDiag, rsp.Code))
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

		var freeDiskSpace int64
		localDiag := errors.NewDiagnostics("GetCacheDiskSpaceFallback")
		if ds := diskspace.Get(ctx).GetCacheDiskSpace(ctx, localDiag); !localDiag.HasErrors() {
			freeDiskSpace = ds
		} else if hwInfo, err := serviceprovider.Get().System.GetHardwareInfo(ctx); err == nil {
			freeDiskSpace = int64(hwInfo.FreeDiskSize)
		}

		if cfg.IsHost() {
			responseManifests.CacheConfig = &models.CatalogCacheConfig{
				Enabled:                 cfg.IsCatalogCachingEnable(),
				KeepFreeDiskSpace:       cfg.CacheKeepFreeDiskSpace(),
				MaxSize:                 cfg.CacheMaxSize(freeDiskSpace),
				AllowAboveFreeDiskSpace: cfg.AllowCacheAboveFreeDiskSpace(),
			}
			if path, err := cfg.CatalogCacheFolder(); err == nil {
				responseManifests.CacheConfig.Folder = path
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/cache [delete]
func DeleteCatalogCacheHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		deleteCatalogCacheDiag := errors.NewDiagnostics("/cache")
		catalogCacheSvc, err := cacheservice.NewCacheService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			deleteCatalogCacheDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "NewCacheService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheDiag, rsp.Code))
			return
		}

		err = catalogCacheSvc.RemoveAllCacheItems()
		if err != nil {
			rsp := models.NewFromError(err)
			deleteCatalogCacheDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "RemoveAllCacheItems")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheDiag, rsp.Code))
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
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
		deleteCatalogCacheItemDiag := errors.NewDiagnostics("/cache/" + catalogId)
		if catalogId == "" {
			deleteCatalogCacheItemDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Catalog ID is required", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheItemDiag, http.StatusBadRequest))
			return
		}

		catalogCacheSvc, err := cacheservice.NewCacheService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			deleteCatalogCacheItemDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "NewCacheService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheItemDiag, rsp.Code))
			return
		}

		err = catalogCacheSvc.RemoveCacheItem(catalogId, "")
		if err != nil {
			rsp := models.NewFromError(err)
			deleteCatalogCacheItemDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "RemoveCacheItem")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheItemDiag, rsp.Code))
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
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
		deleteCatalogCacheItemVersionDiag := errors.NewDiagnostics("/cache/" + catalogId + "/" + version)
		if catalogId == "" {
			deleteCatalogCacheItemVersionDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Catalog ID is required", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheItemVersionDiag, http.StatusBadRequest))
			return
		}

		if version == "" {
			deleteCatalogCacheItemVersionDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Version is required", "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheItemVersionDiag, http.StatusBadRequest))
			return
		}

		catalogCacheSvc, err := cacheservice.NewCacheService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			deleteCatalogCacheItemVersionDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "NewCacheService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheItemVersionDiag, rsp.Code))
			return
		}

		err = catalogCacheSvc.RemoveCacheItem(catalogId, version)
		if err != nil {
			rsp := models.NewFromError(err)
			deleteCatalogCacheItemVersionDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "RemoveCacheItem")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteCatalogCacheItemVersionDiag, rsp.Code))
			return
		}

		// responseManifests := mappers.DtoCatalogManifestsToApi(manifestsDto)

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Manifests cached item %v removed", len(catalogId))
	}
}
