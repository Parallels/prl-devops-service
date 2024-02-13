package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/security/password"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerUsersHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Users handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/users").
		WithRequiredClaim(constants.LIST_USER_CLAIM).
		WithHandler(GetUsersHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/users/{id}").
		WithRequiredClaim(constants.LIST_USER_CLAIM).
		WithHandler(GetUserHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/users").
		WithRequiredClaim(constants.CREATE_USER_CLAIM).
		WithHandler(CreateUserHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/auth/users/{id}").
		WithRequiredClaim(constants.UPDATE_USER_CLAIM).
		WithHandler(UpdateUserHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/auth/users/{id}").
		WithRequiredClaim(constants.DELETE_USER_CLAIM).
		WithHandler(DeleteUserHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/users/{id}/roles").
		WithRequiredClaim(constants.LIST_USER_CLAIM).
		WithHandler(GetUserRolesHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/users/{id}/roles").
		WithRequiredClaim(constants.UPDATE_USER_CLAIM).
		WithHandler(AddRoleToUserHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/auth/users/{id}/roles/{role_id}").
		WithRequiredClaim(constants.UPDATE_USER_CLAIM).
		WithHandler(RemoveRoleFromUserHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/users/{id}/claims").
		WithRequiredClaim(constants.LIST_USER_CLAIM).
		WithHandler(GetUserClaimsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/users/{id}/claims").
		WithRequiredClaim(constants.UPDATE_USER_CLAIM).
		WithHandler(AddClaimToUserHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/auth/users/{id}/claims/{claim_id}").
		WithRequiredClaim(constants.UPDATE_USER_CLAIM).
		WithHandler(RemoveClaimFromUserHandler()).
		Register()
}

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
func GetUsersHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		users, err := dbService.GetUsers(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(users) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.ApiUser, 0)
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Users returned: %v", len(response))
			return
		}

		result := mappers.DtoUsersToApiResponse(users)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Users returned: %v", len(result))
	}
}

// @Summary		Gets a user
// @Description	This endpoint returns a user
// @Tags			Users
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	models.ApiUser
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}  [get]
func GetUserHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		user, err := dbService.GetUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoUserToApiResponse(*user)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("User returned: %v", response.ID)
	}
}

// @Summary		Creates a user
// @Description	This endpoint creates a user
// @Tags			Users
// @Produce		json
// @Param			body	body		models.UserCreateRequest	true	"User"
// @Success		201		{object}	models.ApiUser
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users  [post]
func CreateUserHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UserCreateRequest
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
		if request.Password != "" {
			passwordSvc := password.Get()
			if valid, diag := passwordSvc.CheckPasswordComplexity(request.Password); diag.HasErrors() {
				ReturnApiError(ctx, w, models.ApiErrorResponse{
					Message: diag.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			} else {
				if !valid {
					ReturnApiError(ctx, w, models.ApiErrorResponse{
						Message: "Invalid Password, please check complexity rules",
						Code:    http.StatusBadRequest,
					})
					return
				}
			}
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoUser, err := dbService.CreateUser(ctx, mappers.ApiUserCreateRequestToDto(request))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoUserToApiResponse(*dtoUser)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("User created: %v", response.ID)
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
func DeleteUserHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("User deleted: %v", id)
	}
}

// @Summary		Update a user
// @Description	This endpoint updates a user
// @Tags			Users
// @Produce		json
// @Param			body	body		models.UserCreateRequest	true	"User"
// @Success		202		{object}	models.ApiUser
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}  [put]
func UpdateUserHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UserCreateRequest
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
		ctx.LogInfof("User updated: %v", id)
	}
}

// @Summary		Gets all the roles for a user
// @Description	This endpoint returns all the roles for a user
// @Tags			Users
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	models.RoleResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/roles  [get]
func GetUserRolesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

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
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Roles returned: %v", len(result))
	}
}

// @Summary		Adds a role to a user
// @Description	This endpoint adds a role to a user
// @Tags			Users
// @Produce		json
// @Param			id		path		string				true	"User ID"
// @Param			body	body		models.RoleRequest	true	"Role Name"
// @Success		201		{object}	models.RoleRequest
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/roles  [post]
func AddRoleToUserHandler() restapi.ControllerHandler {
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

		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddRoleToUser(ctx, id, request.Name); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(request)
		ctx.LogInfof("Role added to user: %v", id)
	}
}

// @Summary		Removes a role from a user
// @Description	This endpoint removes a role from a user
// @Tags			Users
// @Produce		json
// @Param			id		path	string	true	"User ID"
// @Param			role_id	path	string	true	"Role ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/roles/{role_id}  [delete]
func RemoveRoleFromUserHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		roleId := vars["role_id"]

		if err = dbService.RemoveRoleFromUser(ctx, id, roleId); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Role removed from user: %v", id)
	}
}

// @Summary		Gets all the claims for a user
// @Description	This endpoint returns all the claims for a user
// @Tags			Users
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	models.ClaimResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/claims  [get]
func GetUserClaimsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

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
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Claims returned: %v", len(result))
	}
}

// @Summary		Adds a claim to a user
// @Description	This endpoint adds a claim to a user
// @Tags			Users
// @Produce		json
// @Param			id		path		string				true	"User ID"
// @Param			body	body		models.ClaimRequest	true	"Claim Name"
// @Success		201		{object}	models.ClaimRequest
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/claims  [post]
func AddClaimToUserHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.ClaimRequest
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

		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddClaimToUser(ctx, id, request.Name); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(request)
		ctx.LogInfof("Claim added to user: %v", id)
	}
}

// @Summary		Removes a claim from a user
// @Description	This endpoint removes a claim from a user
// @Tags			Users
// @Produce		json
// @Param			id			path	string	true	"User ID"
// @Param			claim_id	path	string	true	"Claim ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/users/{id}/claims/{claim_id}  [delete]
func RemoveClaimFromUserHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]
		claimId := vars["claim_id"]

		if err = dbService.RemoveClaimFromUser(ctx, id, claimId); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Claim removed from user: %v", id)
	}
}
