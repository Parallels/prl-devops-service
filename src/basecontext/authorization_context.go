package basecontext

import (
	"context"
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
)

type AuthorizationContext struct {
	RequestId          string
	Issuer             string
	Scope              string
	Audiences          []string
	BaseUrl            string
	IsAuthorized       bool
	IsMicroService     bool
	IsSuperUser        bool
	AuthorizedBy       string
	ApiKeyName         string
	User               *models.ApiUser
	AuthorizationError *models.OAuthErrorResponse
	// InjectedClaims/InjectedRoles are set from X-Claims/X-Roles headers on
	// trusted requests (microservices, catalog manager forwards). When present
	// they override the user's JWT-based claims/roles for handler-level checks.
	InjectedClaims []string
	InjectedRoles  []string
}

var baseAuthorizationCtx *AuthorizationContext

func InitAuthorizationContext() *AuthorizationContext {
	context := AuthorizationContext{}
	return &context
}

func GetBaseContext() *AuthorizationContext {
	if baseAuthorizationCtx == nil {
		return InitAuthorizationContext()
	}

	return baseAuthorizationCtx
}

func (c *AuthorizationContext) IsUserInRole(role string) bool {
	if c.User == nil {
		return false
	}

	for _, r := range c.User.Roles {
		if strings.EqualFold(r, role) {
			return true
		}
	}

	return false
}

func (c *AuthorizationContext) IsUserInRoles(roles []string) bool {
	if c.User == nil {
		return false
	}

	for _, role := range roles {
		if c.IsUserInRole(role) {
			return true
		}
	}

	return false
}

func (c *AuthorizationContext) UserHasClaim(claim string) bool {
	if c.User == nil {
		return false
	}

	for _, c := range c.User.Claims {
		if strings.EqualFold(c, claim) {
			return true
		}
	}

	return false
}

// GetEffectiveClaims returns InjectedClaims if present (from a trusted X-Claims
// header), otherwise the user's own claims from their JWT.
func (c *AuthorizationContext) GetEffectiveClaims() []string {
	if len(c.InjectedClaims) > 0 {
		return c.InjectedClaims
	}
	if c.User == nil {
		return []string{}
	}
	return c.User.Claims
}

// GetEffectiveRoles returns InjectedRoles if present (from a trusted X-Roles
// header), otherwise the user's own roles from their JWT.
func (c *AuthorizationContext) GetEffectiveRoles() []string {
	if len(c.InjectedRoles) > 0 {
		return c.InjectedRoles
	}
	if c.User == nil {
		return []string{}
	}
	return c.User.Roles
}

func CloneAuthorizationContext() *AuthorizationContext {
	// Creating the new context using the default values if it does not exist
	if baseAuthorizationCtx == nil {
		context := AuthorizationContext{}
		baseAuthorizationCtx = &context
	}

	newContext := AuthorizationContext{
		Issuer:       baseAuthorizationCtx.Issuer,
		Scope:        baseAuthorizationCtx.Scope,
		Audiences:    make([]string, 0),
		BaseUrl:      baseAuthorizationCtx.BaseUrl,
		IsAuthorized: false,
		RequestId:    "",
		AuthorizedBy: "",
		User:         nil,
	}

	return &newContext
}

func GetAuthorizationContext(ctx context.Context) *AuthorizationContext {
	authContext := ctx.Value(constants.AUTHORIZATION_CONTEXT_KEY)
	if authContext == nil {
		return nil
	}

	return authContext.(*AuthorizationContext)
}
