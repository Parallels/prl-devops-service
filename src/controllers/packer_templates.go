package controllers

import (
	"encoding/json"
	"fmt"
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

func registerPackerTemplatesHandlers(ctx basecontext.ApiContext, version string) {
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
// @Description	This endpoint returns all the packer templates
// @Tags			Packer Templates
// @Produce		json
// @Success		200	{object}	[]models.PackerTemplateResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer [get]
func GetPackerTemplatesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		result, err := dbService.GetPackerTemplates(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(result) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.PackerTemplateResponse, 0)
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
// @Description	This endpoint returns a packer template
// @Tags			Packer Templates
// @Produce		json
// @Param			id	path		string	true	"Packer Template ID"
// @Success		200	{object}	models.PackerTemplateResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer/{id} [get]
func GetPackerTemplateHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		params := mux.Vars(r)
		name := params["id"]

		result, err := dbService.GetPackerTemplate(ctx, name)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if result == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Packer template %v not found", name),
				Code:    http.StatusNotFound,
			})
			return
		}

		response := mappers.DtoPackerTemplateToApResponse(*result)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Packer template returned: %v", response.ID)
	}
}

// @Summary		Creates a packer template
// @Description	This endpoint creates a packer template
// @Tags			Packer Templates
// @Produce		json
// @Param			createPackerTemplateRequest	body		models.CreatePackerTemplateRequest	true	"Create Packer Template Request"
// @Success		200							{object}	models.PackerTemplateResponse
// @Failure		400							{object}	models.ApiErrorResponse
// @Failure		401							{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer  [post]
func CreatePackerTemplateHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.CreatePackerTemplateRequest
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

		dto := mappers.DtoPackerTemplateFromApiCreateRequest(request)
		if result, err := dbService.AddPackerTemplate(ctx, &dto); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		} else {
			response := mappers.DtoPackerTemplateToApResponse(*result)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Packer template created: %v", response.ID)
		}
	}
}

// @Summary		Updates a packer template
// @Description	This endpoint updates a packer template
// @Tags			Packer Templates
// @Produce		json
// @Param			createPackerTemplateRequest	body		models.CreatePackerTemplateRequest	true	"Update Packer Template Request"
// @Param			id							path		string								true	"Packer Template ID"
// @Success		200							{object}	models.PackerTemplateResponse
// @Failure		400							{object}	models.ApiErrorResponse
// @Failure		401							{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer/{id}  [PUT]
func UpdatePackerTemplateHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.CreatePackerTemplateRequest
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

		params := mux.Vars(r)
		id := params["id"]

		dto := mappers.DtoPackerTemplateFromApiCreateRequest(request)
		dto.ID = id
		if result, err := dbService.UpdatePackerTemplate(ctx, &dto); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		} else {
			response := mappers.DtoPackerTemplateToApResponse(*result)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Packer template updated: %v", response.ID)
		}
	}
}

// @Summary		Deletes a packer template
// @Description	This endpoint deletes a packer template
// @Tags			Packer Templates
// @Produce		json
// @Param			id	path	string	true	"Packer Template ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/templates/packer/{id}  [DELETE]
func DeletePackerTemplateHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		params := mux.Vars(r)
		id := params["id"]

		if err := dbService.DeletePackerTemplate(ctx, id); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Packer template deleted: %v", id)
	}
}
