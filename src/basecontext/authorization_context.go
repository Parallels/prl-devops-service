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

// HasEffectiveRole checks the effective roles (InjectedRoles when present, else
// User.Roles) for the given role, case-insensitively.
func (c *AuthorizationContext) HasEffectiveRole(role string) bool {
	for _, r := range c.GetEffectiveRoles() {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}

// HasEffectiveClaim checks the effective claims (InjectedClaims when present,
// else User.Claims) for the given claim, case-insensitively.
func (c *AuthorizationContext) HasEffectiveClaim(claim string) bool {
	for _, c := range c.GetEffectiveClaims() {
		if strings.EqualFold(c, claim) {
			return true
		}
	}
	return false
}

// GetEffectiveClaims returns InjectedClaims if present (from a trusted X-Claims
// header). Otherwise it returns the user's full merged claim set: if
// User.EffectiveClaims is populated (includes role-inherited claims) its Name
// values are used; if empty, falls back to the directly-assigned User.Claims.
func (c *AuthorizationContext) GetEffectiveClaims() []string {
	if len(c.InjectedClaims) > 0 {
		return c.InjectedClaims
	}
	if c.User == nil {
		return []string{}
	}
	if len(c.User.EffectiveClaims) > 0 {
		claims := make([]string, 0, len(c.User.EffectiveClaims))
		for _, ec := range c.User.EffectiveClaims {
			if ec.Name != "" {
				claims = append(claims, ec.Name)
			}
		}
		return claims
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
