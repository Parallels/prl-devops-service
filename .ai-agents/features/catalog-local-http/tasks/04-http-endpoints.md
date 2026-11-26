# 04 — HTTP endpoints and registration

**Depends on:** 01, 03. **Blocks:** 05 (needs a server to talk to), 06.

## Steps

1. `src/controllers/catalog_local_http.go` — thin handlers over `localcatalog.Service`, one
   file for read + write, built the way `registerCatalogManifestHandlers` builds
   (`src/controllers/catalog.go:30-103`): `restapi.NewController()` +
   `WithMethod` + `WithVersion` + `WithPath` + `WithRequiredClaim` + `WithHandler` + `Register`.
   - `GET|HEAD /catalog/{catalogId}/{version}/{architecture}/pack` and `/metadata` →
     `WithRequiredClaim(constants.PULL_CATALOG_MANIFEST_CLAIM)` (decision 1).
   - `PUT` on the same two paths → `WithRequiredClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM)`.
   - Handler body: build the service (or receive it), `Resolve`, then `http.ServeContent` with
     an explicit name; on write, `MaxBytesReader` then `Put`.
2. **Override the JSON content type.** `JsonContentMiddlewareAdapter`
   (`src/restapi/http_listener.go:239-242`) stamps `application/json` on every response; set
   `application/octet-stream` explicitly, as `src/controllers/config.go:512` does for
   `text/plain`. Also set `Content-Disposition: attachment; filename="<basename>"`,
   `Accept-Ranges: bytes`, `X-Package-Checksum`, `X-Package-Size`.
3. Principal check inside the handler (plan R6): deny unless
   `authorizationContext.User != nil`, because
   `src/restapi/apikey_authorization_middleware.go:65` runs claim checks only for keys with a
   `UserID`. This is the difference between "a claim gates 50 GB" and "any legacy user-less key
   gates 50 GB".
4. Map the service's typed errors to statuses exactly as plan §1 tabulates
   (`400` tainted, `403` revoked, `404` unknown-or-outside-root, `409` no row for a write,
   `413` too large, `422` checksum mismatch). Use the repo's error helpers
   (`ReturnApiError`, `models.NewFromErrorWithCode`) and `errors.Diagnostics` style, not bare
   `fmt.Errorf`.
5. `registerCatalogLocalHttpHandlers(ctx, version)` called from `src/controllers/main.go` under
   the card-01 gate. **Do not count downloads here** — `/download` already increments
   (`src/controllers/catalog.go:596`); one logical download must remain one counter bump.
6. Swagger annotations on each handler (`@Produce octet-stream`, the three path params,
   `@Success 200` / `206` / the failure shapes) — `make generate-swagger` stays the user's run.

## Tests — `src/controllers/catalog_local_http_test.go`

- **Absent when disabled** (the user's hard requirement): with the flag off, the handler must
  never be registered. Assert it at the registration function, not by hoping a `404` means
  unregistered — call `registerCatalogLocalHttpHandlers` behind the same predicate and assert
  the listener's route table has no `/catalog/.../pack`, and assert the reverse for on.
- Read: `401` unauthenticated; `403` missing claim; **`403` for an authenticated but unassigned
  key** (R6); `200` with `Content-Type` *not* `application/json`; `206` for a `Range`; `400`
  tainted; `403` revoked; `404` for an unknown triple and for a row outside the root.
- Write: `413` over the cap; `422` on checksum mismatch (and the file is gone afterwards);
  `409` with no row; success yields a file with `0640` and no leftover `.part`.
- Reuse the existing controller test harness style in
  `src/controllers/machines_local_catalog_test.go` (`setupTestDBForController`) and, once card 06
  lands, `catalogtest` for the seeded root.

## Definition of done

No path logic in the controller; every status above has a test; routes demonstrably absent with
the flag off; `go vet` clean and the file under ~500 lines (split read/write handlers before
exceeding it).
