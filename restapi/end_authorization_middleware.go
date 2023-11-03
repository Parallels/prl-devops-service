package restapi

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/models"
	"encoding/json"
	"net/http"
)

func EndAuthorizationMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			baseCtx := basecontext.NewBaseContextFromRequest(r)
			authorizationContext := r.Context().Value(constants.AUTHORIZATION_CONTEXT_KEY)
			if authorizationContext != nil {
				auth := authorizationContext.(*basecontext.AuthorizationContext)
				if !auth.IsAuthorized {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(auth.AuthorizationError)
					baseCtx.LogInfo("Authorization layer finished with error")
					return
				}

				next.ServeHTTP(w, r)
				baseCtx.LogInfo("Authorization layer finished with success")
			} else {
				response := models.OAuthErrorResponse{
					Error:            models.OAuthUnauthorizedClient,
					ErrorDescription: "no authorization context was found in the request",
				}

				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response)
				baseCtx.LogInfo("Authorization layer finished with error")
				return
			}
		})
	}
}
