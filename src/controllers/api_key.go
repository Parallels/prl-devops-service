package controllers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerApiKeysHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s ApiKeys handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).WithPath("/auth/api_keys").
		WithRequiredClaim(constants.LIST_API_KEY_CLAIM).
		WithHandler(GetApiKeysHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/api_keys/{id}").
		WithRequiredClaim(constants.LIST_API_KEY_CLAIM).
		WithHandler(GetApiKeyHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/api_keys").
		WithRequiredClaim(constants.CREATE_API_KEY_CLAIM).
		WithHandler(CreateApiKeyHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/auth/api_keys/{id}").
		WithRequiredClaim(constants.DELETE_API_KEY_CLAIM).
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

// @Summary		Gets all the api keys
// @Description	This endpoint returns all the api keys
// @Tags			Api Keys
// @Produce		json
// @Success		200	{object}	[]models.ApiKeyResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys [get]
func GetApiKeysHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoApiKeys, err := dbService.GetApiKeys(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := mappers.ApiKeysDtoToApiKeyResponse(dtoApiKeys)

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
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys/{id} [delete]
func DeleteApiKeyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteApiKey(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
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
// @Success		200	{object}	models.ApiKeyResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys/{id} [get]
func GetApiKeyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dtoApiKey, err := dbService.GetApiKey(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := mappers.ApiKeyDtoToApiKeyResponse(*dtoApiKey)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Api Key returned successfully")
	}
}

// @Summary		Creates an api key
// @Description	This endpoint creates an api key
// @Tags			Api Keys
// @Produce		json
// @Param			apiKey	body		models.ApiKeyRequest	true	"Body"
// @Success		200		{object}	models.ApiKeyResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys [post]
func CreateApiKeyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.ApiKeyRequest
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

		dtoApiKey := mappers.ApiKeyRequestToDto(request)

		dtoApiKeyResult, err := dbService.CreateApiKey(ctx, dtoApiKey)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}
		response := mappers.ApiKeyDtoToApiKeyResponse(*dtoApiKeyResult)
		response.Encoded = base64.StdEncoding.EncodeToString([]byte(request.Key + ":" + request.Secret))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Api Key created successfully")
	}
}

// @Summary		Revoke an api key
// @Description	This endpoint revokes an api key
// @Tags			Api Keys
// @Produce		json
// @Param			id	path	string	true	"Api Key ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/api_keys/{id}/revoke [put]
func RevokeApiKeyHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.RevokeKey(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Api Key revoked successfully")
	}
}
