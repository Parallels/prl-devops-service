package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	bruteforceguard "github.com/Parallels/pd-api-service/security/brute_force_guard"
	"github.com/Parallels/pd-api-service/security/jwt"
	"github.com/Parallels/pd-api-service/security/password"
	"github.com/Parallels/pd-api-service/serviceprovider"

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
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Router			/v1/auth/token [post]
func GetTokenHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)

		var request models.LoginRequest
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

		user, err := dbService.GetUser(ctx, request.Email)
		if err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid User or Password",
				Code:    http.StatusUnauthorized,
			})
			return
		}

		if user == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid User or Password",
				Code:    http.StatusUnauthorized,
			})
			return
		}

		bruteForceSvc := bruteforceguard.Get()

		passwdSvc := password.Get()
		if err := passwdSvc.Compare(request.Password, user.ID, user.Password); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid User or Password",
				Code:    http.StatusUnauthorized,
			})

			if diag := bruteForceSvc.Process(user.ID, false, "Invalid Password"); diag.HasErrors() {
				ctx.LogErrorf("Error processing brute force guard: %v", diag)
			}
			return
		}

		userRoles := make([]string, 0)
		userClaims := make([]string, 0)
		for _, userRole := range user.Roles {
			userRoles = append(userRoles, userRole.Name)
		}
		for _, userClaim := range user.Claims {
			userClaims = append(userClaims, userClaim.Name)
		}

		claims := map[string]interface{}{
			"email":  request.Email,
			"uid":    user.ID,
			"roles":  userRoles,
			"claims": userClaims,
		}
		tokenSvc := jwt.Get()
		tokenStr, err := tokenSvc.Sign(claims)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 401))
			if diag := bruteForceSvc.Process(user.ID, false, err.Error()); diag.HasErrors() {
				ctx.LogErrorf("Error processing brute force guard: %v", diag)
			}
			return
		}
		token, err := tokenSvc.Parse(tokenStr)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 401))
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
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Router			/v1/auth/token/validate [post]
func ValidateTokenHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)

		var request models.ValidateTokenRequest
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

		tokenSvc := jwt.Get()
		token, err := tokenSvc.Parse(request.Token)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 401))
			return
		}

		isValid, err := token.Valid()
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 401))
			return
		}

		if !isValid {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 401))
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
