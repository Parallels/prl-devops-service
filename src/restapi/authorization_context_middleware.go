package restapi

import (
	"context"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
)

func AddAuthorizationContextMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Context().Value(constants.REQUEST_ID_KEY)
			authorizationContext := basecontext.CloneAuthorizationContext()

			// Adding the request id if it exist
			if id != nil {
				authorizationContext.RequestId = id.(string)
			}

			// Adding a new Authorization Request to the Request
			ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
			baseCtx := basecontext.NewBaseContextFromContext(ctx)
			baseCtx.LogInfo("Authorization layer started")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
