# Brief — local catalog served over HTTP (catalog mock system, step 1)

## What the user asked for

Ways of testing changes as complex as the placement work **without complex setups**. This is
the first step toward a more complex mock system. The concrete shape requested: expose the
catalog so machines can be **stored locally** and expose **endpoints for a host to connect to
and download them**, exactly as it would from any other catalog. It must stay secure and the
endpoints must **not exist at all** unless enabled by an environment variable or the settings
service. The existing `local` provider "does not work as expected".

## How it works today (verified, with citations)

A catalog pull uses **two transports**, and only one of them is networked.

| Step | Transport | Where |
|---|---|---|
| Manifest metadata | HTTP `GET /api/v1/catalog/{id}/{ver}/{arch}/download` when `Provider.IsRemote()` | `src/catalog/pull.go:179-242` |
| **Pack bytes** | the storage provider's `PullFile(Path, PackFile, dest)` | `src/catalog/pull.go:761`, `:789` |

Providers are chosen by the first `Check(ctx, manifest.Provider.String())` that returns true,
in registration order s3 → **local** → azure → artifactory → minio
(`src/catalog/main.go:79-85`). There is **no endpoint in the repository that streams pack
bytes from disk** — zero hits for `http.ServeFile`, `http.ServeContent`, `Content-Disposition`;
`/download` writes JSON only (`src/controllers/catalog.go:625-626`).

Consequence for `provider=local-storage`: a remote host fetches the manifest over HTTP and
then calls the local provider, which does `os.Open(filepath.Join(manifest.Path,
manifest.PackFile))` **on its own filesystem** — while `manifest.Path` is an absolute path on
the *catalog node* (`src/catalog/push.go:330`:
`manifest.Path = filepath.Join(rs.GetProviderRootPath(ctx), manifest.CatalogId)`). The bytes
are unreachable; the failure is a confusing 404, or worse, reading an unrelated file that
happens to live at that path on the pulling machine.

Why testing needs credentials today: the conformance harness `src/catalog/tester` is driven
only from `src/tests/catalog_provider_testing.go` via `cmd test-providers`, which runs a
provider only when `AWS_S3_TEST_CONNECTION` / `AZURE_SA_TEST_CONNECTION` /
`ARTIFACTORY_TEST_CONNECTION` is set. There is no Go test that exercises push→pull for any
provider.

## Bugs found while tracing (each blocks or undermines this feature)

1. **`CatalogManifestProvider.String()` has inverted emptiness checks**
   (`src/catalog/models/catalog_manifest_provider.go:19-33`): `if m.Host == "" { r += ";host="
   + m.Host }` — same for port/user/password/api_key. So the string handed to every
   `Check()` (`src/catalog/pull.go:309`) **drops** the first-class credential and host fields.
   Cloud providers survive only because their keys land in `Meta`. An HTTP client provider
   cannot recover its host/credentials without this fixed. Its two log call sites
   (`src/catalog/push.go:421`, `src/catalog/import_vm.go:136`) would then print secrets, which
   `src/catalog/AGENTS.md` forbids — so redaction ships with it.
2. **Local provider containment is a string-prefix test** (`src/catalog/providers/local/main.go`
   :110, :147, :169, :210, :239, :251, :259, :282, :293, :308): `..` survives
   `filepath.Join`+`Clean`, and a sibling root (`/srv/catalog` vs `/srv/catalog-evil`) passes
   `HasPrefix`. Root silently defaults to `$(cwd)/catalog` created `os.ModePerm` (0777) when
   `catalog_path=` is absent.
3. **No checksum is ever verified on pull.** `FileChecksum` is called exactly once, to *name*
   the cache entry (`src/catalog/cacheservice/main.go:115`); `CompressedChecksum` is never
   compared and is dropped by the API mapper (`src/mappers/catalog.go:396-457`). Contradicts
   `src/catalog/AGENTS.md`: "verify checksums on pull".
4. **A `Check()` error aborts the whole provider loop** (`src/catalog/pull.go:310-314`) instead
   of continuing to the next provider, so one malformed connection family is a hard failure.

Deliberately **out of scope** (found, documented, not fixed here): the file-type cache
`CopyDir` bug (`pull.go:715` vs `helpers/os.go:381`), artifactory `PullFileAndDecompress`
never decompressing, version/architecture being ignored during manifest selection
(`pull.go:185-192`, `:249`), ETag-as-checksum semantics, the hardcoded `time.Sleep(2s)`s in
the pull path.

## Decisions taken (by the user, binding)

1. **Read gate**: byte endpoints require `PULL_CATALOG_MANIFEST_CLAIM` — stricter than today's
   `/download` (which needs only `LIST_...`) and matching `PUT /v1/catalog/pull`. Accepted
   risk: a principal that can create VMs but lacks PULL breaks on this path.
2. **Write side is in scope**: push is exposed over HTTP as well, not read-only.
3. **Scope**: the feature plus the four bugs it trips over. Each gets a regression test.

## Non-goals

- No new storage format, no change to what a `.pdpack` contains.
- No chunked/parallel download for the HTTP provider in this step (`CanStream() == false`),
  which keeps the code path identical to today's local provider. The chunked manager
  (`src/catalog/chunkmanagerservice`) assumes tar.gz and unclamped ranges
  (`chunkmanagerservice/service.go:322-323`) — revisit only when a real need appears.
- No replacement of the catalog-manager concept; the new endpoints are another way to reach a
  catalog that already exists, and the manager proxy
  (`src/controllers/catalog_managers.go:853-874`, `:1261-1399`) stays as is.
- No cloud-provider behaviour changes beyond the shared serialization fix.
