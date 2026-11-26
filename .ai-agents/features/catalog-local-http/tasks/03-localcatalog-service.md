# 03 — `localcatalog` service: resolve, contain, open, upload

**Depends on:** 01, 02. **Blocks:** 04, 05, 06.

The filesystem authority for the HTTP surface. Deliberately a package, not controller code
(`src/controllers/AGENTS.md`: no business logic in handlers), and the single place that decides
whether a byte range of the catalog is reachable.

## Package `src/catalog/localcatalog`

```go
type Artifact int
const (ArtifactPack Artifact = iota; ArtifactMetadata)

type Resolved struct { Manifest *data_models.CatalogManifest; FullPath string; Size int64; Checksum string; ModTime time.Time }

type Service struct { /* db accessor, root, maxUploadBytes */ }

func New(ctx basecontext.ApiContext, root string, maxUploadBytes int64) (*Service, error)
func (s *Service) Resolve(ctx basecontext.ApiContext, catalogID, version, arch string, a Artifact) (*Resolved, error)
func (s *Service) Put(ctx basecontext.ApiContext, catalogID, version, arch string, a Artifact, body io.Reader, declaredChecksum, jobID string) (*Resolved, error)
```

### `Resolve` — order matters, and every rejection is a distinct internal error

1. Validate the triple (plan R4: no separator, no `..`, no NUL; normalise the way `pull.go:192`
   does) **before** touching the database, so a hostile URL is cheap to reject.
2. Load the row by `(catalogId, version, architecture)` — the same call `/download` uses
   (`dbService.GetCatalogManifestsByCatalogIdVersionAndArch`, `src/controllers/catalog.go:574`),
   so the two endpoints cannot disagree about which row an identity means.
3. Tainted / revoked surfaced as typed errors the handler maps to `400` / `403` (plan R5).
4. Derive the expected filename from the identity (`getPackFilename` / `getMetaFilename`
   semantics, `src/catalog/main.go:318-334`) and compare with the row's `PackFile` /
   `MetadataFile`; a mismatch is a `404`-class error — the row has drifted from the naming rule
   and we will not serve "whatever the row happens to say".
5. Containment (plan R1–R3): resolve under `s.root` with `EvalSymlinks` + `filepath.Rel`,
   never a prefix test. Anything outside is the **same** not-found error as a missing file, so
   the endpoint is not an oracle for the node's directory layout.
6. `os.Stat` for size/`Last-Modified`; checksum from the row's `CompressedChecksum` — never
   computed here, because a `HEAD` must not hash a 50 GB pack.

### `Put`

- `Content-Length` known and ≤ `maxUploadBytes`; wrap the body in `http.MaxBytesReader` at the
  handler and treat "limit exceeded" as a clean `413`, not a panic.
- Write `.<name>.part-<jobID>` in the target folder, `io.TeeReader` into `md5.New()`, compare
  with `declaredChecksum`; on mismatch `os.Remove` the temp and return a `422`-class error.
- Publish with `os.Rename` (same filesystem, atomic) and `0640`. A crashed upload therefore
  never leaves something that looks like a pack.
- Refuse revoked manifests; refuse when no row exists — that is the `409` card 05 handles.
- Disk-space guard before accepting any bytes (reuse the probe behind
  `src/catalog/cacheservice/disk_space_unix.go`).

## Tests (`service_test.go`, `upload_test.go`)

- `Resolve` table: happy pack and metadata; unknown triple; tainted; revoked; row names
  mismatch; symlink planted outside the root; sibling-prefix root (`root + "-evil"`); row whose
  `Path` is absolute-outside-root ⇒ same not-found error as a missing file (assert the two
  errors are indistinguishable).
- `Put`: checksum mismatch leaves **no** file; success is atomic (target absent until rename);
  oversized body refused before writing; revoked refused; missing row refused with the distinct
  error; `..` in the triple never reaches the filesystem.
- Everything on `t.TempDir()`, no network, no `serviceprovider` singleton beyond the injected
  database handle.

## Definition of done

No path decision exists outside this package; `grep -n "HasPrefix" src/catalog/localcatalog`
contains no containment check; all tests pass with `-race`.
