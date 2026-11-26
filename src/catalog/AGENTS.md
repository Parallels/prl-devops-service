# catalog

VM image catalog: push a VM to remote storage with a manifest, pull it back, import, delete. Used by the catalog module (host side) and the catalog manager (index of manifests).

- `main.go` — `CatalogManifestService`: owns the set of `remoteServices`. `push.go`, `pull.go`, `import.go`, `import_vm.go`, `delete.go`, `push_metadata.go` are the operations; progress goes through `jobs/tracker`.
- `providers/` — storage backends implementing `interfaces.RemoteStorageService`: `local`, `aws_s3_bucket`, `azurestorageaccount`, `artifactory`, `minio`. To add one: new subpackage implementing the interface, register it in the manifest service, add `tests/catalog_provider_*` coverage, and document the connection string in `docs/docs/devops/catalog/`.
- `cacheservice/` — local cache of pulled images (integrity check, cleanup, disk-space guards). `chunkmanagerservice/`, `cleanupservice/` — chunked transfer and cleanup helpers.
- `models/` — manifest and request types.
- Rules: connection strings and credentials are secrets, never log them; verify checksums on pull; operations must be cancellable and report progress.
- Related: `../pdfile` drives catalog pull/push from PDFiles. Docs: `docs/docs/devops/catalog/` (concepts, overview, pdfile).
