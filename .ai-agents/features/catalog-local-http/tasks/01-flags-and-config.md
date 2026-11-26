# 01 — Flags, config accessors, settings catalogue

**Depends on:** nothing. **Blocks:** 03, 04, 05.

## Goal

One switch, one root, one upload cap — readable from env, config file and the settings service
with identical semantics, and off by default so an upgrade changes nothing.

## Steps

1. `src/constants/catalog_local_http.go`: `CATALOG_LOCAL_HTTP_ENABLED_ENV_VAR`,
   `CATALOG_LOCAL_PATH_ENV_VAR`, `CATALOG_LOCAL_MAX_UPLOAD_BYTES_ENV_VAR`, plus defaults
   (`false`, `""`, `51200` MiB). Mirror the shape and comment style of
   `src/constants/placement.go`.
2. `src/config/catalog_local_http.go`: `CatalogLocalHttpEnabled() bool`,
   `CatalogLocalPath() string` (trimmed, trailing separator stripped, refuses a relative path —
   an operator must not get a cwd-relative root), `CatalogLocalMaxUploadBytes() int64`.
   Mirror `src/config/placement.go`, including how it clamps and how it reports a bad value.
3. `src/settings/catalog.go`: three definitions, orchestrator-applied, all
   `RequiresRestart: true`, defaults taken from the constants so the two cannot drift (same
   trick as the placement keys).
4. Wire the gate in `src/controllers/main.go` next to the reverse-proxy precedent
   (`:50-52`): register only when `config.Get().IsCatalog() && cfg.CatalogLocalHttpEnabled()`
   **and** the root resolves. Enabled with an unusable root logs one error at startup naming
   the key, and registers nothing — never falls back to a guessed directory.

## Tests

- `src/config/catalog_local_http_test.go`: table over unset / `"true"` / `"not-a-bool"` /
  whitespace, absolute vs relative root, MiB parsing, cap bounds; and one case proving the
  settings override wins over env (reuse the `SetOverrideResolver` pattern from
  `src/config/override_test.go`).
- Settings: the three keys exist, types and defaults match the constants, `RequiresRestart` is
  true on all three.

## Definition of done

`go test ./config/ ./settings/` green; with the flag unset, no new route exists (asserted in
card 04); nothing in the codebase reads the env var directly.
