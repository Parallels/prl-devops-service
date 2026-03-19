package controllers

import (
	"encoding/json"
	"html"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerUserConfigsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s UserConfigs handlers", version)

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/user/configs").
		WithAuthorization().
		WithHandler(GetUserConfigsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/user/configs/{id}").
		WithAuthorization().
		WithHandler(GetUserConfigHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/user/configs").
		WithAuthorization().
		WithHandler(CreateUserConfigHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/user/configs/{id}").
		WithAuthorization().
		WithHandler(UpdateUserConfigHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/user/configs/{id}").
		WithAuthorization().
		WithHandler(DeleteUserConfigHandler()).
		Register()
}

// @Summary		Gets all user configs
// @Description	This endpoint returns all configuration entries for the authenticated user
// @Tags			User Configs
// @Produce		json
// @Success		200	{object}	[]models.UserConfigResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs [get]
func GetUserConfigsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "user not found"})
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoConfigs, err := dbService.GetUserConfigs(ctx, userContext.ID, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := mappers.UserConfigsDtoToResponse(dtoConfigs)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("User configs returned successfully")
	}
}

// @Summary		Gets a user config by id or slug
// @Description	This endpoint returns a single configuration entry for the authenticated user
// @Tags			User Configs
// @Param			id	path	string	true	"Config ID or Slug"
// @Produce		json
// @Success		200	{object}	models.UserConfigResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Failure		404	{object}	models.ApiErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs/{id} [get]
func GetUserConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "user not found"})
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := html.UnescapeString(vars["id"])

		dtoConfig, err := dbService.GetUserConfig(ctx, userContext.ID, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusNotFound))
			return
		}

		response := mappers.UserConfigDtoToResponse(*dtoConfig)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("User config returned successfully")
	}
}

// @Summary		Creates a user config
// @Description	This endpoint creates a configuration entry for the authenticated user
// @Tags			User Configs
// @Produce		json
// @Param			userConfig	body	models.UserConfigRequest	true	"Body"
// @Success		201	{object}	models.UserConfigResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs [post]
func CreateUserConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "user not found"})
			return
		}

		var request models.UserConfigRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoConfig := mappers.UserConfigRequestToDto(userContext.ID, request)

		existing, _ := dbService.GetUserConfig(ctx, userContext.ID, request.Slug)

		dtoResult, err := dbService.UpsertUserConfig(ctx, dtoConfig)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.UserConfigDtoToResponse(*dtoResult)

		if existing != nil {
			w.WriteHeader(http.StatusOK)
			ctx.LogInfof("User config updated successfully (upsert)")
		} else {
			w.WriteHeader(http.StatusCreated)
			ctx.LogInfof("User config created successfully")
		}
		_ = json.NewEncoder(w).Encode(response)
	}
}

// @Summary		Updates a user config
// @Description	This endpoint updates a configuration entry for the authenticated user
// @Tags			User Configs
// @Param			id			path	string						true	"Config ID or Slug"
// @Param			userConfig	body	models.UserConfigUpdateRequest	true	"Body"
// @Produce		json
// @Success		200	{object}	models.UserConfigResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Failure		404	{object}	models.ApiErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs/{id} [put]
func UpdateUserConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "user not found"})
			return
		}

		vars := mux.Vars(r)
		id := html.UnescapeString(vars["id"])

		var request models.UserConfigUpdateRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		existing, err := dbService.GetUserConfig(ctx, userContext.ID, id)
		if err != nil {
			// Not found — create a new record using the path id as slug.
			name := request.Name
			if name == "" {
				name = id
			}
			cfgType := data_models.UserConfigValueType(request.Type)
			if cfgType == "" {
				cfgType = data_models.UserConfigValueTypeString
			}
			newCfg := data_models.UserConfig{
				UserID: userContext.ID,
				Slug:   id,
				Name:   name,
				Type:   cfgType,
				Value:  request.Value,
			}
			dtoResult, createErr := dbService.UpsertUserConfig(ctx, newCfg)
			if createErr != nil {
				ReturnApiError(ctx, w, models.NewFromError(createErr))
				return
			}
			response := mappers.UserConfigDtoToResponse(*dtoResult)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("User config created successfully (upsert)")
			return
		}

		if request.Name != "" {
			existing.Name = request.Name
		}
		if request.Type != "" {
			existing.Type = data_models.UserConfigValueType(request.Type)
		}
		if request.Value != "" {
			existing.Value = request.Value
		}

		dtoResult, err := dbService.UpdateUserConfig(ctx, *existing)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.UserConfigDtoToResponse(*dtoResult)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("User config updated successfully")
	}
}

// @Summary		Deletes a user config
// @Description	This endpoint deletes a configuration entry for the authenticated user
// @Tags			User Configs
// @Param			id	path	string	true	"Config ID or Slug"
// @Produce		json
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Failure		404	{object}	models.ApiErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs/{id} [delete]
func DeleteUserConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		userContext := ctx.GetUser()
		if userContext == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{Code: http.StatusUnauthorized, Message: "user not found"})
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := html.UnescapeString(vars["id"])

		if err := dbService.DeleteUserConfig(ctx, userContext.ID, id); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("User config deleted successfully")
	}
}
