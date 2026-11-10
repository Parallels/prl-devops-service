package query

// Transcription of data/authorize_layer.go:8-100 and data/api_key.go:19-32
// (PBI #421 task 03, D-4). Zero semantic drift — bodies verbatim; the ONLY
// rename is GetOwnRecords → OwnRecords (the query-language name).
//
// SECURITY NOTE carried over verbatim: the hasInjected suppression of the
// api-key/microservice/root bypass below is an explicit past security fix
// (:45-54 comment in the legacy source). Do not "simplify" it.

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
		if (authContext.AuthorizedBy == "ApiKeyAuthorization" && authContext.IsMicroService) || authContext.AuthorizedBy == "RootAuthorization" {
			return true
		}
	}

	if !authContext.IsAuthorized {
		return false
	}
	if authContext.User == nil && !hasInjected {
		return false
	}

	// Super-user bypass: applies for direct access, and also for forwarded
	// requests when the upstream explicitly flagged the calling user as a
	// super-user via X-Super-User (authContext.IsSuperUser set by middleware).
	if authContext.IsSuperUser || (!hasInjected && authContext.HasEffectiveRole(constants.SUPER_USER_ROLE)) {
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

// OwnRecords is the query-language name for legacy data.GetOwnRecords
// (api_key.go:19-32): passthrough SAME slice (not a copy) when there is no
// auth context or no user — pinned; the ownership comparison is a
// case-SENSITIVE ==.
func OwnRecords[T interface{ GetUserID() string }](ctx basecontext.ApiContext, records ...T) []T {
	authContext := ctx.GetAuthorizationContext()
	if authContext == nil || authContext.User == nil {
		return records
	}

	result := make([]T, 0)
	for _, record := range records {
		if record.GetUserID() == authContext.User.ID {
			result = append(result, record)
		}
	}

	return result
}
