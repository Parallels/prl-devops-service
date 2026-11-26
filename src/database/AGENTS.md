# database

The live persistence layer: `GormDatabase` implements `data/interfaces.DatabaseService` on SQLite (pure-Go `glebarez/sqlite`, keeps the build CGO-free) or Postgres. Selected by `DB_TYPE`, `DB_PATH`, `DB_DSN`; `DB_AUTO_MIGRATE` runs `RunAutoMigrate` (`migrate.go`) over `AllGormModels()`.

Layout: one file per domain (`users.go`, `jobs.go`, `catalog.go`, `orchestrator*.go`, `snapshots.go`, ...), `connection.go` / `config.go` (engine setup), `health.go`, `errors.go` (driver error -> interface sentinel mapping), `roworder.go` (deterministic row ordering on both dialects), `legacy_migrator/` (JSON -> SQL import), `parity/` (two-engine tests), `testutil/` (test harness).

Working here:
- Implement against the interface in `../data/interfaces`; return its sentinel errors, not raw GORM errors. Never import package `data` outside `_test` scopes (parity is the only exception).
- New table/column: edit `../data/models`, register in `AllGormModels`, refresh `data/models/testdata/schema_snapshot.sql`, and make it work on SQLite **and** Postgres.
- Don't use the `serviceprovider` singleton or `config.Get()` inside this package; pass config via the `FromApplicationConfig` seam. Tests: one in-memory DB per test via `testutil`, no `t.Parallel()` where `t.Setenv` is used.
- Keep results ordered explicitly (see `roworder.go`); unordered queries behave differently across engines.
- Gates: `go test -count=5 ./database/...` and `-shuffle=on` must pass; Postgres tests run with `make test-integration`.
- Never log DSNs or passwords. Docs: `docs/docs/devops/getting-started/postgresql.md`, `database-upgrade.md`; design history in `.ai-agents/features/417-gorm-database-layer/`.
