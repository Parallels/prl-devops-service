package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/gorilla/mux"
)

func registerCatalogManagerHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Catalog Managers handlers", version)

	// GET ALL
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers").
		WithHandler(GetCatalogManagersHandler()).
		Register()

	// GET BY ID
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/catalog-managers/{id}").
		WithHandler(GetCatalogManagerByIdHandler()).
		Register()

	// POST
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/catalog-managers").
		WithRequiredClaim(constants.CREATE_CATALOG_MANAGER_CLAIM).
		WithRequiredClaim(constants.CREATE_OWN_CATALOG_MANAGER_CLAIM).
		WithHandler(CreateCatalogManagerHandler()).
		Register()

	// PUT
	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/catalog-managers/{id}").
		WithRequiredClaim(constants.UPDATE_CATALOG_MANAGER_CLAIM).
		WithHandler(UpdateCatalogManagerHandler()).
		Register()

	// DELETE
	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/catalog-managers/{id}").
		WithRequiredClaim(constants.DELETE_CATALOG_MANAGER_CLAIM).
		WithHandler(DeleteCatalogManagerHandler()).
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

		canListAll := false
		canListOwn := false
		for _, claim := range user.Claims {
			if claim == constants.LIST_CATALOG_MANAGER_CLAIM {
				canListAll = true
			}
			if claim == constants.LIST_OWN_CATALOG_MANAGER_CLAIM {
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
						for _, userClaim := range user.Claims {
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

		canListAll := false
		canListOwn := false
		for _, claim := range user.Claims {
			if claim == constants.LIST_CATALOG_MANAGER_CLAIM {
				canListAll = true
			}
			if claim == constants.LIST_OWN_CATALOG_MANAGER_CLAIM {
				canListOwn = true
			}
		}

		isAuthorized := canListAll || (canListOwn && mgr.OwnerID == user.ID)
		if !isAuthorized && canListOwn && mgr.Global {
			hasAllRequired := true
			if len(mgr.RequiredClaims) > 0 {
				for _, reqClaim := range mgr.RequiredClaims {
					found := false
					for _, userClaim := range user.Claims {
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

		includeCredentials := false
		if canListAll || mgr.OwnerID == user.ID {
			includeCredentials = true
		}

		response := mappers.ToCatalogManagerResponse(mgr, includeCredentials)
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
			if claim == constants.CREATE_CATALOG_MANAGER_CLAIM {
				canCreateGlobalInternal = true
			}
		}
		if req.Global || req.Internal {
			if !canCreateGlobalInternal {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Not authorized to create global or internal catalog manager"), http.StatusForbidden))
				return
			}
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
		canSystemUpdate := false
		for _, claim := range user.Claims {
			if claim == constants.UPDATE_CATALOG_MANAGER_CLAIM {
				canSystemUpdate = true
			}
		}
		for _, role := range user.Roles {
			if role == constants.SUPER_USER_ROLE {
				canSystemUpdate = true
			}
		}

		if mgr.OwnerID != user.ID && !canSystemUpdate {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(errors.New("Not authorized to update this catalog manager"), http.StatusForbidden))
			return
		}

		mappers.UpdateCatalogManagerFromRequest(mgr, &req)
		mgr.UpdatedAt = helpers.GetUtcCurrentDateTime()

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
		canSystemDelete := false
		for _, claim := range user.Claims {
			if claim == constants.DELETE_CATALOG_MANAGER_CLAIM {
				canSystemDelete = true
			}
		}
		for _, role := range user.Roles {
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
