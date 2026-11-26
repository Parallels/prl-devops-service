# security

Security primitives. Treat every change here as security-sensitive and add tests.

- `jwt/` — token creation/validation and options (`jwt.New(ctx)` at startup).
- `password/` — password hashing and options.
- `apikey/` — API key generation/handling.
- `brute_force_guard/` — login throttling/lockout.
- `main.go` — RSA key generation and string encrypt/decrypt helpers, random string generation (use these `crypto/rand` based helpers, never `math/rand`, for secrets).

Rules: never log secrets, tokens, keys or password hashes; never store plaintext secrets; do not weaken defaults (lengths, algorithms, lockout thresholds) or add bypasses. Authorization (claims/roles) is enforced in `../restapi` middleware, documented in `docs/docs/devops/security/` (`rbac.md`, `hardening-security.md`). After a security-related change run the full `make test`.
