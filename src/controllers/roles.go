package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"

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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles  [get]
func GetRolesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getRolesDiag := errors.NewDiagnostics("/auth/roles [get]")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getRolesDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getRolesDiag, rsp.Code))
			return
		}

		dtoRoles, err := dbService.GetRoles(ctx, GetFilterHeader(r))
		if err != nil {
			rsp := models.NewFromError(err)
			getRolesDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DatabaseService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getRolesDiag, rsp.Code))
			return
		}

		if len(dtoRoles) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.RoleResponse, 0)
			defer r.Body.Close()
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles/{id}  [get]
func GetRoleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getRoleDiag := errors.NewDiagnostics("/auth/roles/{id} [get]")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getRoleDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getRoleDiag, rsp.Code))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dtoRole, err := dbService.GetRole(ctx, strings.ToUpper(helpers.NormalizeString(id)))
		if err != nil {
			rsp := models.NewFromError(err)
			getRoleDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DatabaseService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getRoleDiag, rsp.Code))
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
// @Failure		400			{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401			{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles  [post]
func CreateRoleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.RoleRequest
		createRoleDiag := errors.NewDiagnostics("/auth/roles [post]")
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			createRoleDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createRoleDiag, http.StatusBadRequest))
			return
		}
		if err := request.Validate(); err != nil {
			createRoleDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createRoleDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			createRoleDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createRoleDiag, rsp.Code))
			return
		}

		dtoRole := mappers.ApiRoleToDto(request)

		role, err := dbService.CreateRole(ctx, dtoRole)
		if err != nil {
			rsp := models.NewFromError(err)
			createRoleDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DatabaseService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createRoleDiag, rsp.Code))
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles/{id}  [delete]
func DeleteRoleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		deleteRoleDiag := errors.NewDiagnostics("/auth/roles/{id} [delete]")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {

			rsp := models.NewFromError(err)
			deleteRoleDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteRoleDiag, rsp.Code))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteRole(ctx, id)
		if err != nil {
			rsp := models.NewFromError(err)
			deleteRoleDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DatabaseService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteRoleDiag, rsp.Code))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Role deleted successfully")
	}
}
