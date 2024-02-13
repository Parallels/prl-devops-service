package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerRolesHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Roles handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/roles").
		WithRequiredClaim(constants.LIST_ROLE_CLAIM).
		WithHandler(GetRolesHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/roles/{id}").
		WithRequiredClaim(constants.LIST_ROLE_CLAIM).
		WithHandler(GetRoleHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/roles").
		WithRequiredClaim(constants.CREATE_ROLE_CLAIM).
		WithHandler(CreateRoleHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/auth/roles/{id}").
		WithRequiredClaim(constants.DELETE_ROLE_CLAIM).
		WithHandler(DeleteRoleHandler()).
		Register()
}

// @Summary		Gets all the roles
// @Description	This endpoint returns all the roles
// @Tags			Roles
// @Produce		json
// @Success		200	{object}	[]models.RoleResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles  [get]
func GetRolesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoRoles, err := dbService.GetRoles(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(dtoRoles) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.RoleResponse, 0)
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Roles returned: %v", len(response))
			return
		}

		result := mappers.DtoRolesToApi(dtoRoles)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Roles returned successfully")
	}
}

// @Summary		Gets a role
// @Description	This endpoint returns a role
// @Tags			Roles
// @Produce		json
// @Param			id	path		string	true	"Role ID"
// @Success		200	{object}	models.RoleResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles/{id}  [get]
func GetRoleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dtoRole, err := dbService.GetRole(ctx, strings.ToUpper(helpers.NormalizeString(id)))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := mappers.DtoRoleToApi(*dtoRole)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Role returned successfully")
	}
}

// @Summary		Gets a role
// @Description	This endpoint returns a role
// @Tags			Roles
// @Produce		json
// @Param			roleRequest	body		models.RoleRequest	true	"Role Request"
// @Success		200			{object}	models.RoleResponse
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles  [post]
func CreateRoleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.RoleRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
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

		dtoRole := mappers.ApiRoleToDto(request)

		role, err := dbService.CreateRole(ctx, dtoRole)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoRoleToApi(*role)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Role created successfully")
	}
}

// @Summary		Delete a role
// @Description	This endpoint deletes a role
// @Tags			Roles
// @Produce		json
// @Param			id	path	string	true	"Role ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles/{id}  [delete]
func DeleteRoleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteRole(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Role deleted successfully")
	}
}
