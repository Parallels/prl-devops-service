package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

//	@Summary		Gets all the roles
//	@Description	This endpoint returns all the roles
//	@Tags			Roles
//	@Produce		json
//	@Success		200	{object}	[]models.RoleResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/roles  [get]
func GetRolesController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		dtoRoles, err := dbService.GetRoles(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(dtoRoles) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.RoleResponse, 0)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Roles returned: %v", len(response))
			return
		}

		result := mappers.DtoRolesToApi(dtoRoles)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Roles returned successfully")
	}
}

//	@Summary		Gets a role
//	@Description	This endpoint returns a role
//	@Tags			Roles
//	@Produce		json
//	@Param			id	path		string	true	"Role ID"
//	@Success		200	{object}	models.RoleResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/roles/{id}  [get]
func GetRoleController() restapi.ControllerHandler {
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

		dtoRole, err := dbService.GetRole(ctx, strings.ToUpper(helpers.NormalizeString(id)))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := mappers.DtoRoleToApi(*dtoRole)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Role returned successfully")
	}
}

//	@Summary		Gets a role
//	@Description	This endpoint returns a role
//	@Tags			Roles
//	@Produce		json
//	@Param			roleRequest	body		models.RoleRequest	true	"Role Request"
//	@Success		200			{object}	models.RoleResponse
//	@Failure		400			{object}	models.ApiErrorResponse
//	@Failure		401			{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/roles  [post]
func CreateRoleController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.RoleRequest
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

		dtoRole := mappers.ApiRoleToDto(request)

		role, err := dbService.CreateRole(ctx, dtoRole)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoRoleToApi(*role)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Role created successfully")
	}
}

//	@Summary		Delete a role
//	@Description	This endpoint deletes a role
//	@Tags			Roles
//	@Produce		json
//	@Param			id	path	string	true	"Role ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/roles/{id}  [delete]
func DeleteRoleController() restapi.ControllerHandler {
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

		err = dbService.DeleteRole(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Role deleted successfully")
	}
}
