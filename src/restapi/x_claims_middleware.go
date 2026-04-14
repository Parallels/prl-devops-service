package restapi

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
)

// XClaimsMiddlewareAdapter extracts X-Claims and X-Roles headers from trusted
// sources (microservice API-key auth or catalog-manager forwarded requests) and
// stores them in the authorization context as injected overrides. When present
// they take precedence over the user's JWT-based claims/roles for all
// handler-level permission checks.
//
// Security: headers are only honoured when the request is already identified as
// coming from a trusted source (IsMicroService=true or
// X-SOURCE=CATALOG_MANAGER_REQUEST), preventing end-users from escalating their
// own permissions by adding these headers directly.
func XClaimsMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			baseCtx := basecontext.NewBaseContextFromRequest(r)
			authCtx := baseCtx.GetAuthorizationContext()

			if authCtx == nil {
				next.ServeHTTP(w, r)
				return
			}

			isTrustedSource := authCtx.IsMicroService ||
				strings.EqualFold(r.Header.Get("X-SOURCE"), "CATALOG_MANAGER_REQUEST")

			if !isTrustedSource {
				next.ServeHTTP(w, r)
				return
			}

			if claimsHeader := r.Header.Get(constants.X_CLAIMS_HEADER); claimsHeader != "" {
				decoded, err := base64.StdEncoding.DecodeString(claimsHeader)
				if err == nil && len(decoded) > 0 {
					parts := strings.Split(string(decoded), ",")
					claims := make([]string, 0, len(parts))
					for _, p := range parts {
						if trimmed := strings.TrimSpace(p); trimmed != "" {
							claims = append(claims, trimmed)
						}
					}
					authCtx.InjectedClaims = claims
				}
			}

			if rolesHeader := r.Header.Get(constants.X_ROLES_HEADER); rolesHeader != "" {
				decoded, err := base64.StdEncoding.DecodeString(rolesHeader)
				if err == nil && len(decoded) > 0 {
					parts := strings.Split(string(decoded), ",")
					roles := make([]string, 0, len(parts))
					for _, p := range parts {
						if trimmed := strings.TrimSpace(p); trimmed != "" {
							roles = append(roles, trimmed)
						}
					}
					authCtx.InjectedRoles = roles
				}
			}

			superUserHeader := r.Header.Get(constants.X_SUPER_USER_HEADER)
			if strings.EqualFold(superUserHeader, "true") {
				// Only set IsSuperUser when the X-Super-User header is explicitly "true"
				// This is critical for security: when claims/roles are forwarded from a downstream
				// service, the super-user status should ONLY apply if explicitly declared via header.
				authCtx.IsSuperUser = true
			} else if len(authCtx.InjectedClaims) > 0 || len(authCtx.InjectedRoles) > 0 {
				// When injected claims/roles are present but X-Super-User is not "true",
				// reset IsSuperUser to false to prevent using the authenticated user's super-user status.
				// This ensures that forwarded permissions are strictly limited to the forwarded claims/roles.
				authCtx.IsSuperUser = false
			}

			ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
