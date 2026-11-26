# Plan — local catalog served over HTTP

Design for storing catalog machines in a local directory and serving them to hosts over HTTP
through the same code path a host uses against any other catalog — and for testing all of it
without cloud credentials. Read `brief.md` first: the four bugs listed there are part of this
change, and the three user decisions are binding.

---

## 1. Wire contract

Both groups are registered **only** when the feature is on (see §3) and only inside the
`catalog` module (`config.Get().IsCatalog()`, `src/controllers/main.go:26-28`). Identity comes
from the URL triple plus the database row, never from a client-supplied path or filename.

### Read

```
GET|HEAD  /api/v1/catalog/{catalogId}/{version}/{architecture}/pack
GET|HEAD  /api/v1/catalog/{catalogId}/{version}/{architecture}/metadata
```

| Response | When |
|---|---|
| `200` + bytes | row resolves, file exists, containment proven, not tainted/revoked |
| `206` + partial | request carries `Range` (`http.ServeContent` gives this, `Accept-Ranges: bytes`) |
| `400` | manifest tainted (parity with `/download`, `catalog.go:580-586`) |
| `401` / `403` | unauthenticated / revoked (`catalog.go:588-594`) / missing `PULL_CATALOG_MANIFEST_CLAIM` |
| `404` | no such manifest, file absent, **or the row points outside the configured root** |

`404` for the outside-root case on purpose: distinguishing it from a missing file would make
the endpoint an oracle for the catalog node's directory layout.

Headers on `200`/`206`: `Content-Type: application/octet-stream` (**must** be set explicitly —
`JsonContentMiddlewareAdapter` stamps `application/json` on every response,
`src/restapi/http_listener.go:239-242`; `src/controllers/config.go:512` is the precedent for
overriding it), `Content-Disposition: attachment; filename="<basename>"`,
`Accept-Ranges: bytes`, `Last-Modified`, and `X-Package-Checksum` / `X-Package-Size` taken
**from the manifest row**, not computed — hashing a 50 GB pack per `HEAD` is not acceptable.
The row already carries both (`catalog/models/virtual_machine_manifest.go:37-38`, written at
push time by `helpers.GetFileMD5Checksum`, `catalog/main.go:230-234`).

### Write

```
PUT  /api/v1/catalog/{catalogId}/{version}/{architecture}/pack
PUT  /api/v1/catalog/{catalogId}/{version}/{architecture}/metadata
```

- Requires `PUSH_CATALOG_MANIFEST_CLAIM` (the claim of `PUT /v1/catalog/pull`).
- `Content-Length` is required (chunked bodies refused) and must not exceed
  `CATALOG_LOCAL_MAX_UPLOAD_BYTES`; `http.MaxBytesReader` enforces it — the repo has **no**
  body limiter anywhere, so this endpoint provides its own.
- Filename is derived server-side from the identity, exactly as the client derives it
  (`getPackFilename` / `getMetaFilename`, `catalog/main.go:318-334`) — a caller can never
  choose a filename.
- The client declares `X-Package-Checksum`; the server recomputes MD5 while writing
  (`io.TeeReader`) and on mismatch deletes the temp and answers `422`. Bytes are written to
  `.<name>.part-<jobid>` inside the target folder and `os.Rename`d into place only after the
  checksum matches, so a crashed or refused upload never leaves a half file that looks like a
  pack.
- `409` when no manifest row exists yet for that triple (see §5 on ordering), `403` if revoked.
- `DELETE` is **not** exposed: `DeleteFile`/`DeleteFolder` on the HTTP provider return
  "not supported", and removal stays a catalog-node operation.

### Client side

New provider `src/catalog/providers/localhttp` registered **before** `local` in
`src/catalog/main.go:79-85`:

- `Check()` matches `provider=local-storage` **and** a resolvable `host=`; the plain `local`
  provider keeps the same-node, no-host case exactly as today. No change to
  `buildLocalCatalogConnection` (`src/controllers/machines.go:2200-2271`) is needed: the
  manifest carries `provider=local-storage`, and the merge at `pull.go:217-238` keeps the
  server's `Type` while supplying the host's own credentials.
- The provider is **path-blind**: it sends only the identity triple. That is deliberate — the
  `/download` handler obfuscates provider `Meta` for external callers (`catalog.go:615-620`),
  and `catalog_path` lives in `Meta`, so any design needing it client-side would break under
  obfuscation. It gets `Path`/`PackFile`/`MetadataFile` only as *names* to validate against.
- `CanStream()` returns `false`, keeping the dispatch path identical to today's local provider
  (`pull.go:730`, `cacheservice/main.go:869`).
- `PullFile` verifies the downloaded bytes against the checksum from the row and deletes the
  file on mismatch. This is bug 3 finally being honoured, and it is the only place in the pull
  path where it can be done without changing every other provider.

---

## 2. Trust model

The endpoint turns "read any file the process can read" into a network call, so every rule
below is load-bearing. The resolving logic lives in `src/catalog/localcatalog` (a service, not
a controller — `src/controllers/AGENTS.md`: no business logic in handlers).

| # | Rule | Why |
|---|---|---|
| R1 | The client supplies only `{catalogId, version, architecture}`. Paths and filenames come from the DB row, then are **re-derived** and cross-checked against `manifest.Path`/`PackFile`/`MetadataFile`. | A tampered or drifted row cannot become an arbitrary file read. |
| R2 | Containment is proven with `filepath.EvalSymlinks` + `filepath.Rel(root, resolved)` and `!strings.HasPrefix(rel, "..")` — never `strings.HasPrefix(a, b)`. | Kills both the `..` case and the sibling-prefix case (`/srv/catalog` vs `/srv/catalog-evil`) that the current provider accepts. |
| R3 | One authoritative root, `CATALOG_LOCAL_PATH`, configured by the operator. The row's absolute `Path` must be contained in it. | Serving whatever `catalog_path=` the connection string mentions is a remote file-read primitive. |
| R4 | `catalogId`/`version`/`architecture` are validated before use: no separator, no `..`, no NUL, matched against the same normalization the rest of the catalog applies (`helpers.NormalizeStringUpper` / `NormalizeString`, used at `pull.go:192`). | Defense in depth behind R1/R2, and it makes a hostile URL cheap to reject. |
| R5 | Tainted → `400`, revoked → `403`, mirroring `/download`. | The two endpoints must not disagree about whether an artifact is usable. |
| R6 | Authenticated principal must have a real identity: `authorizationContext.User != nil`. | `src/restapi/apikey_authorization_middleware.go:65` runs claim checks **only** when `result.UserID != ""` — an unassigned legacy key passes every claim gate. Serving multi-gigabyte images to a key with no user is not acceptable. The temp `local-catalog-*` keys set `UserID` (`machines.go:2250`), so legitimate flows are unaffected. |
| R7 | Uploads additionally: size cap, `Content-Length` required, checksum verified, atomic publish, disk-space guard before accepting (reuse the probe in `catalog/cacheservice/disk_space_unix.go`). | The only abuse surface the read path does not already have. |
| R8 | No credential or connection string in any log line; `CatalogManifestProvider.String()` gains redaction as part of fixing bug 1. | `src/catalog/AGENTS.md` treats connection strings as secrets. |

---

## 3. Gating

Three new keys, each in `src/constants/catalog_local_http.go` and catalogued in
`src/settings/catalog.go`, read through `config.Get().GetKey(...)` so the settings override
applies automatically (`src/config/main.go:317-321`):

| Key | Default | `RequiresRestart` | Notes |
|---|---|---|---|
| `CATALOG_LOCAL_HTTP_ENABLED` | `false` | **true** | Off ⇒ `registerCatalogLocalHttpHandlers` is never called: the routes do not exist, they are not "403 forever". Precedent: `IsReverseProxyEnabled()` at `src/controllers/main.go:50-52`. |
| `CATALOG_LOCAL_PATH` | empty | **true** | Enabled with an empty root ⇒ routes are **not** registered and startup logs an error. Guessing `$(cwd)/catalog` is how bug 2 got its default. |
| `CATALOG_LOCAL_MAX_UPLOAD_BYTES` | `51200` MiB | **true** | Bounds R7; 0 refuses uploads entirely. |

`RequiresRestart` is genuinely required, not decorative: `settings.InitService` runs inside
`startup.Init` (`src/startup/main.go:160`), which `cmd/api.go:34` calls *before*
`startup.InitApi()` at `cmd/api.go:85` — so a stored value is readable at registration time,
but changing it later cannot add routes to a live listener. Mirrors the placement TTL keys.

---

## 4. Layout

| Path | Contents |
|---|---|
| `src/constants/catalog_local_http.go` | the three key names + defaults (mirrors `constants/placement.go`) |
| `src/config/catalog_local_http.go` | `CatalogLocalHttpEnabled()`, `CatalogLocalPath()`, `CatalogLocalMaxUploadBytes()` (mirrors `config/placement.go`, same clamping/validation style) |
| `src/settings/catalog.go` | three new `SettingDefinition`s, orchestrator/catalog-scoped, with the restart flags |
| `src/catalog/localcatalog/` | `service.go` (resolve row → prove containment → open), `upload.go` (bounded write, tee-hash, atomic rename). Filesystem authority lives here, not in the controller. |
| `src/controllers/catalog_local_http.go` | thin handlers + `registerCatalogLocalHttpHandlers(ctx, version)` with swagger annotations |
| `src/catalog/providers/localhttp/main.go` | the HTTP client provider |
| `src/catalog/catalogtest/` | the reusable mock: temp root + in-process `httptest.Server` mounting the **real** service/handlers + a ready connection string (naming mirrors `database/testutil`) |

Keep every new file under ~500 lines and split before exceeding it.

---

## 5. Push ordering (the one real sequencing problem)

`Push` currently writes bytes **before** the manifest row exists
(`pushCreateNewManifest` sets path/filenames at `push.go:330-332`, pushes pack and meta, and
only then writes the row at `push.go:440/447`). That is fine for S3, which never consults a
row. Our `PUT` refuses unknown identities (R7 and the `409`), so the localhttp path must
register first.

Approach: a **capability interface asserted locally**, not a new method on
`RemoteStorageService` — the same pattern used for the placement/database seam, so no provider
is forced to implement it:

```go
// manifestFirstUploader is what the HTTP provider reports: it cannot accept
// bytes for an identity the catalog has never heard of.
type manifestFirstUploader interface{ UploadRequiresManifestFirst() bool }
```

`pushCreateNewManifest` / `pushUpdateManifest` check the assertion and, when true, perform the
manifest registration (the `POST` at `push.go:419-420`, or the local DB write) before calling
`PushFile`. Existing providers take the unchanged path. The alternative — accepting orphan
uploads that a later registration adopts — is rejected: it makes unbounded storage reachable
before any authorization beyond the claim, and it leaves garbage when the registration never
comes.

---

## 6. The four fixes

| Bug | Fix | Regression test |
|---|---|---|
| 1 `String()` inversion | emit every non-empty field; omit empty ones; keep `Meta` last; redact `password`/`api_key` in the two log sites (`push.go:421`, `import_vm.go:136`) | round-trip `Parse` → `String` → `Check` for each provider family, asserting a credential-bearing string still selects the intended provider |
| 2 local containment | one private `resolve(path, name) (string, error)` used by all ten methods (R2 rules), refuse absolute escapes, no silent re-rooting; `catalog_path` mandatory — refuse with a clear error instead of `$(cwd)/catalog`; directories `0750`, files `0640` | traversal table: `../../etc/passwd`, sibling-prefix root, absolute `path`, symlink out of root — each refused, and nothing outside the root touched (assert by scanning the temp tree) |
| 3 checksum on pull | client provider recomputes MD5 of what it received and compares with the row's `CompressedChecksum`; mismatch deletes the file and fails the step | serve a corrupted body from `catalogtest` and assert failure; serve a truncated body and assert no partial VM folder survives |
| 4 `Check()` error aborts the loop | `pull.go:310-314` and `main.go:95-98`: record at debug level and continue to the next provider | a connection string that makes one provider error must still reach and match a later one |

---

## 7. Test strategy — this is the mock system

Everything below runs with `go test`, no docker, no cloud, no network.

1. **Conformance suite** — `catalog/tester` gains `RunConformance(t, connection)` covering all
   sixteen `RemoteStorageService` methods as a table, run against the `local` provider on a
   `t.TempDir()` root and against `localhttp` through `catalogtest`. One table, two providers:
   any new provider has to pass the same suite. This is what `cmd test-providers` does today
   against live S3, finally runnable offline.
2. **Traversal table** for the local provider and the HTTP service (bug 2).
3. **`catalogtest`** — `New(t) *Catalog` with `Root()`, `Handler()`, `Connection()`,
   `SeedManifest(t, fields)`, `PutPack(t, bytes)`, plus `WithAuth(claims...)` so callers can
   exercise the claim and unassigned-key paths. Reusable by `controllers`, `orchestrator`,
   `pdfile` and `tests` — the seed of the wider mock system.
4. **Controller tests** (`controllers/catalog_local_http_test.go`): routes **absent** when the
   flag is off (assert the handler is never registered, not merely unauthorized); `401`, claim
   denial, unassigned-key denial (R6); tainted/revoked parity with `/download`; `Content-Type`
   is not `application/json`; `Range` yields `206`.
5. **End-to-end, zero setup** (`catalog/e2e_local_test.go`): build a tiny fake `.pvm`
   (`config.pvs`, `NVRAM.dat`, one `*.hdd` — enough for `cacheservice/integrity_check.go:11-43`),
   `Push` it with `provider=local-storage`, then `Pull` it back through `localhttp` and assert
   the VM folder, the checksum and the flattened layout (`common.CleanAndFlatten`). This single
   test replaces the cloud-credential requirement for the local case.
6. **Provider registration order** test: with `host=` present the HTTP provider wins, without
   it the filesystem provider still wins — so the same-node path cannot be silently diverted
   onto HTTP.
7. **`cmd test-providers`** gains a credential-free leg that runs the local + localhttp
   conformance always, keeping the cloud legs env-gated as today.

Known cost: five hardcoded `time.Sleep(2 * time.Second)`s sit in the pull path
(`pull.go:167, 361, 380, 395, 401`), so the e2e test carries ~10 s. Removing them is out of
scope here; if it becomes annoying, replace them with a package-level test seam rather than
deleting behaviour.

---

## 8. Rollout, docs, verification

- Off by default; the endpoints do not exist unless the operator opts in. `false` in
  `settings.Catalog()` means an existing installation changes nothing on upgrade.
- Docs: new page `docs/docs/devops/catalog/local-catalog.md` (setup, the three keys, the wire
  contract, the security model), a pointer from `docs/docs/devops/catalog/` concepts, and a
  line in `docs/docs/devops/settings/overview.md` for the new keys. `CHANGELOG.md` entry under
  1.2.0.
- `make generate-swagger` is the user's to run (needs docker); annotations are written for it.
- Gates: `gofmt -l`, `go vet ./...`, `go test -count=1 ./...`, `-race` on
  `./catalog/... ./controllers/ ./config/ ./settings/`, and `go test -count=5 -shuffle=on` on
  any touched `./database/...` package.

## 9. Risks

- **Claim tightening (decision 1)** can break a deployment whose VM-creating principal lacks
  `PULL_CATALOG_MANIFEST_CLAIM`. Mitigation: the startup log line and the docs name the exact
  claim, and the `403` body says which claim was missing.
- **R6 rejects unassigned legacy keys.** Correct by design, but it is a behaviour change for
  anyone using a user-less key against the catalog — worth an explicit changelog sentence.
- **Row-vs-disk drift**: if `CATALOG_LOCAL_PATH` differs from the `catalog_path` used at push
  time, every read is `404`. Mitigation: the resolution failure logs the resolved path and the
  configured root (path only, never credentials) so it is diagnosable in one line.
- **MD5 as the integrity primitive** is inherited (`GetFileMD5Checksum`), not chosen. It is
  fine for corruption detection and is what the row stores; upgrading it is a separate change.
