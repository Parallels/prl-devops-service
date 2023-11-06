package basecontext

import (
	"context"
	"strings"

	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/models"
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
	User               *models.ApiUser
	AuthorizationError *models.OAuthErrorResponse
}

var (
	baseAuthorizationCtx *AuthorizationContext
)

func InitAuthorizationContext() *AuthorizationContext {
	if baseAuthorizationCtx == nil {
		context := AuthorizationContext{}

		baseAuthorizationCtx = &context
	}

	return baseAuthorizationCtx
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
