package controllers

import (
	"encoding/json"
	"html"
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs [get]
func GetUserConfigsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getUserConfigsDiag := errors.NewDiagnostics("/user/configs")
		userContext := ctx.GetUser()
		if userContext == nil {
			getUserConfigsDiag.AddError(strconv.Itoa(http.StatusUnauthorized), "user not found", "GetUser")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigsDiag, http.StatusUnauthorized))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getUserConfigsDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigsDiag, rsp.Code))
			return
		}

		dtoConfigs, err := dbService.GetUserConfigs(ctx, userContext.ID, GetFilterHeader(r))
		if err != nil {
			rsp := models.NewFromError(err)
			getUserConfigsDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetUserConfigs")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigsDiag, rsp.Code))
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs/{id} [get]
func GetUserConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := html.UnescapeString(vars["id"])
		getUserConfigDiag := errors.NewDiagnostics("/user/configs/" + id)
		userContext := ctx.GetUser()
		if userContext == nil {
			getUserConfigDiag.AddError(strconv.Itoa(http.StatusUnauthorized), "user not found", "GetUser")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigDiag, http.StatusUnauthorized))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigDiag, rsp.Code))
			return
		}

		dtoConfig, err := dbService.GetUserConfig(ctx, userContext.ID, id)
		if err != nil {
			rsp := models.NewFromError(err)
			getUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetUserConfig")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigDiag, rsp.Code))
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs [post]
func CreateUserConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		createUserConfigDiag := errors.NewDiagnostics("/user/configs")
		userContext := ctx.GetUser()
		if userContext == nil {
			createUserConfigDiag.AddError(strconv.Itoa(http.StatusUnauthorized), "user not found", "GetUser")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createUserConfigDiag, http.StatusUnauthorized))
			return
		}

		var request models.UserConfigRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			createUserConfigDiag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createUserConfigDiag, http.StatusBadRequest))
			return
		}

		if err := request.Validate(); err != nil {
			createUserConfigDiag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createUserConfigDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			createUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createUserConfigDiag, rsp.Code))
			return
		}

		dtoConfig := mappers.UserConfigRequestToDto(userContext.ID, request)

		existing, _ := dbService.GetUserConfig(ctx, userContext.ID, request.Slug)

		dtoResult, err := dbService.UpsertUserConfig(ctx, dtoConfig)
		if err != nil {
			rsp := models.NewFromError(err)
			createUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "UpsertUserConfig")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createUserConfigDiag, rsp.Code))
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs/{id} [put]
func UpdateUserConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := html.UnescapeString(vars["id"])
		updateUserConfigDiag := errors.NewDiagnostics("/user/configs/" + id)
		userContext := ctx.GetUser()
		if userContext == nil {
			updateUserConfigDiag.AddError(strconv.Itoa(http.StatusUnauthorized), "user not found", "GetUser")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, http.StatusUnauthorized))
			return
		}

		var request models.UserConfigUpdateRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			updateUserConfigDiag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, http.StatusBadRequest))
			return
		}

		if err := request.Validate(); err != nil {
			updateUserConfigDiag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid request body: "+err.Error(), "")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			updateUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, rsp.Code))
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
				rsp := models.NewFromError(createErr)
				updateUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "UpsertUserConfig")
				ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, rsp.Code))
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
			rsp := models.NewFromError(err)
			updateUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "UpdateUserConfig")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, rsp.Code))
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		404	{object}	models.ApiErrorDiagnosticsResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/user/configs/{id} [delete]
func DeleteUserConfigHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		vars := mux.Vars(r)
		id := html.UnescapeString(vars["id"])
		deleteUserConfigDiag := errors.NewDiagnostics("/user/configs/" + id)
		userContext := ctx.GetUser()
		if userContext == nil {
			deleteUserConfigDiag.AddError(strconv.Itoa(http.StatusUnauthorized), "user not found", "GetUser")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteUserConfigDiag, http.StatusUnauthorized))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			deleteUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteUserConfigDiag, rsp.Code))
			return
		}

		if err := dbService.DeleteUserConfig(ctx, userContext.ID, id); err != nil {
			rsp := models.NewFromError(err)
			deleteUserConfigDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DeleteUserConfig")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteUserConfigDiag, rsp.Code))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("User config deleted successfully")
	}
}
