package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	bruteforceguard "github.com/Parallels/prl-devops-service/security/brute_force_guard"
	"github.com/Parallels/prl-devops-service/security/jwt"
	"github.com/Parallels/prl-devops-service/security/password"
	"github.com/Parallels/prl-devops-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
)

func registerAuthorizationHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Authorization handlers", version)
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/token").
		WithHandler(GetTokenHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/token/validate").
		WithHandler(ValidateTokenHandler()).
		Register()
}

// @Summary		Generates a token
// @Description	This endpoint generates a token
// @Tags			Authorization
// @Produce		json
// @Param			login	body		models.LoginRequest	true	"Body"
// @Success		200		{object}	models.LoginResponse
// @Failure		400		{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401		{object}	models.ApiErrorDiagnosticsResponse
// @Router			/v1/auth/token [post]
func GetTokenHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		getTokenDiag := errors.NewDiagnostics("/auth/token")
		var request models.LoginRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			getTokenDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "MapRequestBody")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getTokenDiag, http.StatusBadRequest))
			return
		}
		if err := request.Validate(); err != nil {
			getTokenDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "Validate")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getTokenDiag, http.StatusBadRequest))
			return
		}

		dbService, err := serviceprovider.GetDatabaseService(ctx)
		if err != nil {
			rsp := models.NewFromError(err)
			getTokenDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "ServiceProvider")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getTokenDiag, rsp.Code))
			return
		}

		user, err := dbService.GetUser(ctx, request.Email)
		if err != nil {
			rsp := models.NewFromError(err)
			getTokenDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "GetUser")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getTokenDiag, rsp.Code))
			return
		}

		if user == nil {
			getTokenDiag.AddError(strconv.Itoa(http.StatusUnauthorized), "Invalid User or Password", "GetUser")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getTokenDiag, http.StatusUnauthorized))
			return
		}

		bruteForceSvc := bruteforceguard.Get()

		passwdSvc := password.Get()
		if err := passwdSvc.Compare(request.Password, user.ID, user.Password); err != nil {
			rsp := models.NewFromError(err)
			getTokenDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Compare")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getTokenDiag, rsp.Code))

			if diag := bruteForceSvc.Process(user.ID, false, "Invalid Password"); diag.HasErrors() {
				ctx.LogErrorf("Error processing brute force guard: %v", diag)
			}
			return
		}

		userRoles := make([]string, 0)
		for _, userRole := range user.Roles {
			userRoles = append(userRoles, userRole.Name)
		}
		// Use effective claims (direct + role-inherited, deduplicated) so the JWT
		// reflects the user's full permission set.
		userClaims := mappers.ComputeEffectiveClaimIDs(*user)

		claims := map[string]interface{}{
			"email":    user.Email,
			"username": user.Name,
			"uid":      user.ID,
			"roles":    userRoles,
			"claims":   userClaims,
		}
		tokenSvc := jwt.Get()
		tokenStr, err := tokenSvc.Sign(claims)
		if err != nil {
			rsp := models.NewFromError(err)
			getTokenDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Sign")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getTokenDiag, rsp.Code))
			if diag := bruteForceSvc.Process(user.ID, false, err.Error()); diag.HasErrors() {
				ctx.LogErrorf("Error processing brute force guard: %v", diag)
			}
			return
		}
		token, err := tokenSvc.Parse(tokenStr)
		if err != nil {
			rsp := models.NewFromError(err)
			getTokenDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Parse")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(getTokenDiag, rsp.Code))
			if diag := bruteForceSvc.Process(user.ID, false, err.Error()); diag.HasErrors() {
				ctx.LogErrorf("Error processing brute force guard: %v", diag)
			}
			return
		}

		response := models.LoginResponse{
			Token:     tokenStr,
			Email:     request.Email,
			ExpiresAt: int64(token.Claims["exp"].(float64)),
		}

		if diag := bruteForceSvc.Process(user.ID, true, "Success"); diag.HasErrors() {
			ctx.LogErrorf("Error processing brute force guard: %v", diag)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfof("User %s logged in", request.Email)
	}
}

// @Summary		Validates a token
// @Description	This endpoint validates a token
// @Tags			Authorization
// @Produce		json
// @Param			tokenRequest	body		models.ValidateTokenRequest	true	"Body"
// @Success		200				{object}	models.ValidateTokenResponse
// @Failure		400				{object}	models.ApiErrorDiagnosticsResponse
// @Failure		401				{object}	models.ApiErrorDiagnosticsResponse
// @Router			/v1/auth/token/validate [post]
func ValidateTokenHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		validateTokenDiag := errors.NewDiagnostics("/auth/token/validate")
		var request models.ValidateTokenRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			validateTokenDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "MapRequestBody")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(validateTokenDiag, http.StatusBadRequest))
			return
		}
		if err := request.Validate(); err != nil {
			validateTokenDiag.AddError(strconv.Itoa(http.StatusBadRequest), "Invalid request body: "+err.Error(), "Validate")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(validateTokenDiag, http.StatusBadRequest))
			return
		}

		tokenSvc := jwt.Get()
		token, err := tokenSvc.Parse(request.Token)
		if err != nil {
			rsp := models.NewFromError(err)
			validateTokenDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Parse")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(validateTokenDiag, rsp.Code))
			return
		}

		isValid, err := token.Valid()
		if err != nil {
			rsp := models.NewFromError(err)
			validateTokenDiag.AddError(strconv.Itoa(rsp.Code), rsp.Message, "Valid")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(validateTokenDiag, rsp.Code))
			return
		}

		if !isValid {
			validateTokenDiag.AddError(strconv.Itoa(http.StatusUnauthorized), "Invalid token", "Valid")
			ReturnApiErrorWithDiagnostics(ctx, w, models.NewDiagnosticsWithCode(validateTokenDiag, http.StatusUnauthorized))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(models.ValidateTokenResponse{
			Valid: true,
		})
		email, _ := token.GetEmail()
		ctx.LogInfof("Token for user %s is valid", email)
	}
}
