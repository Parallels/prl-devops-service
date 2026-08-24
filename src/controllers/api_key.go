package controllers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func enrichApiKeyWithUser(ctx *basecontext.BaseContext, apiKey *models.ApiKeyResponse) {
	if apiKey.UserID == "" {
		return
	}

	dbService, diag := serviceprovider.GetDatabaseService(ctx)
	if diag != nil && diag.HasErrors() {
		return
	}

	user, diag := dbService.Stores().User().GetUserByID(*ctx, apiKey.UserID)
	if diag != nil && diag.HasErrors() {
		return
	}

	if user != nil {
		apiKey.UserEmail = user.Email
		apiKey.UserName = user.Name
		apiKey.UserUsername = user.Username
	}
}

func registerApiKeysHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s ApiKeys handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).WithPath("/auth/api_keys").
		WithRequiredClaim(constants.LIST_API_KEY_CLAIM).
		WithRequiredClaim(constants.LIST_OWN_API_KEY_CLAIM).
		WithHandler(GetApiKeysHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/api_keys/{id}").
		WithRequiredClaim(constants.LIST_API_KEY_CLAIM).
		WithRequiredClaim(constants.LIST_OWN_API_KEY_CLAIM).
		WithHandler(GetApiKeyHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/api_keys").
		WithRequiredClaim(constants.CREATE_API_KEY_CLAIM).
		WithRequiredClaim(constants.CREATE_OWN_API_KEY_CLAIM).
		WithHandler(CreateApiKeyHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/auth/api_keys/{id}").
		WithRequiredClaim(constants.DELETE_API_KEY_CLAIM).
		WithRequiredClaim(constants.DELETE_OWN_API_KEY_CLAIM).
		WithHandler(DeleteApiKeyHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/auth/api_keys/{id}/revoke").
		WithRequiredRole(constants.SUPER_USER_ROLE).
		WithHandler(RevokeApiKeyHandler()).
		Register()
}

// @Summary		Creates an api key
// @Description	This endpoint creates an api key
// @Content		# This endpoint will create an api key in the system
// @Content		API Keys are used to authenticate with the system from external applications
// @Content		## How are they different from a user?
// @Content		A user normally has a password and is used to authenticate with the system
// @Content		An api key is used to authenticate with the system from an external application
// @Tags			Api Keys
// @Produce		json
// @Claims			"CREATE_API_KEY"
// @Claims			"LIST"
// @Roles			"SUPER_USER"
// @Param			apiKey		body			models.ApiKeyRequest	true	"Body"
// @HeaderParam	x-filter	string  false	"Filter entities"
// @Success		200			{object}		models.ApiKeyResponse
// @Failure		400			{object}		models.ApiErrorDiagnosticsResponse
// @Failure		401			{object}		models.OAuthErrorResponse
// @Examples		{
// @Examples		"key": "Some Key",
// @Examples		"secret": "SomeLongSecret"
// @Examples		}
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys [post]
func CreateApiKeyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		createApiKeyDiag := errors.NewDiagnostics("/auth/api_keys [post]")
		var request models.ApiKeyRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			createApiKeyDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createApiKeyDiag, http.StatusBadRequest))
			return
		}

		if err := request.Validate(); err != nil {
			createApiKeyDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createApiKeyDiag, http.StatusBadRequest))
			return
		}

		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil && diag.HasErrors() {
			createApiKeyDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createApiKeyDiag, http.StatusInternalServerError))
			return
		}

		dbApiKey := mappers.ApiKeyRequestToDbModel(request)

		authContext := ctx.GetAuthorizationContext()
		if authContext != nil && authContext.User != nil {
			hasFullCreateClaim := authContext.HasEffectiveClaim(constants.CREATE_API_KEY_CLAIM)
			hasOwnCreateClaim := authContext.HasEffectiveClaim(constants.CREATE_OWN_API_KEY_CLAIM)

			if hasOwnCreateClaim {
				// Users with CREATE_OWN_API_KEY_CLAIM can only create for themselves
				if !hasFullCreateClaim && request.UserID != "" && request.UserID != authContext.User.ID {
					createApiKeyDiag.AddError(strconv.Itoa(http.StatusForbidden), "You do not have permission to create API keys for other users", "Validation", nil)
					ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createApiKeyDiag, http.StatusForbidden))
					return
				}
				// Auto-assign to user's ID if not provided
				if dbApiKey.UserID == "" {
					dbApiKey.UserID = authContext.User.ID
				}
			}
		}

		dbApiKeyResult, diag := dbService.Stores().ApiKey().CreateApiKey(*ctx, &dbApiKey)
		if diag != nil && diag.HasErrors() {
			createApiKeyDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createApiKeyDiag, http.StatusInternalServerError))
			return
		}
		response := mappers.ApiKeyDbModelToApiKeyResponse(*dbApiKeyResult)
		response.Encoded = base64.StdEncoding.EncodeToString([]byte(request.Key + ":" + request.Secret))

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Api Key created successfully")
	}
}

// @Summary		Gets all the api keys
// @Description	This endpoint returns all the api keys
// @Tags			Api Keys
// @Produce		json
// @Claims			"LIST_API_KEY"
// @Success		200	{object}	[]models.ApiKeyResponse
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys [get]
func GetApiKeysHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getApiKeysDiag := errors.NewDiagnostics("/auth/api_keys [get]")
		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil && diag.HasErrors() {
			getApiKeysDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getApiKeysDiag, http.StatusInternalServerError))
			return
		}

		queryBuilder := filters.NewQueryBuilder(GetFilterHeader(r))
		queryResponse, diag := dbService.Stores().ApiKey().GetApiKeysByQuery(*ctx, queryBuilder)
		if diag != nil && diag.HasErrors() {
			getApiKeysDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getApiKeysDiag, http.StatusInternalServerError))
			return
		}

		result := mappers.ApiKeysDbModelToApiKeyResponse(queryResponse.Items)

		for i := range result {
			enrichApiKeyWithUser(ctx, &result[i])
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Api Keys returned successfully")
	}
}

// @Summary		Deletes an api key
// @Description	This endpoint deletes an api key
// @Tags			Api Keys
// @Param			id	path	string	true	"Api Key ID"
// @Produce		json
// @Claims			"DELETE_API_KEY"
// @Success		202
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys/{id} [delete]
func DeleteApiKeyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		deleteApiKeyDiag := errors.NewDiagnostics("/auth/api_keys/{id} [delete]")
		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil && diag.HasErrors() {
			deleteApiKeyDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteApiKeyDiag, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		diag = dbService.Stores().ApiKey().DeleteApiKey(*ctx, id)
		if diag != nil && diag.HasErrors() {
			deleteApiKeyDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteApiKeyDiag, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Api Key deleted successfully")
	}
}

// @Summary		Gets an api key by id or name
// @Description	This endpoint returns an api key by id or name
// @Tags			Api Keys
// @Param			id	path	string	true	"Api Key ID"
// @Produce		json
// @Claims			"LIST_API_KEY"
// @Success		200	{object}	models.ApiKeyResponse
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys/{id} [get]
func GetApiKeyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getApiKeyDiag := errors.NewDiagnostics("/auth/api_keys/{id} [get]")
		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil && diag.HasErrors() {
			getApiKeyDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getApiKeyDiag, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dbApiKey, diag := dbService.Stores().ApiKey().GetApiKey(*ctx, id)
		if diag != nil && diag.HasErrors() {
			getApiKeyDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getApiKeyDiag, http.StatusInternalServerError))
			return
		}
		
		if dbApiKey == nil {
			getApiKeyDiag.AddError(strconv.Itoa(http.StatusNotFound), "API Key not found", "GetApiKey", nil)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getApiKeyDiag, http.StatusNotFound))
			return
		}

		response := mappers.ApiKeyDbModelToApiKeyResponse(*dbApiKey)

		enrichApiKeyWithUser(ctx, &response)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Api Key returned successfully")
	}
}

// @Summary		Revoke an api key
// @Description	This endpoint revokes an api key
// @Tags			Api Keys
// @Produce		json
// @Claims			"LIST_API_KEY"
// @Claims			"DELETE_API_KEY"
// @Roles			"SUPER_USER"
// @Param			id	path	string	true	"Api Key ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys/{id}/revoke [put]
func RevokeApiKeyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		revokeApiKeyDiag := errors.NewDiagnostics("/auth/api_keys/{id}/revoke [put]")
		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil && diag.HasErrors() {
			revokeApiKeyDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(revokeApiKeyDiag, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		diag = dbService.Stores().ApiKey().RevokeApiKey(*ctx, id)
		if diag != nil && diag.HasErrors() {
			revokeApiKeyDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(revokeApiKeyDiag, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Api Key revoked successfully")
	}
}
