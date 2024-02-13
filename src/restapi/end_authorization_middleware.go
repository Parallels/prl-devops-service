package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
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
				_ = json.NewEncoder(w).Encode(response)
				baseCtx.LogInfo("Authorization layer finished with error")
				return
			}
		})
	}
}
