# restapi

HTTP infrastructure: listener (TLS, CORS), the `Controller` builder, middleware chain, swagger serving. No business logic.

- Register an endpoint with the builder (see any `controllers/*.go`):
  `restapi.NewController().WithMethod(restapi.GET).WithVersion(v).WithPath("/x").WithRequiredClaim(...).WithOrClaims().WithHandler(h).Register()`.
- Auth: `token_authorization_middleware.go` (JWT), `apikey_authorization_middleware.go`, `enrollment_token_middleware.go`, `x_claims_middleware.go`, role/claim checks in `token_role_claim_validation.go`. `end_authorization_middleware.go` runs last. `comparison_operation.go` defines AND/OR semantics for roles and claims.
- Auth context for the request is built in `authorization_context_middleware.go` (see `basecontext`).
- Changing middleware order or claim logic is security-sensitive: update tests (`*_middleware_test.go`) and `docs/docs/devops/security/rbac.md` and `restapi/authentication.md`.
- Listener options live in `http_listener_options.go`; TLS/CORS behavior is tested in `http_listener_*_test.go`.
