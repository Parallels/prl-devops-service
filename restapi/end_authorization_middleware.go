package restapi

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/models"
	"encoding/json"
	"net/http"
)

func EndAuthorizationMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorizationContext := r.Context().Value(constants.AUTHORIZATION_CONTEXT_KEY)
			if authorizationContext != nil {
				auth := authorizationContext.(*basecontext.AuthorizationContext)
				if !auth.IsAuthorized {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(auth.AuthorizationError)
					common.Logger.Info("%sAuthorization Layer Finished", common.Logger.GetRequestPrefix(r, false))
					return
				}

				next.ServeHTTP(w, r)
				common.Logger.Info("%sAuthorization Layer Finished", common.Logger.GetRequestPrefix(r, false))
			} else {
				response := models.OAuthErrorResponse{
					Error:            models.OAuthUnauthorizedClient,
					ErrorDescription: "no authorization context was found in the request",
				}

				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response)
				return
			}
		})
	}
}
