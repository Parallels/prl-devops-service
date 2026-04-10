# IsSuperUser with Injected Claims/Fix

## Problem Description

When claims and roles are forwarded via `X-Claims` and `X-Roles` headers from trusted sources (catalog managers, microservices), the `IsSuperUser` check in `authorize_layer.go` was incorrectly using the authenticated user's super-user status instead of relying solely on the `X-Super-User` header.

### Root Cause

The `XClaimsMiddlewareAdapter` was only setting `IsSuperUser = true` when the `X-Super-User` header was explicitly "true", but it was NOT clearing `IsSuperUser` when:
1. Injected claims/roles were present (from `X-Claims`/`X-Roles` headers)
2. The `X-Super-User` header was NOT present

This meant that if a super-user had previously made a request that set `IsSuperUser = true` in the base context, subsequent requests with forwarded non-super-user claims would inherit `IsSuperUser = true`, bypassing authorization checks.

### Security Impact

A non-super-user making requests through catalog managers could potentially:
- Bypass role/claim-based authorization checks
- Access all catalog manifests regardless of required roles/claims
- View resources they should not have access to

### Example Scenario

1. **User A (Super User)** makes request to `/catalog-managers/{id}/catalog`
   - `authCtx.IsSuperUser = true` (from JWT)
   - `catalog_managers.go` forwards with User A's claims and `X-Super-User: true`

2. **User B (Non-Super User)** makes request to `/catalog-managers/{id}/catalog`
   - `authCtx.IsSuperUser = false` (from JWT)
   - `catalog_managers.go` forwards with User B's claims, NO `X-Super-User` header
   - **BUG**: `authCtx.IsSuperUser` remains `true` from User A's request in base context
   - User B can now access everything because `IsSuperUser` is true

## Solution

Modified `XClaimsMiddlewareAdapter` in `src/restapi/x_claims_middleware.go` to explicitly clear `IsSuperUser` when:
1. Injected claims/roles are present AND
2. `X-Super-User` header is not set to "true"

### Code Change

```go
superUserHeader := r.Header.Get(constants.X_SUPER_USER_HEADER)
if strings.EqualFold(superUserHeader, "true") {
    // Only set IsSuperUser when the X-Super-User header is explicitly "true"
    // This is critical for security: when claims/roles are forwarded from a downstream
    // service, the super-user status should ONLY apply if explicitly declared via header.
    authCtx.IsSuperUser = true
} else if len(authCtx.InjectedClaims) > 0 || len(authCtx.InjectedRoles) > 0 {
    // When injected claims/roles are present but X-Super-User is not "true",
    // reset IsSuperUser to false to prevent using the authenticated user's super-user status.
    // This ensures that forwarded permissions are strictly limited to the forwarded claims/roles.
    authCtx.IsSuperUser = false
}
```

### How It Works

1. **X-Super-User = "true"**: `IsSuperUser = true` - Allow full access
2. **X-Super-User not present, but claims/roles forwarded**: `IsSuperUser = false` - Use only forwarded claims/roles
3. **No injected claims/roles**: `IsSuperUser` unchanged - Use authenticated user's status

### Authorization Flow with Fix

```
Request arrives at /catalog endpoint
    |
    v
ApiKeyAuthorization: IsMicroService=true, IsSuperUser=??? (cloned from base)
    |
    v
XClaimsMiddleware:
    - Reads X-Claims → InjectedClaims = [user-claims]
    - Reads X-Roles → InjectedRoles = [user-roles]
    - Reads X-Super-User → NOT PRESENT
    - BECAUSE injected claims/roles exist AND X-Super-User != "true":
      IsSuperUser = false
    |
    v
authorize_layer.go:
    hasInjected = true
    if authContext.IsSuperUser || (!hasInjected && authContext.HasEffectiveRole(SUPER_USER)) {
        return true  // NOT TRIGGERED: IsSuperUser=false
    }
    // Authorization proceeds with injected claims/roles only
```

## Files Modified

- `/Users/cjlapao/code/GitHub/devops-workspace/devops-service/src/restapi/x_claims_middleware.go`

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

## Related Files

- `src/restapi/x_claims_middleware.go` - Fixed
- `src/restapi/authorization_context_middleware.go` - Creates cloned context
- `src/basecontext/authorization_context.go` - CloneAuthorizationContext (fixed in separate issue)
- `src/data/authorize_layer.go` - Uses IsSuperUser and injected claims/roles
- `src/controllers/catalog_managers.go` - Sets X-Super-User header when forwarding

## Security Notes

This fix implements the principle of **explicit permission declaration**:
- Super-user status should ONLY be granted when explicitly declared via the `X-Super-User` header
- Forwarded permissions should be strictly limited to the claims/roles that were forwarded
- Never inherit super-user status implicitly from the authentication context
