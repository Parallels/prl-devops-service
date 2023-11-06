package data

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
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

func IsAuthorized[T AuthorizedRecord](ctx basecontext.ApiContext, t T) bool {
	authContext := ctx.GetAuthorizationContext()
	if authContext == nil {
		return false
	}
	if authContext.AuthorizedBy == "ApiKeyAuthorization" || authContext.AuthorizedBy == "RootAuthorization" {
		return true
	}
	if authContext.User == nil || !authContext.IsAuthorized {
		return false
	}
	if authContext.IsUserInRole(constants.SUPER_USER_ROLE) {
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
			if authContext.IsUserInRole(role) {
				isAuthorized = true
			}
		}
	}

	if len(requiredClaims) == 0 {
		hasClaims = true
	} else {
		for _, claim := range requiredClaims {
			if authContext.UserHasClaim(claim) {
				hasClaims = true
			}
		}
	}

	if isAuthorized && hasClaims {
		return true
	}

	return false
}
