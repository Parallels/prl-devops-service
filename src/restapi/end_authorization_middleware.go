package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/models"
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
					_ = json.NewEncoder(w).Encode(auth.AuthorizationError)
					baseCtx.LogInfof("Authorization layer finished with error")
					return
				}

				next.ServeHTTP(w, r)
				baseCtx.LogInfof("Authorization layer finished with success")
			} else {
				response := models.OAuthErrorResponse{
					Error:            models.OAuthUnauthorizedClient,
					ErrorDescription: "no authorization context was found in the request",
				}

				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(response)
				baseCtx.LogInfof("Authorization layer finished with error")
				return
			}
		})
	}
}
