package controllers

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	catalog_helpers "github.com/Parallels/prl-devops-service/catalog/common"
	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/security"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerCatalogManagerHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Catalog Managers handlers", version)

	// GET ALL
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_OWN_CLAIM).
		WithHandler(GetCatalogManagersHandler()).
		Register()

	// GET BY ID
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers/{id}").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_OWN_CLAIM).
		WithHandler(GetCatalogManagerByIdHandler()).
		Register()

	// POST
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/catalog-managers").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_CREATE_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_CREATE_OWN_CLAIM).
		WithHandler(CreateCatalogManagerHandler()).
		Register()

	// PUT
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog-managers/{id}").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_OWN_CLAIM).
		WithHandler(UpdateCatalogManagerHandler()).
		Register()

	// DELETE
	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog-managers/{id}").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_DELETE_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_DELETE_OWN_CLAIM).
		WithHandler(DeleteCatalogManagerHandler()).
		Register()

	registerCatalogManagerCatalogHandlers(version)
}

func registerCatalogManagerCatalogHandlers(version string) {
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog").
		WithOrClaims().
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog")).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}").
		WithOrClaims().
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}")).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}").
		WithOrClaims().
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}")).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}").
		WithOrClaims().
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}")).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog").
		WithOrClaims().
		WithRequiredClaim(constants.CREATE_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_CREATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog")).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}").
		WithOrClaims().
		WithRequiredClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_DELETE_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}")).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}").
		WithOrClaims().
		WithRequiredClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_DELETE_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}")).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}").
		WithOrClaims().
		WithRequiredClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_DELETE_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}")).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/download").
		WithOrClaims().
		WithRequiredClaim(constants.LIST_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_LIST_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/download")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/taint").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/taint")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/untaint").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/untaint")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/revoke").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/revoke")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/claims").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/claims")).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/claims").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/claims")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/roles").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/roles")).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/roles").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/roles")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/tags").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/tags")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PATCH).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/connection").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/connection")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/metadata").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/metadata")).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/{catalogId}/{version}/{architecture}/tags").
		WithOrClaims().
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_UPDATE_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/{catalogId}/{version}/{architecture}/tags")).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/push").
		WithOrClaims().
		WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_PUSH_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/push")).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/push/async").
		WithOrClaims().
		WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_PUSH_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(AsyncPushCatalogManifestToCatalogManagerHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/import").
		WithOrClaims().
		WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_IMPORT_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/import")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/import-vm").
		WithOrClaims().
		WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_IMPORT_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(ForwardCatalogManagerCatalogRequestHandler("/catalog/import-vm")).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/pull").
		WithOrClaims().
		WithRequiredClaim(constants.PULL_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_PULL_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(PullCatalogManifestFromCatalogManagerHandler(false)).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog-managers/{id}/catalog/pull/async").
		WithOrClaims().
		WithRequiredClaim(constants.PULL_CATALOG_MANIFEST_CLAIM).
		WithRequiredClaim(constants.CATALOG_MANAGER_PULL_CATALOG_MANIFEST_OWN_CLAIM).
		WithHandler(PullCatalogManifestFromCatalogManagerHandler(true)).
		Register()
}

// @Summary		Gets all the catalog managers
// @Description	This endpoint returns all the catalog managers
// @Tags			CatalogManagers
// @Produce		json
// @Success		200	{object}	[]models.CatalogManager
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog-managers [get]
func GetCatalogManagersHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		user := ctx.GetUser()
		if user == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("User not contextually found"), http.StatusUnauthorized))
			return
		}

		authCtx := ctx.GetAuthorizationContext()
		effectiveClaims := authCtx.GetEffectiveClaims()

		canListAll := false
		canListOwn := false
		for _, claim := range effectiveClaims {
			if claim == constants.CATALOG_MANAGER_LIST_CLAIM {
				canListAll = true
			}
			if claim == constants.CATALOG_MANAGER_LIST_OWN_CLAIM {
				canListOwn = true
			}
		}

		if !canListAll && !canListOwn {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Not authorized to view catalog managers"), http.StatusForbidden))
			return
		}

		allMgrs, err := dbService.GetCatalogManagers()
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		filteredMgrs := make([]data_models.CatalogManager, 0)

		for _, mgr := range allMgrs {
			if canListAll || (mgr.OwnerID == user.ID) {
				filteredMgrs = append(filteredMgrs, mgr)
				continue
			}

			// If it's global, user must meet the required claims
			if canListOwn && mgr.Global {
				hasAllRequired := true
				if len(mgr.RequiredClaims) > 0 {
					for _, reqClaim := range mgr.RequiredClaims {
						found := false
						for _, userClaim := range effectiveClaims {
							if userClaim == reqClaim {
								found = true
								break
							}
						}
						if !found {
							hasAllRequired = false
							break
						}
					}
				}
				if hasAllRequired {
					filteredMgrs = append(filteredMgrs, mgr)
				}
			}
		}

		// determine if we should show credentials (e.g if they are an admin or owner)
		// For list, we'll obscure it to be safe, they can GET individually if they want to view details
		response := mappers.ToCatalogManagerResponseList(filteredMgrs, false)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// @Summary		Gets a specific catalog manager
// @Description	This endpoint returns a catalog manager
// @Tags			CatalogManagers
// @Produce		json
// @Param			id	path		string	true	"Manager ID"
// @Success		200	{object}	models.CatalogManager
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog-managers/{id} [get]
func GetCatalogManagerByIdHandler() restapi.ControllerHandler {
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
		id := vars["id"]

		user := ctx.GetUser()
		if user == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("User not contextually found"), http.StatusUnauthorized))
			return
		}

		mgr, err := dbService.GetCatalogManager(id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusNotFound))
			return
		}

		authCtx := ctx.GetAuthorizationContext()
		effectiveClaims := authCtx.GetEffectiveClaims()

		canListAll := false
		canListOwn := false
		for _, claim := range effectiveClaims {
			if claim == constants.CATALOG_MANAGER_LIST_CLAIM {
				canListAll = true
			}
			if claim == constants.CATALOG_MANAGER_LIST_OWN_CLAIM {
				canListOwn = true
			}
		}

		isAuthorized := canListAll || (canListOwn && mgr.OwnerID == user.ID)
		if !isAuthorized && canListOwn && mgr.Global {
			hasAllRequired := true
			if len(mgr.RequiredClaims) > 0 {
				for _, reqClaim := range mgr.RequiredClaims {
					found := false
					for _, userClaim := range effectiveClaims {
						if userClaim == reqClaim {
							found = true
							break
						}
					}
					if !found {
						hasAllRequired = false
						break
					}
				}
			}
			isAuthorized = hasAllRequired
		}

		if !isAuthorized {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Not authorized to view this catalog manager"), http.StatusForbidden))
			return
		}

		response := mappers.ToCatalogManagerResponse(mgr, false)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// @Summary		Creates a catalog manager
// @Description	This endpoint creates a catalog manager
// @Tags			CatalogManagers
// @Produce		json
// @Success		200	{object}	models.CatalogManager
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog-managers [post]
func CreateCatalogManagerHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var req models.CatalogManagerRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusBadRequest))
			return
		}
		defer r.Body.Close()

		user := ctx.GetUser()

		newMgr := mappers.FromCatalogManagerRequest(&req)
		newMgr.OwnerID = user.ID
		newMgr.CreatedAt = helpers.GetUtcCurrentDateTime()
		newMgr.UpdatedAt = helpers.GetUtcCurrentDateTime()

		canCreateGlobalInternal := false
		for _, claim := range user.Claims {
			if claim == constants.CATALOG_MANAGER_CREATE_CLAIM {
				canCreateGlobalInternal = true
			}
		}
		if req.Global || req.Internal {
			if !canCreateGlobalInternal {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Not authorized to create global or internal catalog manager"), http.StatusForbidden))
				return
			}
		}

		if err := validateCatalogManagerConnection(ctx, newMgr.URL, newMgr.Username, decryptCatalogManagerSecret(newMgr.Password), decryptCatalogManagerSecret(newMgr.ApiKey)); err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		err = dbService.AddCatalogManager(ctx, *newMgr)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mappers.ToCatalogManagerResponse(newMgr, false))
	}
}

// @Summary		Updates a catalog manager
// @Description	This endpoint updates a catalog manager
// @Tags			CatalogManagers
// @Produce		json
// @Param			id	path		string	true	"Manager ID"
// @Success		200	{object}	models.CatalogManager
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog-managers/{id} [put]
func UpdateCatalogManagerHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		var req models.CatalogManagerRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusBadRequest))
			return
		}
		defer r.Body.Close()

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		mgr, err := dbService.GetCatalogManager(id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusNotFound))
			return
		}

		user := ctx.GetUser()
		authCtxUpdate := ctx.GetAuthorizationContext()
		effectiveClaimsUpdate := authCtxUpdate.GetEffectiveClaims()
		effectiveRolesUpdate := authCtxUpdate.GetEffectiveRoles()
		canSystemUpdate := false
		for _, claim := range effectiveClaimsUpdate {
			if claim == constants.CATALOG_MANAGER_UPDATE_CLAIM {
				canSystemUpdate = true
			}
		}
		for _, role := range effectiveRolesUpdate {
			if role == constants.SUPER_USER_ROLE {
				canSystemUpdate = true
			}
		}

		if mgr.OwnerID != user.ID && !canSystemUpdate {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Not authorized to update this catalog manager"), http.StatusForbidden))
			return
		}

		updatedMgr := *mgr
		mappers.UpdateCatalogManagerFromRequest(&updatedMgr, &req)
		updatedMgr.UpdatedAt = helpers.GetUtcCurrentDateTime()

		if err := validateCatalogManagerConnection(ctx, updatedMgr.URL, updatedMgr.Username, decryptCatalogManagerSecret(updatedMgr.Password), decryptCatalogManagerSecret(updatedMgr.ApiKey)); err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusBadRequest))
			return
		}

		*mgr = updatedMgr

		err = dbService.UpdateCatalogManager(ctx, *mgr)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mappers.ToCatalogManagerResponse(mgr, false))
	}
}

// @Summary		Deletes a catalog manager
// @Description	This endpoint deletes a catalog manager
// @Tags			CatalogManagers
// @Produce		json
// @Param			id	path		string	true	"Manager ID"
// @Success		200	{object}	models.ApiCommonResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/catalog-managers/{id} [delete]
func DeleteCatalogManagerHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		id := vars["id"]

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		mgr, err := dbService.GetCatalogManager(id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusNotFound))
			return
		}

		user := ctx.GetUser()
		authCtxDelete := ctx.GetAuthorizationContext()
		effectiveClaimsDelete := authCtxDelete.GetEffectiveClaims()
		effectiveRolesDelete := authCtxDelete.GetEffectiveRoles()
		canSystemDelete := false
		for _, claim := range effectiveClaimsDelete {
			if claim == constants.CATALOG_MANAGER_DELETE_CLAIM {
				canSystemDelete = true
			}
		}
		for _, role := range effectiveRolesDelete {
			if role == constants.SUPER_USER_ROLE {
				canSystemDelete = true
			}
		}

		if mgr.OwnerID != user.ID && !canSystemDelete {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Not authorized to delete this catalog manager"), http.StatusForbidden))
			return
		}

		err = dbService.DeleteCatalogManager(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusOK)
		res := models.ApiCommonResponse{
			Operation: "Catalog Manager deleted successfully",
			Success:   true,
		}
		_ = json.NewEncoder(w).Encode(res)
	}
}

// AsyncPushCatalogManifestToCatalogManagerHandler handles async catalog pushes
// to a remote catalog manager. It runs the full push pipeline locally (compress,
// upload to storage, register manifest) using the connection string stored on the
// catalog manager, and tracks progress via the local job system.
func AsyncPushCatalogManifestToCatalogManagerHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "User not found"})
			return
		}

		vars := mux.Vars(r)
		managerID := vars["id"]

		mgr, errResp := getAuthorizedCatalogManagerForUse(ctx, managerID)
		if errResp != nil {
			ReturnApiError(ctx, w, *errResp)
			return
		}

		var request catalog_models.PushCatalogManifestRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if request.Connection == "" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "connection is required in the request body (storage provider connection string, e.g. provider=minio;endpoint=...;bucket=...)",
			})
			return
		}

		// Build the host= part from the catalog manager credentials and merge with the
		// user-provided storage connection (stripping any host= they may have included).
		hostPart, err := buildCatalogManagerConnection(*mgr)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusBadRequest))
			return
		}
		storageParts := stripHostFromConnection(request.Connection)
		request.Connection = hostPart + ";" + storageParts

		arch, err := catalog_helpers.ValidateArch(request.Architecture)
		if err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid architecture: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		request.Architecture = arch

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		jobManager := jobs.Get(ctx)
		if jobManager == nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Job Manager is not available"), http.StatusInternalServerError))
			return
		}

		localJob, err := jobManager.CreateNewJob(userContext.ID, "catalog", "push", "Initializing catalog push to manager "+mgr.Name)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		asyncCtx := basecontext.NewRootBaseContext()
		manifestService := catalog.NewManifestService(asyncCtx)
		go manifestService.AsyncPush(localJob.ID, &request)

		response := mappers.MapJobToApiJob(*localJob)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("[CatalogManager] Async push to catalog manager %s started, job ID: %s", mgr.Name, localJob.ID)
	}
}

func ForwardCatalogManagerCatalogRequestHandler(remotePathTemplate string) restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		vars := mux.Vars(r)
		managerID := vars["id"]

		mgr, err := getAuthorizedCatalogManagerForUse(ctx, managerID)
		if err != nil {
			ReturnApiError(ctx, w, *err)
			return
		}

		relativePath := resolveMuxPath(remotePathTemplate, vars)
		if err := forwardCatalogManagerRequest(ctx, w, r, mgr, relativePath); err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusBadGateway))
			return
		}
	}
}

func PullCatalogManifestFromCatalogManagerHandler(async bool) restapi.ControllerHandler {
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

		if request.Connection != "" {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "connection is not allowed for catalog manager pull endpoints",
				Code:    http.StatusBadRequest,
			})
			return
		}

		vars := mux.Vars(r)
		managerID := vars["id"]

		mgr, errResp := getAuthorizedCatalogManagerForUse(ctx, managerID)
		if errResp != nil {
			ReturnApiError(ctx, w, *errResp)
			return
		}

		connection, err := buildCatalogManagerConnection(*mgr)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusBadRequest))
			return
		}

		request.Connection = connection
		payload, err := json.Marshal(request)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(payload))
		r.ContentLength = int64(len(payload))

		if async {
			AsyncPullCatalogManifestHandler()(w, r)
		} else {
			PullCatalogManifestHandler()(w, r)
		}
	}
}

func getAuthorizedCatalogManagerForUse(ctx basecontext.ApiContext, managerID string) (*data_models.CatalogManager, *models.ApiErrorResponse) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		apiErr := models.NewFromErrorWithCode(err, http.StatusInternalServerError)
		return nil, &apiErr
	}

	user := ctx.GetUser()
	if user == nil {
		apiErr := models.NewFromErrorWithCode(errors.New("User not contextually found"), http.StatusUnauthorized)
		return nil, &apiErr
	}

	mgr, err := dbService.GetCatalogManager(managerID)
	if err != nil {
		apiErr := models.NewFromErrorWithCode(err, http.StatusNotFound)
		return nil, &apiErr
	}

	if !mgr.Active {
		apiErr := models.NewFromErrorWithCode(errors.New("catalog manager is disabled"), http.StatusBadRequest)
		return nil, &apiErr
	}

	authCtxUse := ctx.GetAuthorizationContext()
	effectiveClaimsUse := authCtxUse.GetEffectiveClaims()
	effectiveRolesUse := authCtxUse.GetEffectiveRoles()

	isSuperUser := false
	for _, role := range effectiveRolesUse {
		if role == constants.SUPER_USER_ROLE {
			isSuperUser = true
			break
		}
	}

	if isSuperUser || mgr.OwnerID == user.ID {
		return mgr, nil
	}

	if mgr.Global {
		hasAllRequiredClaims := true
		for _, requiredClaim := range mgr.RequiredClaims {
			found := false
			for _, userClaim := range effectiveClaimsUse {
				if userClaim == requiredClaim {
					found = true
					break
				}
			}

			if !found {
				hasAllRequiredClaims = false
				break
			}
		}

		if hasAllRequiredClaims {
			return mgr, nil
		}
	}

	apiErr := models.NewFromErrorWithCode(errors.New("Not authorized to use this catalog manager"), http.StatusForbidden)
	return nil, &apiErr
}

func resolveMuxPath(pathTemplate string, vars map[string]string) string {
	resolvedPath := pathTemplate
	for key, value := range vars {
		resolvedPath = strings.ReplaceAll(resolvedPath, "{"+key+"}", value)
	}
	return resolvedPath
}

func buildCatalogManagerTargetUrl(baseURL string, endpointPath string, rawQuery string) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	basePath := strings.TrimRight(parsedURL.Path, "/")
	if strings.HasSuffix(basePath, "/api/v1") {
		parsedURL.Path = basePath + endpointPath
	} else {
		parsedURL.Path = path.Join(basePath, "api/v1") + endpointPath
	}

	parsedURL.RawQuery = rawQuery
	return parsedURL.String(), nil
}

func decryptCatalogManagerSecret(value string) string {
	if value == "" {
		return ""
	}

	cfg := config.Get()
	if cfg == nil || cfg.EncryptionPrivateKey() == "" {
		return value
	}

	decrypted, err := security.DecryptString(cfg.EncryptionPrivateKey(), []byte(value))
	if err != nil {
		return value
	}

	return decrypted
}

func getCatalogManagerAuthorizer(ctx basecontext.ApiContext, manager data_models.CatalogManager, targetUrl string) (*apiclient.HttpClientServiceAuthorizer, error) {
	password := decryptCatalogManagerSecret(manager.Password)
	apiKey := decryptCatalogManagerSecret(manager.ApiKey)

	client := apiclient.NewHttpClient(ctx)
	if apiKey != "" {
		client.AuthorizeWithApiKey(apiKey)
	}

	if manager.Username != "" && password != "" {
		client.AuthorizeWithUsernameAndPassword(manager.Username, password)
	}

	if apiKey == "" && (manager.Username == "" || password == "") {
		return nil, nil
	}

	return client.Authorize(ctx, targetUrl)
}

func buildCatalogManagerConnection(manager data_models.CatalogManager) (string, error) {
	host := strings.TrimSpace(manager.URL)
	if host == "" {
		return "", errors.New("catalog manager url is required")
	}

	password := decryptCatalogManagerSecret(manager.Password)
	apiKey := decryptCatalogManagerSecret(manager.ApiKey)

	if apiKey != "" {
		return fmt.Sprintf("host=%s@%s", apiKey, host), nil
	}

	if manager.Username != "" && password != "" {
		return fmt.Sprintf("host=%s:%s@%s", manager.Username, password, host), nil
	}

	return "host=" + host, nil
}

// stripHostFromConnection removes any host= segment from a semicolon-separated
// connection string, returning only the storage-provider parts.
func stripHostFromConnection(connection string) string {
	parts := strings.Split(connection, ";")
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(p)), "host=") {
			filtered = append(filtered, p)
		}
	}
	return strings.Join(filtered, ";")
}

func validateCatalogManagerConnection(ctx basecontext.ApiContext, managerURL string, username string, password string, apiKey string) error {
	if strings.TrimSpace(managerURL) == "" {
		return errors.New("catalog manager url is required")
	}

	if apiKey == "" && username == "" && password == "" {
		return errors.New("missing authentication credentials, provide api_key or username/password")
	}

	if apiKey == "" && ((username != "" && password == "") || (username == "" && password != "")) {
		return errors.New("both username and password are required for password authentication")
	}

	targetURL, err := buildCatalogManagerTargetUrl(managerURL, "/catalog", "")
	if err != nil {
		return fmt.Errorf("invalid catalog manager url: %w", err)
	}

	client := apiclient.NewHttpClient(ctx)
	if apiKey != "" {
		client.AuthorizeWithApiKey(apiKey)
	} else {
		client.AuthorizeWithUsernameAndPassword(username, password)
	}

	var response interface{}
	if _, err := client.Get(targetURL, &response); err != nil {
		return fmt.Errorf("unable to authenticate catalog manager connection: %w", err)
	}

	return nil
}

func forwardCatalogManagerRequest(ctx basecontext.ApiContext, w http.ResponseWriter, r *http.Request, manager *data_models.CatalogManager, endpointPath string) error {
	targetURL, err := buildCatalogManagerTargetUrl(manager.URL, endpointPath, r.URL.RawQuery)
	if err != nil {
		return err
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	outboundRequest, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}

	authorizer, err := getCatalogManagerAuthorizer(ctx, *manager, targetURL)
	if err != nil {
		return err
	}

	for headerKey, values := range r.Header {
		if strings.EqualFold(headerKey, "Authorization") ||
			strings.EqualFold(headerKey, "X-Api-Key") ||
			strings.EqualFold(headerKey, constants.INTERNAL_API_CLIENT) ||
			strings.EqualFold(headerKey, constants.X_CLAIMS_HEADER) ||
			strings.EqualFold(headerKey, constants.X_ROLES_HEADER) {
			continue
		}
		for _, value := range values {
			outboundRequest.Header.Add(headerKey, value)
		}
	}
	outboundRequest.Header.Set("X-SOURCE", "CATALOG_MANAGER_REQUEST")
	outboundRequest.Header.Set(constants.INTERNAL_API_CLIENT, "false")

	// Inject the current user's effective claims and roles so the downstream
	// catalog service can filter results using the calling user's permissions
	// rather than the stored catalog-manager credentials.
	if fwdUser := ctx.GetUser(); fwdUser != nil {
		fwdAuthCtx := ctx.GetAuthorizationContext()
		claims := fwdAuthCtx.GetEffectiveClaims()
		roles := fwdAuthCtx.GetEffectiveRoles()
		if len(claims) > 0 {
			outboundRequest.Header.Set(constants.X_CLAIMS_HEADER,
				base64.StdEncoding.EncodeToString([]byte(strings.Join(claims, ","))))
		}
		if len(roles) > 0 {
			outboundRequest.Header.Set(constants.X_ROLES_HEADER,
				base64.StdEncoding.EncodeToString([]byte(strings.Join(roles, ","))))
		}
	}

	if authorizer != nil {
		if authorizer.BearerToken != "" {
			outboundRequest.Header.Set("Authorization", "Bearer "+authorizer.BearerToken)
		} else if authorizer.ApiKey != "" {
			outboundRequest.Header.Set("X-Api-Key", authorizer.ApiKey)
		}
	}

	cfg := config.Get()
	disableTLSValidation := false
	if cfg != nil {
		disableTLSValidation = cfg.DisableTlsValidation()
	}
	transport := &http.Transport{
		TLSHandshakeTimeout: 30 * time.Second,
		IdleConnTimeout:     60 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: disableTLSValidation,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Minute,
	}

	response, err := httpClient.Do(outboundRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Skip headers managed by local middleware to avoid duplicates
	skipHeaders := map[string]bool{
		"Access-Control-Allow-Origin":      true,
		"Access-Control-Allow-Methods":     true,
		"Access-Control-Allow-Headers":     true,
		"Access-Control-Allow-Credentials": true,
		"Access-Control-Expose-Headers":    true,
		"Access-Control-Max-Age":           true,
		"Permissions-Policy":               true,
		"Referrer-Policy":                  true,
		"Strict-Transport-Security":        true,
		"X-Content-Type-Options":           true,
		"X-Frame-Options":                  true,
		"X-Robots-Tag":                     true,
		"Via":                              true,
	}
	for key, values := range response.Header {
		if skipHeaders[key] {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(response.StatusCode)
	_, err = io.Copy(w, response.Body)
	return err
}
