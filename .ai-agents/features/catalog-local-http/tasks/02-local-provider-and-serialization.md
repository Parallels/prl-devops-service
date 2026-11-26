# 02 — Provider and serialization fixes (bugs 1, 2, 4)

**Depends on:** nothing. **Blocks:** 05 (the HTTP provider cannot recover its host without
bug 1 fixed).

## Bug 1 — `CatalogManifestProvider.String()` drops credentials

`src/catalog/models/catalog_manifest_provider.go:19-33` guards every first-class field with
`if m.X == ""` before appending it, so the string handed to `Check()` at `catalog/pull.go:309`
carries only `provider=` plus `Meta`.

- Emit a field only when non-empty; keep `Meta` appended last; keep the `;` joining and the
  trailing-`;` trim.
- Add a `SafeString()` (or an explicit `redact` argument) that masks `password`/`api_key`
  values, and use it at the two log sites: `src/catalog/push.go:421` and
  `src/catalog/import_vm.go:136` (`src/catalog/AGENTS.md`: connection strings are secrets).
- Do **not** "fix" `Parse`'s `strings.Contains(part, "k=")` matching here — it is a separate
  wart; note it and leave it.

## Bug 2 — local provider path containment

`src/catalog/providers/local/main.go` repeats a string-prefix test in ten places
(`:110, :147, :169, :210, :239, :251, :259, :282, :293, :308`).

- One private `resolve(path, name string) (string, error)`: join, `filepath.Clean`,
  `filepath.EvalSymlinks` where the target exists, then contain with `filepath.Rel` and a
  `..` check — never `strings.HasPrefix`. Every method calls it; the silent
  "re-root under `Config.Path`" behaviour disappears, so a caller that passes an out-of-root
  path gets an error instead of a surprise location.
- `catalog_path` becomes mandatory: `Check()` returns `false` (do not error) when absent, and
  the reason is logged once. Delete the `$(cwd)/catalog` fallback (bug 2's root cause).
- Directories `0750`, files `0640` — not `os.ModePerm`.
- Keep `CanStream()` false; make `SetJobId`/`SetCurrentAction` mutate under a mutex and
  document that one instance must not be shared across concurrent jobs.

## Bug 4 — a `Check()` error kills the provider loop

`src/catalog/pull.go:310-314` and `src/catalog/main.go:95-98` break out on error. Log at debug
and continue to the next provider; keep "no provider matched" as the user-visible failure.

## Tests

- `catalog/models/catalog_manifest_provider_test.go`: `Parse` → `String` → `Check` round-trip
  per provider family, asserting a credential-bearing string still selects the intended
  provider; `SafeString` never contains the secret substring.
- `catalog/providers/local/containment_test.go` (table): `../../etc/passwd`, sibling prefix
  (`root+"/-evil"`), absolute `path`, symlink out of root, empty name, name with a trailing
  separator — each refused, and a full tree scan proves nothing outside the root was created.
  Plus: missing `catalog_path` ⇒ `Check()` false and no directory created anywhere.
- `catalog/pull_provider_loop_test.go`: a fake provider whose `Check` errors must not prevent a
  later provider from matching.

## Definition of done

`go test ./catalog/...` green; no `strings.HasPrefix` containment check remains in the local
provider; no log line in `catalog/` can print a secret.
