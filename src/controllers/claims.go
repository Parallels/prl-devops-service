package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"

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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/claims [get]
func GetClaimsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getClaimsDiag := errors.NewDiagnostics("/auth/claims [get]")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getClaimsDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getClaimsDiag, rsp.Code))
			return
		}

		dtoClaims, err := dbService.GetClaims(ctx, GetFilterHeader(r))
		if err != nil {
			rsp := models.NewFromError(err)
			getClaimsDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetClaims")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getClaimsDiag, rsp.Code))
			return
		}

		if len(dtoClaims) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.ClaimResponse, 0)
			defer r.Body.Close()
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/claims/{id} [get]
func GetClaimHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getClaimDiag := errors.NewDiagnostics("/auth/claims/{id} [get]")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getClaimDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getClaimDiag, rsp.Code))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		dtoClaim, err := dbService.GetClaim(ctx, strings.ToUpper(helpers.NormalizeString(id)))
		if err != nil {
			rsp := models.NewFromError(err)
			getClaimDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetClaim")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getClaimDiag, rsp.Code))
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
// @Failure		400				{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/claims [post]
func CreateClaimHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		createClaimDiag := errors.NewDiagnostics("/auth/claims [post]")
		var request models.ClaimRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			createClaimDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "MapRequestBody")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createClaimDiag, http.StatusBadRequest))
			return
		}
		if err := request.Validate(); err != nil {
			createClaimDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "Validate")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createClaimDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			createClaimDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createClaimDiag, rsp.Code))
			return
		}

		dtoClaim := mappers.ApiClaimToDto(request)

		claim, err := dbService.CreateClaim(ctx, dtoClaim)
		if err != nil {
			rsp := models.NewFromError(err)
			createClaimDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "CreateClaim")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(createClaimDiag, rsp.Code))
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
// @Failure		400	{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/auth/claims/{id} [delete]
func DeleteClaimHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		deleteClaimDiag := errors.NewDiagnostics("/auth/claims/{id} [delete]")
		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			deleteClaimDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteClaimDiag, rsp.Code))
			return
		}

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteClaim(ctx, id)
		if err != nil {
			rsp := models.NewFromError(err)
			deleteClaimDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "DeleteClaim")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(deleteClaimDiag, rsp.Code))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Claim deleted successfully")
	}
}
