package data

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
)

type AuthorizedRecord interface {
	GetRequiredClaims() []string
	GetRequiredRoles() []string
}

func GetAuthorizedRecords[T AuthorizedRecord](ctx basecontext.ApiContext, t ...T) []T {
	result := make([]T, 0)
	for _, record := range t {
		if IsAuthorized(ctx, record) {
			result = append(result, record)
		}
	}

	return result
}

func IsRootUser(ctx basecontext.ApiContext) bool {
	authContext := ctx.GetAuthorizationContext()
	if authContext == nil {
		return false
	}
	if authContext.AuthorizedBy == "RootAuthorization" {
		return true
	}
	if authContext.User != nil && authContext.User.Username == "root" {
		return true
	}

	return false
}

func IsAuthorized[T AuthorizedRecord](ctx basecontext.ApiContext, t T) bool {
	authContext := ctx.GetAuthorizationContext()
	if authContext == nil {
		return false
	}

	// When injected claims/roles are present (catalog manager or microservice
	// forwarded request) we must apply record-level filtering using those values
	// rather than bypassing. Otherwise API-key / root requests skip filtering.
	hasInjected := len(authContext.InjectedClaims) > 0 || len(authContext.InjectedRoles) > 0

	if !hasInjected {
		if authContext.AuthorizedBy == "ApiKeyAuthorization" || authContext.AuthorizedBy == "RootAuthorization" {
			return true
		}
	}

	if !authContext.IsAuthorized {
		return false
	}
	if authContext.User == nil && !hasInjected {
		return false
	}

	if authContext.HasEffectiveRole(constants.SUPER_USER_ROLE) {
		return true
	}

	isAuthorized := false
	hasClaims := false

	requiredRoles := t.GetRequiredRoles()
	requiredClaims := t.GetRequiredClaims()
	if len(requiredRoles) == 0 {
		isAuthorized = true
	} else {
		for _, role := range requiredRoles {
			if authContext.HasEffectiveRole(role) {
				isAuthorized = true
			}
		}
	}

	if len(requiredClaims) == 0 {
		hasClaims = true
	} else {
		for _, claim := range requiredClaims {
			if authContext.HasEffectiveClaim(claim) {
				hasClaims = true
			}
		}
	}

	if isAuthorized && hasClaims {
		return true
	}

	return false
}
