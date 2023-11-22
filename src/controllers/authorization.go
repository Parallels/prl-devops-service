package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/dgrijalva/jwt-go"
)

func registerAuthorizationHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfo("Registering version %s Authorization handlers", version)
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/auth/token").
		WithHandler(GetTokenHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
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
		cfg := config.NewConfig()
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
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if user == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: data.ErrUserNotFound.Error(),
				Code:    http.StatusUnauthorized,
			})
			return
		}

		// Hash the password with SHA-256
		hashedPassword := helpers.Sha256Hash(request.Password)
		if hashedPassword != user.Password {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid Password",
				Code:    http.StatusUnauthorized,
			})
			return
		}

		roles := make([]string, 0)
		claims := make([]string, 0)
		for _, role := range user.Roles {
			roles = append(roles, role.Name)
		}
		for _, claim := range user.Claims {
			claims = append(claims, claim.Name)
		}

		expiresAt := time.Now().Add(time.Minute * time.Duration(cfg.GetTokenDurationMinutes())).Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email":  request.Email,
			"roles":  roles,
			"claims": claims,
			"exp":    expiresAt,
		})

		// We either signing the token with the HMAC secret or the secret from the config
		var key []byte
		if cfg.GetHmacSecret() == "" {
			key = []byte(cfg.GetHmacSecret())
		} else {
			key = []byte(serviceprovider.Get().HardwareSecret)
		}

		tokenString, err := token.SignedString(key)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 401))
			return
		}

		response := models.LoginResponse{
			Token:     tokenString,
			Email:     request.Email,
			ExpiresAt: expiresAt,
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
		ctx.LogInfo("User %s logged in", request.Email)
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
		cfg := config.NewConfig()
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

		token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			// We either signing the token with the HMAC secret or the secret from the config
			var key []byte
			if cfg.GetHmacSecret() == "" {
				key = []byte(cfg.GetHmacSecret())
			} else {
				key = []byte(serviceprovider.Get().HardwareSecret)
			}
			return key, nil
		})

		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 401))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(models.ValidateTokenResponse{
				Valid: true,
			})
			ctx.LogInfo("Token for user %s is valid", claims["email"])
			return
		} else {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, 401))
			ctx.LogError("Token is invalid")
			return
		}
	}
}
