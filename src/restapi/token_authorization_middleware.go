package restapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	data_modules "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/security/jwt"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
)

// TokenAuthorizationMiddlewareAdapter validates a Authorization Bearer during a rest api call
// It can take an array of roles and claims to further validate the token in a more granular
// view, it also can take an OR option in both if the role or claim are coma separated.
// For example a claim like "_read,_write" will be valid if the user either has a _read claim
// or a _write claim making them both valid
func TokenAuthorizationMiddlewareAdapter(roles []string, claims []string) Adapter {
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

			if authorizationContext.IsAuthorized || HasApiKeyAuthorizationHeader(r) {
				baseCtx.LogDebugf("No bearer token was found in the request, skipping")
				next.ServeHTTP(w, r)
				return
			}

			// this is not for us, move on
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
				authorizationContext.IsAuthorized = false
				ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			db := serviceprovider.Get().JsonDatabase

			// we do not have enough information to validate the token
			if db == nil {
				authorizationContext.IsAuthorized = false
				ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Setting the tenant in the context
			authorizationContext.Issuer = "Global"

			// Starting authorization layer of the token
			authorized := true
			baseCtx.LogInfof("Token Authorization layer started")

			// Getting the token for validation
			jwt_token, valid := http_helper.GetAuthorizationToken(r.Header)
			if !valid {
				authorized = false
				validateError := errors.New("bearer token not found in request")
				baseCtx.LogErrorf("Error validating token, %v", validateError.Error())
			}

			// Validating userToken against the keys
			var token *jwt.JwtSystemToken
			// Validating if the token can be parsed
			if authorized {
				jwtSvc := jwt.Get()
				var err error
				token, err = jwtSvc.Parse(jwt_token)
				if err != nil || token == nil {
					authorized = false
					response := models.OAuthErrorResponse{
						Error:            models.OAuthUnauthorizedClient,
						ErrorDescription: err.Error(),
					}
					authorizationContext.IsAuthorized = false
					authorizationContext.AuthorizationError = &response
					baseCtx.LogErrorf("Request failed to authorize, %v", response.ErrorDescription)
				}
			}

			// Validating if the token is valid
			if authorized {
				valid, err := token.Valid()
				if err != nil || !valid {
					authorized = false
					if err == nil {
						err = errors.New("invalid token")
					}
					response := models.OAuthErrorResponse{
						Error:            models.OAuthUnauthorizedClient,
						ErrorDescription: err.Error(),
					}

					authorizationContext.IsAuthorized = false
					authorizationContext.AuthorizationError = &response
					baseCtx.LogErrorf("Request failed to authorize, %v", response.ErrorDescription)
				}
			}

			// Validating if the token has the correct email
			var email interface{}
			if authorized {
				var err error
				email, err = token.GetClaim("email")
				if err != nil {
					authorized = false
					response := models.OAuthErrorResponse{
						Error:            models.OAuthUnauthorizedClient,
						ErrorDescription: err.Error(),
					}
					authorizationContext.IsAuthorized = false
					authorizationContext.AuthorizationError = &response
					baseCtx.LogErrorf("Request failed to authorize, %v", response.ErrorDescription)
				}
			}

			// Validating if the token has the correct user
			var dbUser *data_modules.User
			if authorized {
				db := serviceprovider.Get().JsonDatabase
				var err error

				// validating if the database is connected
				if err = db.Connect(baseCtx); err != nil {
					authorized = false
					response := models.OAuthErrorResponse{
						Error:            models.OAuthUnauthorizedClient,
						ErrorDescription: fmt.Sprintf("Error connecting to database, %v", err.Error()),
					}
					authorizationContext.IsAuthorized = false
					authorizationContext.AuthorizationError = &response
					baseCtx.LogErrorf("Request failed to authorize, %v", response.ErrorDescription)
				}

				// validating if the user exists
				if authorized {
					dbUser, err = db.GetUser(baseCtx, email.(string))
					if err != nil || dbUser == nil {
						authorized = false
						response := models.OAuthErrorResponse{
							Error:            models.OAuthUnauthorizedClient,
							ErrorDescription: fmt.Sprintf("Error connecting to database, %v", err.Error()),
						}
						authorizationContext.IsAuthorized = false
						authorizationContext.AuthorizationError = &response
						baseCtx.LogErrorf("Request failed to authorize, %v", response.ErrorDescription)
					}
				}

				if authorized {
					// Checking for the Super Duper User
					authorizationContext.IsSuperUser = false
					for _, userRole := range dbUser.Roles {
						if strings.EqualFold(constants.SUPER_USER_ROLE, userRole.Name) {
							authorizationContext.IsSuperUser = true
							break
						}
					}

					// Validating if the user has the correct roles and claims
					if !authorizationContext.IsSuperUser {
						// Validating if the user has the correct roles
						if len(roles) > 0 {
							rolesCheck := TokenRoleClaimValidationList{}
							for _, role := range roles {
								roleCheck := &TokenRoleClaimValidation{Name: role}
								for _, userRole := range dbUser.Roles {
									if strings.EqualFold(role, userRole.Name) {
										roleCheck.SetExists(true)
										break
									}
								}
								rolesCheck = append(rolesCheck, roleCheck)
							}

							if len(roles) != len(rolesCheck) || !rolesCheck.Exists() {
								failed := rolesCheck.GetFailed()
								authorized = false
								response := models.OAuthErrorResponse{
									Error:            models.OAuthUnauthorizedClient,
									ErrorDescription: fmt.Sprintf("User does not contain enough permissions, does not have roles, %v", failed),
								}

								authorizationContext.IsAuthorized = false
								authorizationContext.AuthorizationError = &response
								baseCtx.LogErrorf("Request failed to authorize, %v", response.ErrorDescription)
							}
						}

						if authorized {
							// Validating if the user has the correct claims
							if len(claims) > 0 {
								claimsCheck := TokenRoleClaimValidationList{}
								for _, claim := range claims {
									claimCheck := &TokenRoleClaimValidation{Name: claim}
									for _, userClaim := range dbUser.Claims {
										if strings.EqualFold(claim, userClaim.Name) {
											claimCheck.SetExists(true)
											break
										}
									}
									claimsCheck = append(claimsCheck, claimCheck)
								}

								if len(claims) != len(claimsCheck) || !claimsCheck.Exists() {
									failed := claimsCheck.GetFailed()
									authorized = false
									response := models.OAuthErrorResponse{
										Error:            models.OAuthUnauthorizedClient,
										ErrorDescription: fmt.Sprintf("User does not contain enough permissions, does not have claims, %v", failed),
									}

									authorizationContext.IsAuthorized = false
									authorizationContext.AuthorizationError = &response
									baseCtx.LogErrorf("Request failed to authorize, %v", response.ErrorDescription)
								}
							}
						}
					}
				}
			}

			if authorized {
				user := mappers.DtoUserToApiResponse(*dbUser)
				authorizationContext.User = &user
				authorizationContext.IsAuthorized = true
				authorizationContext.AuthorizedBy = "TokenAuthorization"
			}

			ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
			if authorizationContext.User != nil {
				baseCtx.LogInfof("Token Authorization layer finished, user %v authorized", authorizationContext.User.Email)
			} else {
				baseCtx.LogInfof("Token Authorization layer finished, no user authorized")
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
