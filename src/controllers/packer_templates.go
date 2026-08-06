package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
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
		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			getPackerTemplatesDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplatesDiag, http.StatusInternalServerError))
			return
		}

		// Build query from URL query params (e.g., ?type=bool&order_by=name&order=desc)
		queryBuilder := filters.NewQueryBuilder(r.URL.RawQuery)

		// Access store directly - NO domain layer, NO convenience methods
		store := dbService.Stores().PackerTemplate()
		result, storeDiag := store.Find(*ctx, queryBuilder)
		if storeDiag != nil {
			getPackerTemplatesDiag.AddError(strconv.Itoa(http.StatusInternalServerError), storeDiag.GetSummary(), "Store.Find")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplatesDiag, http.StatusInternalServerError))
			return
		}

		if result == nil || result.Total == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.PackerTemplateResponse, 0)
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		response := mappers.GormPackerTemplatesQueryResponseToResponse(result)
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
		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			getPackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		// Access store directly - NO domain layer, NO convenience methods
		store := dbService.Stores().PackerTemplate()
		result, storeDiag := store.Get(*ctx, name)
		if storeDiag != nil {
			getPackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), storeDiag.GetSummary(), "Store.Get")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		if result == nil {
			getPackerTemplateDiag.AddError(strconv.Itoa(http.StatusNotFound), fmt.Sprintf("Packer template %v not found", name), "GetPackerTemplate", nil)
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getPackerTemplateDiag, http.StatusNotFound))
			return
		}

		response := mappers.GormPackerTemplateDtoToResponse(*result)
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

		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			createPackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createPackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		dto := mappers.GormPackerTemplateRequestToDto(request)

		// Access store directly - NO domain layer, NO convenience methods
		store := dbService.Stores().PackerTemplate()
		result, storeDiag := store.Create(*ctx, &dto)
		if storeDiag != nil {
			createPackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), storeDiag.GetSummary(), "Store.Create")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createPackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		response := mappers.GormPackerTemplateDtoToResponse(*result)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Packer template created: %v", response.ID)
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

		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			updatePackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updatePackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		dto := mappers.GormPackerTemplateRequestToDto(request)
		dto.ID = id

		// Access store directly - NO domain layer, NO convenience methods
		store := dbService.Stores().PackerTemplate()
		storeDiag := store.Update(*ctx, &dto)
		if storeDiag != nil {
			updatePackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), storeDiag.GetSummary(), "Store.Update")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updatePackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		result, storeDiagGet := store.Get(*ctx, id)
		if storeDiagGet != nil {
			updatePackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), storeDiagGet.GetSummary(), "Store.Get")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(updatePackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		response := mappers.GormPackerTemplateDtoToResponse(*result)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Packer template updated: %v", response.ID)
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
		dbService, diag := serviceprovider.GetDatabaseService(ctx)
		if diag != nil {
			deletePackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), diag.GetSummary(), "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deletePackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		// Access store directly - NO domain layer, NO convenience methods
		store := dbService.Stores().PackerTemplate()
		if storeDiag := store.Delete(*ctx, id); storeDiag != nil {
			deletePackerTemplateDiag.AddError(strconv.Itoa(http.StatusInternalServerError), storeDiag.GetSummary(), "Store.Delete")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deletePackerTemplateDiag, http.StatusInternalServerError))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Packer template deleted: %v", id)
	}
}
