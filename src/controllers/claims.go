package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerClaimsHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Claims handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).WithPath("/auth/claims").
		WithRequiredClaim(constants.LIST_CLAIM_CLAIM).
		WithHandler(GetClaimsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/auth/claims/{id}").
		WithRequiredClaim(constants.LIST_CLAIM_CLAIM).
		WithHandler(GetClaimHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/claims").
		WithRequiredClaim(constants.CREATE_CLAIM_CLAIM).
		WithHandler(CreateClaimHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/auth/claims/{id}").
		WithRequiredClaim(constants.DELETE_CLAIM_CLAIM).
		WithHandler(DeleteClaimHandler()).
		Register()
}

// @Summary		Gets all the claims
// @Description	This endpoint returns all the claims
// @Tags			Claims
// @Produce		json
// @Success		200	{object}	[]models.ClaimResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/claims [get]
func GetClaimsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		dtoClaims, err := dbService.GetClaims(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(dtoClaims) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.ClaimResponse, 0)
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Claims returned: %v", len(response))
			return
		}

		result := mappers.DtoClaimsToApi(dtoClaims)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Claims returned successfully")
	}
}

// @Summary		Gets a claim
// @Description	This endpoint returns a claim
// @Tags			Claims
// @Produce		json
// @Param			id	path		string	true	"Claim ID"
// @Success		200	{object}	models.ClaimResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/claims/{id} [get]
func GetClaimHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dtoClaim, err := dbService.GetClaim(ctx, strings.ToUpper(helpers.NormalizeString(id)))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 404))
			return
		}

		response := mappers.DtoClaimToApi(*dtoClaim)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Claim returned successfully")
	}
}

// @Summary		Creates a claim
// @Description	This endpoint creates a claim
// @Tags			Claims
// @Produce		json
// @Param			claimRequest	body		models.ClaimRequest	true	"Claim Request"
// @Success		200				{object}	models.ClaimResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/claims [post]
func CreateClaimHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.ClaimRequest
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

		dtoClaim := mappers.ApiClaimToDto(request)

		claim, err := dbService.CreateClaim(ctx, dtoClaim)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoClaimToApi(*claim)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("Claim created successfully")
	}
}

// @Summary		Delete a claim
// @Description	This endpoint Deletes a claim
// @Tags			Claims
// @Produce		json
// @Param			id	path	string	true	"Claim ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/claims/{id} [delete]
func DeleteClaimHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteClaim(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Claim deleted successfully")
	}
}
