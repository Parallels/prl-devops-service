package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

// @Summary		Gets all the users
// @Description	This endpoint returns all the users
// @Tags			Users
// @Produce		json
// @Success		200	{object}	[]models.ApiUser
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users  [get]
func GetUsersController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		users, err := dbService.GetUsers(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(users) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.ApiUser, 0)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Users returned: %v", len(response))
			return
		}

		result := mappers.DtoUsersToApiResponse(users)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Users returned: %v", len(result))
	}
}

// @Summary		Gets a user
// @Description	This endpoint returns a user
// @Tags			Users
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		200	{object}	models.ApiUser
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}  [get]
func GetUserController() restapi.ControllerHandler {
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

		user, err := dbService.GetUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoUserToApiResponse(*user)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("User returned: %v", response.ID)
	}
}

// @Summary		Creates a user
// @Description	This endpoint creates a user
// @Tags			Users
// @Produce		json
// @Param			body	body	models.UserCreateRequest	true	"User"
// @Success		201	{object}	models.ApiUser
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users  [post]
func CreateUserController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UserCreateRequest
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

		dtoUser, err := dbService.CreateUser(ctx, mappers.ApiUserCreateRequestToDto(request))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoUserToApiResponse(*dtoUser)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("User created: %v", response.ID)
	}
}

// @Summary		Deletes a user
// @Description	This endpoint deletes a user
// @Tags			Users
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}  [delete]
func DeleteUserController() restapi.ControllerHandler {
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

		err = dbService.DeleteUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("User deleted: %v", id)
	}
}

// @Summary		Update a user
// @Description	This endpoint updates a user
// @Tags			Users
// @Produce		json
// @Param			body	body	models.UserCreateRequest	true	"User"
// @Success		202	{object}	models.ApiUser
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}  [put]
func UpdateUserController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UserCreateRequest
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

		vars := mux.Vars(r)
		id := vars["id"]

		dtoUser := mappers.ApiUserCreateRequestToDto(request)
		dtoUser.ID = id
		err = dbService.UpdateUser(ctx, dtoUser)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("User updated: %v", id)
	}
}

// @Summary		Gets all the roles for a user
// @Description	This endpoint returns all the roles for a user
// @Tags			Users
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		200	{object}	models.RoleResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/roles  [get]
func GetUserRolesController() restapi.ControllerHandler {
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

		dtoUser, err := dbService.GetUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		roles := dtoUser.Roles
		result := mappers.DtoRolesToApi(roles)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Roles returned: %v", len(result))
	}
}

// @Summary		Adds a role to a user
// @Description	This endpoint adds a role to a user
// @Tags			Users
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Param			body	body	models.RoleRequest	true	"Role Name"
// @Success		201	{object}	models.RoleRequest
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/roles  [post]
func AddRoleToUserController() restapi.ControllerHandler {
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

		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddRoleToUser(ctx, id, request.Name); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		dbService.Disconnect(ctx)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(request)
		ctx.LogInfo("Role added to user: %v", id)
	}
}

// @Summary		Removes a role from a user
// @Description	This endpoint removes a role from a user
// @Tags			Users
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Param			role_id	path	string	true	"Role ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/roles/{role_id}  [post]
func RemoveRoleFromUserController() restapi.ControllerHandler {
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
		roleId := vars["role_id"]

		if err = dbService.RemoveRoleFromUser(ctx, id, roleId); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Role removed from user: %v", id)
	}
}

// @Summary		Gets all the claims for a user
// @Description	This endpoint returns all the claims for a user
// @Tags			Users
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		200	{object}	models.ClaimResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/claims  [get]
func GetUserClaimsController() restapi.ControllerHandler {
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

		dtoUser, err := dbService.GetUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		claims := dtoUser.Claims
		result := mappers.DtoClaimsToApi(claims)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Claims returned: %v", len(result))
	}
}

// @Summary		Adds a claim to a user
// @Description	This endpoint adds a claim to a user
// @Tags			Users
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Param			body	body	models.ClaimRequest	true	"Claim Name"
// @Success		201	{object}	models.ClaimRequest
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/claims  [post]
func AddClaimToUserController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.ClaimRequest
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

		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddClaimToUser(ctx, id, request.Name); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		dbService.Disconnect(ctx)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(request)
		ctx.LogInfo("Claim added to user: %v", id)
	}
}

// @Summary		Removes a claim from a user
// @Description	This endpoint removes a claim from a user
// @Tags			Users
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Param			claim_id	path	string	true	"Claim ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/claims/{claim_id}  [post]
func RemoveClaimFromUserController() restapi.ControllerHandler {
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
		claimId := vars["claim_id"]

		if err = dbService.RemoveClaimFromUser(ctx, id, claimId); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Claim removed from user: %v", id)
	}
}
