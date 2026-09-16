package restapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/security/apikey"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

type ApiKeyHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func ApiKeyAuthorizationMiddlewareAdapter(roles []string, claims []string, roleComparisonOperation ComparisonOperation, claimComparisonOperation ComparisonOperation) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			baseCtx := basecontext.NewBaseContextFromRequest(r)
			var authorizationContext *basecontext.AuthorizationContext
			authCtxFromRequest := baseCtx.GetAuthorizationContext()
			if authCtxFromRequest == nil {
				authorizationContext = basecontext.InitAuthorizationContext()
			} else {
				authorizationContext = authCtxFromRequest
			}

			// If the authorization context is already authorized we will skip this middleware
			if authorizationContext.IsAuthorized || HasAuthorizationHeader(r) {
				baseCtx.LogDebugf("No Api Key was found in the request, skipping")
				ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			baseCtx.LogInfof("ApiKey Authorization layer started")
			authError := models.OAuthErrorResponse{
				Error:            models.OAuthUnauthorizedClient,
				ErrorDescription: "The Api Key is not valid",
			}

			db := serviceprovider.Get().JsonDatabase
			_ = db.Connect(baseCtx)

			result, err := apikey.ValidateApiKey(baseCtx, db, r.Header.Get("X-Api-Key"))
			if err != nil {
				authError.ErrorDescription = err.Error()
				authorizationContext.AuthorizationError = &authError
				baseCtx.LogInfof("The Api Key is not valid: %v", err)
				ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// User-assigned keys carry the same identity and route permissions as
			// a bearer token. Unassigned legacy keys retain service authentication.
			if result.UserID != "" {
				user, userErr := db.GetUser(baseCtx, result.UserID)
				if userErr != nil || user == nil {
					authError.ErrorDescription = "The user assigned to this API key was not found"
				} else {
					apiUser := mappers.DtoUserToApiResponse(*user)
					authorizationContext.User = &apiUser
					for _, role := range user.Roles {
						if strings.EqualFold(role.Name, constants.SUPER_USER_ROLE) {
							authorizationContext.IsSuperUser = true
							apiUser.IsSuperUser = true
						}
					}
					authError.ErrorDescription = ""
					if !authorizationContext.IsSuperUser {
						for _, check := range []struct {
							kind      string
							required  []string
							operation ComparisonOperation
							has       func(string) bool
						}{
							{"roles", roles, roleComparisonOperation, authorizationContext.HasEffectiveRole},
							{"claims", claims, claimComparisonOperation, authorizationContext.HasEffectiveClaim},
						} {
							if len(check.required) == 0 {
								continue
							}
							var validations TokenRoleClaimValidationList
							for _, required := range check.required {
								validations = append(validations, &TokenRoleClaimValidation{Name: required, exists: check.has(required)})
							}
							if !validations.Evaluate(check.operation) {
								authError.ErrorDescription = fmt.Sprintf("User does not contain enough permissions for %s (%s), missing %v", check.kind, normalizeComparisonOperation(check.operation), validations.GetFailed())
								break
							}
						}
					}
				}
				if authError.ErrorDescription != "" {
					authorizationContext.IsAuthorized = false
					authorizationContext.AuthorizationError = &authError
					ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			authorizationContext.IsAuthorized = true
			authorizationContext.IsMicroService = result.UserID == "" || strings.EqualFold(result.Type, "internal")
			authorizationContext.AuthorizedBy = "ApiKeyAuthorization"
			authorizationContext.ApiKeyName = result.ApiKeyId
			authorizationContext.AuthorizationError = nil
			ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
			baseCtx.LogInfof("ApiKey Authorization layer finished")
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
