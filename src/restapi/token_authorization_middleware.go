package restapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	data_modules "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/dgrijalva/jwt-go"
)

// TokenAuthorizationMiddlewareAdapter validates a Authorization Bearer during a rest api call
// It can take an array of roles and claims to further validate the token in a more granular
// view, it also can take an OR option in both if the role or claim are coma separated.
// For example a claim like "_read,_write" will be valid if the user either has a _read claim
// or a _write claim making them both valid
func TokenAuthorizationMiddlewareAdapter(roles []string, claims []string) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg := config.NewConfig()
			baseCtx := basecontext.NewBaseContextFromRequest(r)
			var authorizationContext *basecontext.AuthorizationContext
			authCtxFromRequest := baseCtx.GetAuthorizationContext()
			if authCtxFromRequest == nil {
				authorizationContext = basecontext.InitAuthorizationContext()
			} else {
				authorizationContext = authCtxFromRequest
			}

			if authorizationContext.IsAuthorized || HasApiKeyAuthorizationHeader(r) {
				baseCtx.LogDebug("No bearer token was found in the request, skipping")
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

			//Starting authorization layer of the token
			authorized := true
			baseCtx.LogInfo("Token Authorization layer started")

			// Getting the token for validation
			jwt_token, valid := http_helper.GetAuthorizationToken(r.Header)
			if !valid {
				authorized = false
				validateError := errors.New("bearer token not found in request")
				baseCtx.LogError("Error validating token, %v", validateError.Error())
			}

			// Validating userToken against the keys
			if authorized {
				token, err := jwt.Parse(jwt_token, func(token *jwt.Token) (interface{}, error) {
					// Validate the algorithm
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, jwt.ErrSignatureInvalid
					}

					// Return the secret key used to sign the token
					// We either signing the token with the HMAC secret or the secret from the config
					var key []byte
					if cfg.GetHmacSecret() == "" {
						key = []byte(cfg.GetHmacSecret())
					} else {
						key = []byte(serviceprovider.Get().HardwareSecret)
					}
					return key, nil
				})

				if err != nil {
					authorized = false
					response := models.OAuthErrorResponse{
						Error:            models.OAuthUnauthorizedClient,
						ErrorDescription: err.Error(),
					}
					authorizationContext.AuthorizationError = &response
					baseCtx.LogError("Request failed to authorize, %v", response.ErrorDescription)
				}

				if authorized {
					// Check if the token is valid
					if jwtClaims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
						db := serviceprovider.Get().JsonDatabase
						var dbUser *data_modules.User
						var err error
						if err = db.Connect(baseCtx); err != nil {
							authorized = false
						} else {
							dbUser, err = db.GetUser(baseCtx, jwtClaims["email"].(string))
							if err != nil || dbUser == nil {
								authorized = false
							}
							// Checking for the Super Duper User
							authorizationContext.IsSuperUser = false
							for _, userRole := range dbUser.Roles {
								if strings.EqualFold(constants.SUPER_USER_ROLE, userRole.Name) {
									authorizationContext.IsSuperUser = true
									break
								}
							}
							if !authorizationContext.IsSuperUser {
								// Checking if the user has the correct role required by the controller
								if len(roles) > 0 {
									contains := false
									for _, role := range roles {
										for _, userRole := range dbUser.Roles {
											if strings.EqualFold(role, userRole.Name) {
												contains = true
												break
											}
										}
										if contains {
											break
										}
									}
									if !contains {
										authorized = false
										response := models.OAuthErrorResponse{
											Error:            models.OAuthUnauthorizedClient,
											ErrorDescription: "User does not contain enough permissions, not in role",
										}
										authorizationContext.IsAuthorized = false
										authorizationContext.AuthorizationError = &response
										baseCtx.LogError("Request failed to authorize, %v", response.ErrorDescription)
									}
								}

								if len(claims) > 0 {
									contains := false
									for _, claim := range claims {
										for _, userClaim := range dbUser.Claims {
											if strings.EqualFold(claim, userClaim.Name) {
												contains = true
												break
											}
										}
										if contains {
											break
										}
									}
									if !contains {
										authorized = false
										response := models.OAuthErrorResponse{
											Error:            models.OAuthUnauthorizedClient,
											ErrorDescription: "User does not contain enough permissions, does not have claim",
										}
										authorizationContext.IsAuthorized = false
										authorizationContext.AuthorizationError = &response
										baseCtx.LogError("Request failed to authorize, %v", response.ErrorDescription)
									}
								}
							}
						}
						if !authorized {
							response := models.OAuthErrorResponse{
								Error:            models.OAuthUnauthorizedClient,
								ErrorDescription: "User not found",
							}
							authorizationContext.IsAuthorized = false
							if authorizationContext.AuthorizationError == nil {
								authorizationContext.AuthorizationError = &response
							}
							baseCtx.LogError("Request failed to authorize, %v", response.ErrorDescription)
						} else {
							user := mappers.DtoUserToApiResponse(*dbUser)
							authorizationContext.User = &user
							authorizationContext.IsAuthorized = true
							authorizationContext.AuthorizedBy = "TokenAuthorization"
						}
					} else {
						response := models.OAuthErrorResponse{
							Error:            models.OAuthUnauthorizedClient,
							ErrorDescription: "Token is not valid",
						}
						authorizationContext.IsAuthorized = false
						authorizationContext.AuthorizationError = &response
						baseCtx.LogError("Request failed to authorize, %v", response.ErrorDescription)
					}
				}
			}

			ctx := context.WithValue(r.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authorizationContext)
			if authorizationContext.User != nil {
				baseCtx.LogInfo("Token Authorization layer finished, user %v authorized", authorizationContext.User.Email)
			} else {
				baseCtx.LogInfo("Token Authorization layer finished, no user authorized")
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
