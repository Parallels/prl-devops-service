# 05 — HTTP client provider (`providers/localhttp`)

**Depends on:** 02, 03, 04. **Blocks:** 06.

Makes a remote host pull from — and push to — a catalog node through the new endpoints, while a
same-node pull keeps using the filesystem provider untouched.

## Steps

1. New subpackage `src/catalog/providers/localhttp` implementing
   `interfaces.RemoteStorageService` (16 methods, `src/catalog/interfaces/remote_storage_service.go`),
   registered **before** `local.NewLocalProvider()` in `src/catalog/main.go:79-85`.
2. `Check(ctx, connection)`: true only for `provider=local-storage` **and** a resolvable
   `host=` — parse with `CatalogManifestProvider.Parse` so `apikey@host:port` and
   `user:pass@host` both work, exactly as `pull.go:173` already does for the manifest fetch.
   The ordering plus the `host=` condition are what keep the same-node case on the filesystem
   provider; without the `host=` test this provider would hijack every local pull.
3. **Path-blind.** Send only `{catalogId, version, architecture}` in the URL. Never use
   `manifest.Path` in a request: it is an absolute path on the catalog node, and its
   `catalog_path` meta is obfuscated for external callers anyway
   (`src/controllers/catalog.go:615-620`). The interface's `path`/`filename` arguments are
   validated against the naming rule and otherwise ignored.
4. Auth: reuse `GetAuthenticator(provider)` (`src/catalog/common.go:8-16`) so the header shape
   matches what `pull.go:183` already sends against `/download`. Build a dedicated
   `http.Client`: the shared apiclient's 5-minute default
   (`src/serviceprovider/apiclient/main.go:29-30`) is far too short for a multi-gigabyte pack,
   and the server's write budget is 5 h (`src/restapi/http_listener.go:452-455`). No overall
   `Timeout`; a `ResponseHeaderTimeout` and dial timeout instead, so a slow bulk transfer is
   never mistaken for a dead one.
5. Method mapping: `FileExists`/`FileSize`/`FileChecksum` → `HEAD` reading
   `X-Package-Checksum` / `Content-Length` / status; `PullFile` → `GET` to
   `filepath.Join(rootDestination, filename)`; `PullFileAndDecompress` → `GET` + the same
   decompress step the local provider uses; `PullFileToMemory` → `GET` with the existing size
   guard kept honest.
6. **Checksum verification (bug 3).** After the body lands, recompute MD5 and compare with the
   `X-Package-Checksum` the server sent from the manifest row; on mismatch delete the file and
   fail the step. This is the pull path finally honouring `src/catalog/AGENTS.md`.
7. `CanStream() == false`. Chunked streaming needs the server to honour ranges the way the chunk
   manager expects (`chunkmanagerservice/service.go:322-323` sends unclamped ranges) and assumes
   tar.gz — out of scope, and `false` keeps dispatch identical to today's local provider
   (`pull.go:730`, `cacheservice/main.go:869`).
8. `UploadRequiresManifestFirst() bool { return true }` — a capability asserted locally by
   `pushCreateNewManifest` / `pushUpdateManifest` (`src/catalog/push.go:326-447`), which then
   registers the manifest (the `POST` at `:419-420` or the local DB write) **before** calling
   `PushFile`. Not added to `RemoteStorageService`, so the five existing providers are untouched.
   `PUT` is used for `PushFile`; `DeleteFile`/`DeleteFolder`/`CreateFolder` return a clear
   "not supported over http" error.
9. Failures map to the errors the pull path already understands (so the job reports a useful
   message rather than a raw `*url.Error`); never log the connection string or the secret.

## Tests — `providers/localhttp/main_test.go` against a real `httptest.Server` mounting the
real `localcatalog` service (via `catalogtest`, card 06)

- Registration order: with `host=` this provider wins; without it, `local` still wins.
- `HEAD` → `FileSize`/`FileChecksum` return the row's values without reading the pack.
- `PullFile` success writes the file; **corrupted body ⇒ error and no file left**; truncated
  body likewise.
- `403` and `404` surface as distinct, useful errors; a `401` says authentication failed rather
  than "not found".
- `PushFile` before the row exists ⇒ the `409` path; after ⇒ file published, no `.part`.
- Unsupported verbs return the documented error, not a panic.
- All under `-race`.

## Definition of done

A remote host can pull a `local-storage` manifest that lives only on the catalog node, verified
by an end-to-end test; the same-node path is provably unchanged.
