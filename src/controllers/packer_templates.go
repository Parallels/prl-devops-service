package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

//	@Summary		Gets all the packer templates
//	@Description	This endpoint returns all the packer templates
//	@Tags			Packer Templates
//	@Produce		json
//	@Success		200	{object}	[]models.PackerTemplateResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/templates/packer [get]
func GetPackerTemplatesController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		result, err := dbService.GetPackerTemplates(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(result) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.PackerTemplateResponse, 0)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := mappers.DtoPackerTemplatesToApResponse(result)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Packer templates returned: %v", len(response))
	}
}

//	@Summary		Gets a packer template
//	@Description	This endpoint returns a packer template
//	@Tags			Packer Templates
//	@Produce		json
//	@Param			id	path		string	true	"Packer Template ID"
//	@Success		200	{object}	models.PackerTemplateResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/templates/packer/{id} [get]
func GetPackerTemplateController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

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
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Packer template returned: %v", response.ID)
	}
}

//	@Summary		Creates a packer template
//	@Description	This endpoint creates a packer template
//	@Tags			Packer Templates
//	@Produce		json
//	@Param			createPackerTemplateRequest	body		models.CreatePackerTemplateRequest	true	"Create Packer Template Request"
//	@Success		200							{object}	models.PackerTemplateResponse
//	@Failure		400							{object}	models.ApiErrorResponse
//	@Failure		401							{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/templates/packer  [post]
func CreatePackerTemplateController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.CreatePackerTemplateRequest
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

		defer dbService.Disconnect(ctx)

		dto := mappers.DtoPackerTemplateFromApiCreateRequest(request)
		if result, err := dbService.AddPackerTemplate(ctx, &dto); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		} else {
			response := mappers.DtoPackerTemplateToApResponse(*result)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Packer template created: %v", response.ID)
		}
	}
}

//	@Summary		Updates a packer template
//	@Description	This endpoint updates a packer template
//	@Tags			Packer Templates
//	@Produce		json
//	@Param			createPackerTemplateRequest	body		models.CreatePackerTemplateRequest	true	"Update Packer Template Request"
//	@Param			id							path		string								true	"Packer Template ID"
//	@Success		200							{object}	models.PackerTemplateResponse
//	@Failure		400							{object}	models.ApiErrorResponse
//	@Failure		401							{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/templates/packer/{id}  [PUT]
func UpdatePackerTemplateController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.CreatePackerTemplateRequest
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

		defer dbService.Disconnect(ctx)

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
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Packer template updated: %v", response.ID)
		}
	}
}

//	@Summary		Deletes a packer template
//	@Description	This endpoint deletes a packer template
//	@Tags			Packer Templates
//	@Produce		json
//	@Param			id	path	string	true	"Packer Template ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/templates/packer/{id}  [DELETE]
func DeletePackerTemplateController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		params := mux.Vars(r)
		id := params["id"]

		if err := dbService.DeletePackerTemplate(ctx, id); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Packer template deleted: %v", id)
	}
}
