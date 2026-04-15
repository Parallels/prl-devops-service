package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerPackerTemplatesHandlers(ctx basecontext.ApiContext, version string) {
	config := config.Get()

	if !config.GetBoolKey(constants.ENABLE_PACKER_PLUGIN_ENV_VAR) {
		ctx.LogInfof("Packer plugin is disabled, skipping packer template handlers registration")
	}

	ctx.LogInfof("Registering version %s packer template handlers", version)

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/templates/packer").
		WithRequiredClaim(constants.LIST_PACKER_TEMPLATE_CLAIM).
		WithHandler(GetPackerTemplatesHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/templates/packer/{id}").
		WithRequiredClaim(constants.LIST_PACKER_TEMPLATE_CLAIM).
		WithHandler(GetPackerTemplateHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/templates/packer").
		WithRequiredClaim(constants.CREATE_PACKER_TEMPLATE_CLAIM).
		WithHandler(CreatePackerTemplateHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/templates/packer/{id}").
		WithRequiredClaim(constants.UPDATE_PACKER_TEMPLATE_CLAIM).
		WithHandler(UpdatePackerTemplateHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/templates/packer/{id}").
		WithRequiredClaim(constants.DELETE_PACKER_TEMPLATE_CLAIM).
		WithHandler(DeletePackerTemplateHandler()).
		Register()
}

// @Summary		Gets all the packer templates
// @Description	This endpoint returns all the packer templates. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
// @Tags			Packer Templates
// @Produce		json
// @Success		200	{object}	[]models.PackerTemplateResponse
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer [get]
// @deprecated
func GetPackerTemplatesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getPackerTemplatesDiag := errors.NewDiagnostics("/templates/packer")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getPackerTemplatesDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplatesDiag, rsp.Code))
			return
		}

		result, err := dbService.GetPackerTemplates(ctx, GetFilterHeader(r))
		if err != nil {
			rsp := models.NewFromError(err)
			getPackerTemplatesDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetPackerTemplates")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplatesDiag, rsp.Code))
			return
		}

		if len(result) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.PackerTemplateResponse, 0)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		response := mappers.DtoPackerTemplatesToApResponse(result)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Packer templates returned: %v", len(response))
	}
}

// @Summary		Gets a packer template
// @Description	This endpoint returns a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
// @Tags			Packer Templates
// @Produce		json
// @Param			id	path		string	true	"Packer Template ID"
// @Success		200	{object}	models.PackerTemplateResponse
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer/{id} [get]
// @deprecated
func GetPackerTemplateHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		params := mux.Vars(r)
		name := params["id"]
		getPackerTemplateDiag := errors.NewDiagnostics("/templates/packer/" + name)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getPackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplateDiag, rsp.Code))
			return
		}

		result, err := dbService.GetPackerTemplate(ctx, name)
		if err != nil {
			rsp := models.NewFromError(err)
			getPackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetPackerTemplate")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplateDiag, rsp.Code))
			return
		}

		if result == nil {
			getPackerTemplateDiag.AddError(strconv.Itoa(http.StatusNotFound), fmt.Sprintf("Packer template %v not found", name), "GetPackerTemplate")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplateDiag, http.StatusNotFound))
			return
		}

		response := mappers.DtoPackerTemplateToApResponse(*result)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Packer template returned: %v", response.ID)
	}
}

// @Summary		Creates a packer template
// @Description	This endpoint creates a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
// @Tags			Packer Templates
// @Produce		json
// @Param			createPackerTemplateRequest	body		models.CreatePackerTemplateRequest	true	"Create Packer Template Request"
// @Success		200							{object}	models.PackerTemplateResponse
// @Failure		400							{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401							{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer  [post]
// @deprecated
func CreatePackerTemplateHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.CreatePackerTemplateRequest
		createPackerTemplateDiag := errors.NewDiagnostics("/templates/packer")
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			rsp := models.NewFromError(err)
			createPackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "MapRequestBody")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createPackerTemplateDiag, rsp.Code))
			return
		}
		if diag := request.Validate(); diag.HasErrors() {
			createPackerTemplateDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createPackerTemplateDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			createPackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createPackerTemplateDiag, rsp.Code))
			return
		}

		dto := mappers.DtoPackerTemplateFromApiCreateRequest(request)
		if result, err := dbService.AddPackerTemplate(ctx, &dto); err != nil {
			rsp := models.NewFromError(err)
			createPackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "AddPackerTemplate")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createPackerTemplateDiag, rsp.Code))
			return
		} else {
			response := mappers.DtoPackerTemplateToApResponse(*result)
			w.WriteHeader(http.StatusOK)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Packer template created: %v", response.ID)
		}
	}
}

// @Summary		Updates a packer template
// @Description	This endpoint updates a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
// @Tags			Packer Templates
// @Produce		json
// @Param			createPackerTemplateRequest	body		models.CreatePackerTemplateRequest	true	"Update Packer Template Request"
// @Param			id							path		string								true	"Packer Template ID"
// @Success		200							{object}	models.PackerTemplateResponse
// @Failure		400							{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401							{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer/{id}  [PUT]
// @deprecated
func UpdatePackerTemplateHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		params := mux.Vars(r)
		id := params["id"]
		updatePackerTemplateDiag := errors.NewDiagnostics("/templates/packer/" + id)
		var request models.CreatePackerTemplateRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			rsp := models.NewFromError(err)
			updatePackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "MapRequestBody")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updatePackerTemplateDiag, rsp.Code))
			return
		}
		if diag := request.Validate(); diag.HasErrors() {
			updatePackerTemplateDiag.Append(diag)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updatePackerTemplateDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			updatePackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updatePackerTemplateDiag, rsp.Code))
			return
		}

		dto := mappers.DtoPackerTemplateFromApiCreateRequest(request)
		dto.ID = id
		if result, err := dbService.UpdatePackerTemplate(ctx, &dto); err != nil {
			rsp := models.NewFromError(err)
			updatePackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "UpdatePackerTemplate")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updatePackerTemplateDiag, rsp.Code))
			return
		} else {
			response := mappers.DtoPackerTemplateToApResponse(*result)
			w.WriteHeader(http.StatusOK)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Packer template updated: %v", response.ID)
		}
	}
}

// @Summary		Deletes a packer template
// @Description	This endpoint deletes a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
// @Tags			Packer Templates
// @Produce		json
// @Param			id	path	string	true	"Packer Template ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer/{id}  [DELETE]
// @deprecated
func DeletePackerTemplateHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		params := mux.Vars(r)
		id := params["id"]
		deletePackerTemplateDiag := errors.NewDiagnostics("/templates/packer/" + id)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			deletePackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deletePackerTemplateDiag, rsp.Code))
			return
		}

		if err := dbService.DeletePackerTemplate(ctx, id); err != nil {
			rsp := models.NewFromError(err)
			deletePackerTemplateDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DeletePackerTemplate")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deletePackerTemplateDiag, rsp.Code))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Packer template deleted: %v", id)
	}
}
