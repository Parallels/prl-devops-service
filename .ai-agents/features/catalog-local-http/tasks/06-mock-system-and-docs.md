# 06 — Mock system (`catalogtest`), conformance suite, end-to-end test, docs

**Depends on:** 03, 04, 05. This is the card that pays off beyond this feature: it is the first
piece of the wider mock system, and it is what removes cloud credentials from catalog testing.

## `src/catalog/catalogtest` — reusable fixture (naming mirrors `database/testutil`)

```go
type Catalog struct { /* root, srv *httptest.Server, svc, db */ }

func New(t *testing.T, opts ...Option) *Catalog        // temp root + feature on + in-process server
func (c *Catalog) Root() string                        // the directory the service serves
func (c *Catalog) URL() string                          // http://127.0.0.1:port
func (c *Catalog) Connection() string                    // ready host=<apikey@url> for a provider
func (c *Catalog) Handler() http.Handler                 // real handlers, mountable in httptest
func (c *Catalog) SeedManifest(t, f ManifestFields) *Seed // insert a row + place the bytes
func (c *Catalog) WithPrincipal(p Principal) *Catalog     // claims / unassigned key / super user
```

- Mounts the **real** controller + service code, not a re-implementation — a mock that can drift
  from the server is worse than no mock. Options: `WithRoot`, `WithUploadCap`, `WithAuth`,
  `WithoutFeature` (so the absence test in card 04 uses the same fixture).
- Owns the test database through the same helper the controller tests already use
  (`setupTestDBForController`, `src/controllers/machines_local_catalog_test.go:97`) so fixtures
  stay one flavour.
- No cloud SDK import in this package — the point is that it runs anywhere, including CI without
  secrets.

## Conformance suite in `src/catalog/tester`

`RunConformance(t *testing.T, connection string, root string)` — the checks in
`src/catalog/tester/main.go` (`Check`, create/exists/delete folder, push, pull, checksum, size,
delete file) restructured into `t.Run` cases over the 16 `RemoteStorageService` methods, with the
`TESTING`/`testing.txt` constants kept. Run it twice:

1. against the `local` provider on `t.TempDir()`;
2. against `localhttp` through `catalogtest`.

One table, any provider: a new provider must pass the same suite to be worth adding.

## End-to-end, zero setup — `src/catalog/e2e_local_test.go`

Build a tiny fake `.pvm` on disk (`config.pvs`, `NVRAM.dat`, one `*.hdd` — enough to satisfy
`src/catalog/cacheservice/integrity_check.go:11-43`), `CatalogManifestService.Push` it with
`provider=local-storage`, then `Pull` it back through `localhttp` against `catalogtest`, and
assert: the VM folder exists at the destination, the flattened layout is right
(`common.CleanAndFlatten`), the checksum matched, and a deliberately corrupted pack fails the
pull with no partial folder left behind.

Today the equivalent needs a real S3/Azure/Artifactory bucket and a `*_TEST_CONNECTION`
(`src/tests/catalog_provider_testing.go`). That is the "complex setup" this card deletes.

## `cmd test-providers`

Add an always-on leg that runs the local + localhttp conformance with no credentials; keep the
cloud legs gated on `AWS_S3_TEST_CONNECTION` / `AZURE_SA_TEST_CONNECTION` /
`ARTIFACTORY_TEST_CONNECTION` exactly as now.

## Docs and changelog

- New `docs/docs/devops/catalog/local-catalog.md`: what it is, the three keys and that the
  endpoints do not exist when off, the wire contract table, the security model (path-blind,
  containment, `PULL`/`PUSH` claims, principal requirement), and a "test without a cloud account"
  section pointing at `catalogtest`.
- Pointer from the existing catalog concepts page; the three keys added to
  `docs/docs/devops/settings/overview.md`; `CHANGELOG.md` under 1.2.0 with an explicit line that
  user-less API keys are rejected on the new endpoints.

## Definition of done

`go test ./catalog/... ./controllers/...` green offline, no credential in sight; the e2e test
demonstrably exercises push→pull across two "nodes" in one process; the conformance suite runs
against both providers; docs and changelog updated.
