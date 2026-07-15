package controllers

import (
	"encoding/json"
	"html"
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/filters"
	db_models "github.com/Parallels/prl-devops-service/database/models"
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

		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			getUserConfigsDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigsDiag, http.StatusInternalServerError))
			return
		}

		// Build query from URL query params (e.g., ?type=bool&order_by=name&order=desc)
		queryBuilder := filters.NewQueryBuilder(r.URL.RawQuery)

		// Access store directly - NO domain layer, NO convenience methods
		store := dbService.Stores().UserConfig()
		dtoConfigs, diag := store.Find(*ctx, userContext.ID, queryBuilder)
		if diag != nil {
			getUserConfigsDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "Store.Find")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigsDiag, http.StatusInternalServerError))
			return
		}

		result := mappers.GormUserConfigsQueryResponseToResponse(dtoConfigs)

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

		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			getUserConfigDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigDiag, http.StatusInternalServerError))
			return
		}

		// Access store directly
		store := dbService.Stores().UserConfig()
		dtoConfig, diag := store.Get(*ctx, userContext.ID, id)
		if diag != nil {
			getUserConfigDiag.AddError(strconv.Itoa(http.StatusNotFound), diag.GetSummary(), "Store.Get")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getUserConfigDiag, http.StatusNotFound))
			return
		}

		response := mappers.GormUserConfigDtoToResponse(*dtoConfig)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("User config returned successfully")
	}
}

// @Summary		Creates a user config
// @Description	This endpoint creates a configuration entry for the authenticated user
// @Tags			User Configs
// @Produce		json
// @Param			userConfig	body		models.UserConfigRequest	true	"Body"
// @Success		201			{object}	models.UserConfigResponse
// @Failure		400			{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401			{object}	models.ApiErrorDiagnosticsResponse
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

		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			createUserConfigDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createUserConfigDiag, http.StatusInternalServerError))
			return
		}

		store := dbService.Stores().UserConfig()

		// Check if config exists to determine create vs update
		existing, _ := store.Get(*ctx, userContext.ID, request.Slug)

		var response models.UserConfigResponse
		if existing != nil {
			// Update existing config
			existing.Name = request.Name
			existing.Type = db_models.UserConfigValueType(request.Type)
			existing.Value = request.Value

			diag := store.Update(*ctx, existing)
			if diag != nil {
				createUserConfigDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "Store.Update")
				ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createUserConfigDiag, http.StatusInternalServerError))
				return
			}
			response = mappers.GormUserConfigDtoToResponse(*existing)
			w.WriteHeader(http.StatusOK)
			ctx.LogInfof("User config updated successfully (upsert)")
		} else {
			// Create new config
			dtoConfig := mappers.GormUserConfigRequestToDto(userContext.ID, request)
			dtoResult, diag := store.Create(*ctx, &dtoConfig)
			if diag != nil {
				createUserConfigDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "Store.Create")
				ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createUserConfigDiag, http.StatusInternalServerError))
				return
			}
			response = mappers.GormUserConfigDtoToResponse(*dtoResult)
			w.WriteHeader(http.StatusCreated)
			ctx.LogInfof("User config created successfully")
		}
		_ = json.NewEncoder(w).Encode(response)
	}
}

// @Summary		Updates a user config
// @Description	This endpoint updates a configuration entry for the authenticated user
// @Tags			User Configs
// @Param			id			path	string							true	"Config ID or Slug"
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

		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			updateUserConfigDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, http.StatusInternalServerError))
			return
		}

		store := dbService.Stores().UserConfig()
		existing, diag := store.Get(*ctx, userContext.ID, id)
		if diag != nil {
			// Not found — create a new record using the path id as slug.
			name := request.Name
			if name == "" {
				name = id
			}
			cfgType := db_models.UserConfigValueType(request.Type)
			if cfgType == "" {
				cfgType = db_models.UserConfigValueTypeString
			}
			newCfg := db_models.UserConfig{
				UserID: userContext.ID,
				Slug:   id,
				Name:   name,
				Type:   cfgType,
				Value:  request.Value,
			}
			dtoResult, createDiag := store.Create(*ctx, &newCfg)
			if createDiag != nil {
				updateUserConfigDiag.AddError(strconv.Itoa(http.StatusInternalServerError), createDiag.GetSummary(), "Store.Create")
				ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, http.StatusInternalServerError))
				return
			}
			response := mappers.GormUserConfigDtoToResponse(*dtoResult)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("User config created successfully (upsert)")
			return
		}

		if request.Name != "" {
			existing.Name = request.Name
		}
		if request.Type != "" {
			existing.Type = db_models.UserConfigValueType(request.Type)
		}
		if request.Value != "" {
			existing.Value = request.Value
		}

		diag = store.Update(*ctx, existing)
		if diag != nil {
			updateUserConfigDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "Store.Update")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updateUserConfigDiag, http.StatusInternalServerError))
			return
		}

		response := mappers.GormUserConfigDtoToResponse(*existing)

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

		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			deleteUserConfigDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteUserConfigDiag, http.StatusInternalServerError))
			return
		}

		store := dbService.Stores().UserConfig()
		if diag := store.Delete(*ctx, userContext.ID, id); diag != nil {
			deleteUserConfigDiag.AddError(strconv.Itoa(http.StatusNotFound), diag.GetSummary(), "Store.Delete")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteUserConfigDiag, http.StatusNotFound))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("User config deleted successfully")
	}
}
