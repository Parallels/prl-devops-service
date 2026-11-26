# data

Three things live here; know which one you are touching.

- `models/` — persistence models, with GORM tags. Used by the live SQL layer. Adding a table/column starts here, then register it in `../database/migrate.go` (`AllGormModels`). `shape_security_test.go` fails if secrets leak through model shapes; `relationships_parity_test.go` / `automigrate_test.go` guard schema drift (`models/testdata/schema_snapshot.sql` is the snapshot; regenerate deliberately).
- `interfaces/` — `DatabaseService` and the per-domain interfaces and sentinel errors (`errors_*.go`). This is the contract callers depend on. Get the service with `serviceprovider.GetDatabaseService(ctx)`.
- Package `data` (root `*.go`) — the **deprecated JSON-file engine** (`JsonDatabase`, file locks, backups, `*_layer.go` filters). Scheduled for removal in v1.3.0 (#426). Do not add features or new callers; it survives only as the oracle for `../database/parity` tests and for `legacy_load.go` migration input.

Rules: new code depends on `data/interfaces`, never on package `data` or on GORM directly outside `../database`. Keep model fields and API DTOs (`../models`) separate; convert in `../mappers`.
