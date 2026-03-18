package restapi

import (
	"context"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

// EnrollmentTokenAuthorizationMiddlewareAdapter authorises a request that carries a
// valid X-Enrollment-Token header.  It is intended to be added as an ExtraAdapter on
// specific endpoints (e.g. POST /orchestrator/hosts) so that a freshly-deployed agent
// can register itself without requiring a long-lived credential.
//
// If the header is absent the adapter is a no-op and the normal auth chain continues.
// If the header is present but the token is invalid/expired/used, an auth error is set
// and the request is denied regardless of other credentials.
func EnrollmentTokenAuthorizationMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenValue := r.Header.Get(constants.ENROLLMENT_TOKEN_HEADER)
			if tokenValue == "" {
				// No enrollment token – let the normal auth chain decide.
				next.ServeHTTP(w, r)
				return
			}

			baseCtx := basecontext.NewBaseContextFromRequest(r)
			baseCtx.LogInfof("EnrollmentToken Authorization layer started")

			authorizationContext, _ := r.Context().Value(constants.AUTHORIZATION_CONTEXT_KEY).(*basecontext.AuthorizationContext)
			if authorizationContext == nil {
				authorizationContext = basecontext.InitAuthorizationContext()
			}

			// If the request is already authorized via Bearer/ApiKey, leave it alone.
			if authorizationContext.IsAuthorized {
				next.ServeHTTP(w, r)
				return
			}

			authError := models.OAuthErrorResponse{
				Error:            models.OAuthUnauthorizedClient,
				ErrorDescription: "The enrollment token is not valid",
			}

			db := serviceprovider.Get().JsonDatabase
			if err := db.Connect(baseCtx); err != nil {
				authError.ErrorDescription = "database unavailable"
				authorizationContext.IsAuthorized = false
				authorizationContext.AuthorizationError = &authError
				ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			_, err := db.ValidateEnrollmentToken(baseCtx, tokenValue)
			if err != nil {
				authError.ErrorDescription = err.Error()
				authorizationContext.IsAuthorized = false
				authorizationContext.AuthorizationError = &authError
				ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
				baseCtx.LogInfof("EnrollmentToken Authorization layer: invalid token: %v", err)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			authorizationContext.IsAuthorized = true
			authorizationContext.IsMicroService = true
			authorizationContext.AuthorizedBy = "EnrollmentToken"
			authorizationContext.AuthorizationError = nil
			ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
			baseCtx.LogInfof("EnrollmentToken Authorization layer finished successfully")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
