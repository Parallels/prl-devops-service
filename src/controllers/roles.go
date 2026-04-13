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

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/roles/{id}/claims").
		WithRequiredClaim(constants.LIST_ROLE_CLAIM).
		WithHandler(GetRoleClaimsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/roles/{id}/claims").
		WithRequiredClaim(constants.UPDATE_ROLE_CLAIM).
		WithHandler(AddClaimToRoleHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/auth/roles/{id}/claims/{claim_id}").
		WithRequiredClaim(constants.UPDATE_ROLE_CLAIM).
		WithHandler(RemoveClaimFromRoleHandler()).
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

// @Summary		Creates a role
// @Description	This endpoint creates a role
// @Tags			Roles
// @Produce		json
// @Param			roleRequest	body		models.RoleRequest	true	"Role Request"
// @Success		201			{object}	models.RoleResponse
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

		emitAuthEvent(constants.EventAuthRoleAdded, models.AuthRoleEvent{RoleID: response.ID})
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

		emitAuthEvent(constants.EventAuthRoleRemoved, models.AuthRoleEvent{RoleID: id})
		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Role deleted successfully")
	}
}

// @Summary		Gets all claims for a role
// @Description	This endpoint returns all claims associated with a role
// @Tags			Roles
// @Produce		json
// @Param			id	path		string	true	"Role ID"
// @Success		200	{object}	[]models.ClaimResponse
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles/{id}/claims  [get]
func GetRoleClaimsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/auth/roles/{id}/claims [get]")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, rsp.Code))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dtoRole, err := dbService.GetRole(ctx, strings.ToUpper(helpers.NormalizeString(id)))
		if err != nil {
			rsp := models.NewFromError(err)
			diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DatabaseService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, rsp.Code))
			return
		}

		result := mappers.DtoClaimsToApi(dtoRole.Claims)
		if result == nil {
			result = []models.ClaimResponse{}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Role claims returned successfully")
	}
}

// @Summary		Adds a claim to a role
// @Description	This endpoint adds a claim to a role
// @Tags			Roles
// @Produce		json
// @Param			id		path		string					true	"Role ID"
// @Param			body	body		models.RoleClaimRequest	true	"Claim Name"
// @Success		201		{object}	models.ClaimResponse
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles/{id}/claims  [post]
func AddClaimToRoleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/auth/roles/{id}/claims [post]")
		var request models.RoleClaimRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusBadRequest))
			return
		}
		if err := request.Validate(); err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, rsp.Code))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddClaimToRole(ctx, id, request.Name); err != nil {
			rsp := models.NewFromError(err)
			diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DatabaseService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, rsp.Code))
			return
		}

		dtoRole, err := dbService.GetRole(ctx, id)
		if err != nil {
			rsp := models.NewFromError(err)
			diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DatabaseService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, rsp.Code))
			return
		}

		// Return the specific claim that was added.
		claimName := strings.ToUpper(helpers.NormalizeString(request.Name))
		for _, c := range dtoRole.Claims {
			if strings.EqualFold(c.ID, claimName) {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(models.ClaimResponse{ID: c.ID, Name: c.Name})
				ctx.LogInfof("Claim %s added to role %s", request.Name, id)
				return
			}
		}

		emitAuthEvent(constants.EventAuthRoleClaimAdded, models.AuthRoleClaimEvent{RoleID: id, ClaimID: request.Name})
		w.WriteHeader(http.StatusCreated)
		ctx.LogInfof("Claim %s added to role %s", request.Name, id)
	}
}

// @Summary		Removes a claim from a role
// @Description	This endpoint removes a claim from a role
// @Tags			Roles
// @Produce		json
// @Param			id			path	string	true	"Role ID"
// @Param			claim_id	path	string	true	"Claim ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/roles/{id}/claims/{claim_id}  [delete]
func RemoveClaimFromRoleHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		diag := errors.NewDiagnostics("/auth/roles/{id}/claims/{claim_id} [delete]")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, rsp.Code))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		claimId := vars["claim_id"]

		if err := dbService.RemoveClaimFromRole(ctx, id, claimId); err != nil {
			rsp := models.NewFromError(err)
			diag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DatabaseService")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(diag, rsp.Code))
			return
		}

		emitAuthEvent(constants.EventAuthRoleClaimRemoved, models.AuthRoleClaimEvent{RoleID: id, ClaimID: claimId})
		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Claim %s removed from role %s", claimId, id)
	}
}
