package controllers

import (
	"Parallels/pd-api-service/mappers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

// LoginUser is a public function that logs in a user
func GetPackerTemplatesController() restapi.Controller {
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

func GetPackerTemplateController() restapi.Controller {
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

func CreatePackerTemplateController() restapi.Controller {
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

func UpdatePackerTemplateController() restapi.Controller {
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

func DeletePackerTemplateController() restapi.Controller {
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
