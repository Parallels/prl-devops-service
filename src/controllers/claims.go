package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

//	@Summary		Gets all the claims
//	@Description	This endpoint returns all the claims
//	@Tags			Claims
//	@Produce		json
//	@Success		200	{object}	[]models.ClaimResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/claims [get]
func GetClaimsController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		dtoClaims, err := dbService.GetClaims(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(dtoClaims) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.ClaimResponse, 0)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Claims returned: %v", len(response))
			return
		}

		result := mappers.DtoClaimsToApi(dtoClaims)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Claims returned successfully")
	}
}

//	@Summary		Gets a claim
//	@Description	This endpoint returns a claim
//	@Tags			Claims
//	@Produce		json
//	@Param			id	path		string	true	"Claim ID"
//	@Success		200	{object}	models.ClaimResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/claims/{id} [get]
func GetClaimController() restapi.ControllerHandler {
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

		dtoClaim, err := dbService.GetClaim(ctx, strings.ToUpper(helpers.NormalizeString(id)))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := mappers.DtoClaimToApi(*dtoClaim)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Claim returned successfully")
	}
}

//	@Summary		Creates a claim
//	@Description	This endpoint creates a claim
//	@Tags			Claims
//	@Produce		json
//	@Param			claimRequest	body		models.ClaimRequest	true	"Claim Request"
//	@Success		200				{object}	models.ClaimResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/claims [post]
func CreateClaimController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.ClaimRequest
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

		dtoClaim := mappers.ApiClaimToDto(request)

		claim, err := dbService.CreateClaim(ctx, dtoClaim)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoClaimToApi(*claim)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("Claim created successfully")
	}
}

//	@Summary		Delete a claim
//	@Description	This endpoint Deletes a claim
//	@Tags			Claims
//	@Produce		json
//	@Param			id	path	string	true	"Claim ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/auth/claims/{id} [delete]
func DeleteClaimController() restapi.ControllerHandler {
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

		err = dbService.DeleteClaim(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Claim deleted successfully")
	}
}
