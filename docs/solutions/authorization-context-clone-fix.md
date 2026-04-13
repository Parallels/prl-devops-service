# Authorization Context Clone Fix

## Problem Description

When forwarding requests from `catalog_managers.go` endpoints to `catalog.go` endpoints, the authorization context was being lost. This caused the `XClaimsMiddlewareAdapter` to fail to trust injected claims/roles headers, resulting in authorization failures in `authorize_layer.go`.

### Symptoms

- Requests through catalog manager forwarding endpoints would lose authorization context
- `X-Claims` and `X-Roles` headers from trusted sources (catalog managers, microservices) were not processed
- `IsMicroService` flag was reset to `false` on subsequent requests
- Catalog manager catalog forwarding endpoints would fail with 403 Forbidden errors

### Flow Analysis

1. **Incoming request to catalog-managers endpoint** (e.g., `/v1/catalog-managers/{id}/catalog`)
   - Middleware chain: `AddAuthorizationContextMiddlewareAdapter` → `ApiKeyAuthorization` → `XClaimsMiddlewareAdapter`
   - Authorization context correctly populated with `IsMicroService=true` from API key

2. **Forwarding to catalog manager**
   - Request forwarded with user's `X-Claims`, `X-Roles`, `X-Super-User` headers

3. **Catalog manager returns response**
   - Response handled correctly

4. **Catalog endpoint request** (e.g., `/v1/catalog`)
   - **Problem**: `CloneAuthorizationContext()` was called, which **lost** `IsMicroService`, `InjectedClaims`, `InjectedRoles`
   - `XClaimsMiddlewareAdapter` sees `IsMicroService=false` and does NOT trust injected claims

## Root Cause

The `CloneAuthorizationContext()` function in `src/basecontext/authorization_context.go` only copied a subset of fields:

### Fields That Were Copied
- `Issuer`, `Scope`, `Audiences`, `BaseUrl`
- `IsAuthorized`, `RequestId`, `AuthorizedBy`, `User`

### Fields That Were NOT Copied (Lost)
| Field | Purpose |
|-------|---------|
| `IsMicroService` | Marks API-key authenticated requests from microservices |
| `IsSuperUser` | Flags super-user access |
| `InjectedClaims` | Claims from `X-Claims` header on trusted requests |
| `InjectedRoles` | Roles from `X-Roles` header on trusted requests |
| `ApiKeyName` | Name of the API key used |
| `AuthorizationError` | Authorization error details |

### Why This Caused Issues

The `XClaimsMiddlewareAdapter` checks `IsMicroService` to determine if injected claims/roles should be trusted:

```go
isTrustedSource := authCtx.IsMicroService ||
    strings.EqualFold(r.Header.Get("X-SOURCE"), "CATALOG_MANAGER_REQUEST")
```

When `IsMicroService` was lost in `CloneAuthorizationContext()`, the middleware would skip processing `X-Claims` and `X-Roles` headers, causing authorization failures downstream.

## Solution

Updated `CloneAuthorizationContext()` to preserve all authorization-relevant fields:

```go
func CloneAuthorizationContext() *AuthorizationContext {
    // Creating the new context using the default values if it does not exist
    if baseAuthorizationCtx == nil {
        context := AuthorizationContext{}
        baseAuthorizationCtx = &context
    }

    newContext := AuthorizationContext{
        Issuer:             baseAuthorizationCtx.Issuer,
        Scope:              baseAuthorizationCtx.Scope,
        Audiences:          make([]string, 0),
        BaseUrl:            baseAuthorizationCtx.BaseUrl,
        IsAuthorized:       false,
        RequestId:          "",
        AuthorizedBy:       "",
        User:               nil,
        IsMicroService:     baseAuthorizationCtx.IsMicroService,       // ADDED
        IsSuperUser:        baseAuthorizationCtx.IsSuperUser,          // ADDED
        ApiKeyName:         baseAuthorizationCtx.ApiKeyName,           // ADDED
        AuthorizationError: baseAuthorizationCtx.AuthorizationError,   // ADDED
    }

    // Copy injected claims and roles (they are request-independent when set)
    if len(baseAuthorizationCtx.InjectedClaims) > 0 {
        newContext.InjectedClaims = make([]string, len(baseAuthorizationCtx.InjectedClaims))
        copy(newContext.InjectedClaims, baseAuthorizationCtx.InjectedClaims)
    }
    if len(baseAuthorizationCtx.InjectedRoles) > 0 {
        newContext.InjectedRoles = make([]string, len(baseAuthorizationCtx.InjectedRoles))
        copy(newContext.InjectedRoles, baseAuthorizationCtx.InjectedRoles)
    }

    return &newContext
}
```

### Key Changes

1. **Add missing fields** to the new context copy:
   - `IsMicroService` - preserves microservice trust
   - `IsSuperUser` - preserves super-user status
   - `ApiKeyName` - preserves API key identity
   - `AuthorizationError` - preserves error state

2. **Deep copy slices** to prevent modification of base context:
   - `InjectedClaims` - creates new slice with copy
   - `InjectedRoles` - creates new slice with copy

## Files Modified

- `/Users/cjlapao/code/GitHub/devops-workspace/devops-service/src/basecontext/authorization_context.go`

## Testing

```bash
# Build verification
go build ./src/...

# Authorization middleware tests
go test ./src/restapi/... -v -run "Authorization"

# Data layer authorization tests
go test ./src/data/... -v -run "Authorization"
```

All tests pass successfully.

## Impact

- Catalog manager catalog forwarding endpoints now work correctly
- Microservice authentication is properly propagated across requests
- Injected claims/roles from trusted sources are preserved
- Authorization in `authorize_layer.go` now correctly sees the effective claims/roles

## Related Files

- `src/basecontext/authorization_context.go` - Fixed
- `src/restapi/x_claims_middleware.go` - Uses IsMicroService to trust headers
- `src/restapi/apikey_authorization_middleware.go` - Sets IsMicroService
- `src/restapi/authorization_context_middleware.go` - Creates cloned context
- `src/controllers/catalog_managers.go` - Uses catalog manager forwarding
- `src/controllers/catalog.go` - Receives forwarded requests
