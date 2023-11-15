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

//	@Summary		Gets all the api keys
//	@Description	This endpoint returns all the api keys
//	@Tags			Api Keys
//	@Produce		json
//	@Success		200	{object}	[]models.ApiKeyResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/api_keys [get]
func GetApiKeysController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		dtoApiKeys, err := dbService.GetApiKeys(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := mappers.ApiKeysDtoToApiKeyResponse(dtoApiKeys)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Api Keys returned successfully")
	}
}

//	@Summary		Deletes an api key
//	@Description	This endpoint deletes an api key
//	@Tags			Api Keys
//	@Param			id	path	string	true	"Api Key ID"
//	@Produce		json
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/api_keys/{id} [delete]
func DeleteApiKeyController() restapi.ControllerHandler {
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

		err = dbService.DeleteApiKey(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Api Key deleted successfully")
	}
}

//	@Summary		Gets an api key by id or name
//	@Description	This endpoint returns an api key by id or name
//	@Tags			Api Keys
//	@Param			id	path	string	true	"Api Key ID"
//	@Produce		json
//	@Success		200	{object}	models.ApiKeyResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/api_keys/{id} [get]
func GetApiKeyByIdOrNameController() restapi.ControllerHandler {
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

		dtoApiKey, err := dbService.GetApiKey(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := mappers.ApiKeyDtoToApiKeyResponse(*dtoApiKey)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Api Key returned successfully")
	}
}

//	@Summary		Creates an api key
//	@Description	This endpoint creates an api key
//	@Tags			Api Keys
//	@Produce		json
//	@Param			apiKey	body		models.ApiKeyRequest	true	"Body"
//	@Success		200		{object}	models.ApiKeyResponse
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/api_keys [post]
func CreateApiKeyController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.ApiKeyRequest
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

		dtoApiKey := mappers.ApiKeyRequestToDto(request)

		dtoApiKeyResult, err := dbService.CreateApiKey(ctx, dtoApiKey)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}
		response := mappers.ApiKeyDtoToApiKeyResponse(*dtoApiKeyResult)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Api Key created successfully")
	}
}

//	@Summary		Revoke an api key
//	@Description	This endpoint revokes an api key
//	@Tags			Api Keys
//	@Produce		json
//	@Param			id	path	string	true	"Api Key ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/api_keys/{id}/revoke [put]
func RevokeApiKeyController() restapi.ControllerHandler {
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

		err = dbService.RevokeKey(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Api Key revoked successfully")
	}
}
