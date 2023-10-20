package restapi

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/services"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

type ApiKeyHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func ApiKeyAuthorizationMiddlewareAdapter(roles []string, claims []string) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var authorizationContext *AuthorizationContext
			authCtxFromRequest := r.Context().Value(constants.AUTHORIZATION_CONTEXT_KEY)
			if authCtxFromRequest != nil {
				authorizationContext = authCtxFromRequest.(*AuthorizationContext)
			} else {
				authorizationContext = InitAuthorizationContext()
			}

			// If the authorization context is already authorized we will skip this middleware
			if authorizationContext.IsAuthorized || HasAuthorizationHeader(r) {
				common.Logger.Info("%sNo Api Key was found in the request, skipping", common.Logger.GetRequestPrefix(r, false))
				next.ServeHTTP(w, r)
				return
			}

			common.Logger.Info("%sApiKey Authorization layer started", common.Logger.GetRequestPrefix(r, false))
			authError := models.OAuthErrorResponse{
				Error:            models.OAuthUnauthorizedClient,
				ErrorDescription: "The Api Key is not valid",
			}

			apiKey, err := extractApiKey(r.Header)
			if err != nil {
				authError.ErrorDescription = err.Error()
				authorizationContext.AuthorizationError = &authError
				common.Logger.Info("%sNo Api Key was found in the request, skipping", common.Logger.GetRequestPrefix(r, false))
				next.ServeHTTP(w, r)
				return
			}
			isValid := true
			db := services.GetServices().JsonDatabase
			db.Connect()
			dbApiKey, err := db.GetApiKey(apiKey.Key)

			if err != nil || dbApiKey == nil {
				isValid = false
			}

			if isValid {
				if dbApiKey.Revoked {
					isValid = false
					authError.ErrorDescription = "Api Key has been revoked"
				}
			}
			if isValid {
				hashedSecret := helpers.Sha256Hash(apiKey.Value)
				if dbApiKey.Secret != hashedSecret {
					isValid = false
					authError.ErrorDescription = "Api Key is not Valid"
				}
			}

			if !isValid {
				common.Logger.Error("%sThe Api Key is not valid", common.Logger.GetRequestPrefix(r, false))
				authorizationContext.IsAuthorized = false
				authorizationContext.AuthorizationError = &authError

				ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			authorizationContext.IsAuthorized = true
			authorizationContext.IsMicroService = true
			authorizationContext.AuthorizedBy = "ApiKeyAuthorization"
			authorizationContext.AuthorizationError = nil
			ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
			common.Logger.Info("%sApiKey Authorization layer finished", common.Logger.GetRequestPrefix(r, false))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractApiKey(headers http.Header) (*ApiKeyHeader, error) {
	authHeader := headers.Get("X-Api-Key")
	if authHeader == "" {
		err := errors.New("No Api Key was found in the request")
		return nil, err
	}

	decodedKey, err := base64.StdEncoding.DecodeString(authHeader)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(decodedKey), ":")
	if len(parts) != 2 {
		err := errors.New("The Api Key is not in the correct format")
		return nil, err
	}

	return &ApiKeyHeader{
		Key:   parts[0],
		Value: parts[1],
	}, nil
}
