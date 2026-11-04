# PBI Brief #424 — One-time data migration script (legacy → GORM)

| Field | Value |
|---|---|
| PBI | [#424](https://github.com/Parallels/prl-devops-service/issues/424) — `[PBI] One-time data migration script (legacy → GORM)` |
| Labels | `triage`, `pbi` · Assignee: unassigned |
| Round | Deep-dive round **5** |
| Upstream feature docs | [`brief.md`](../brief.md) · [`plan.md`](../plan.md) — required reading, esp. AD-4, AD-5, OQ-1 |
| Planning status | **Planned** (round 5, 2026-11-03): [`plan.md`](./plan.md) + 7 tasks in [`tasks/`](./tasks/) — per the [loop](../plan.md#10-the-per-pbi-deep-dive-loop-repeat-per-pbi-in-order). Awaiting human review (OQ-A…OQ-E in plan.md); task 03 onward also blocked on #419's merged mapping-table document. |

## Description (verbatim from GitHub)

Implement a migration script that detects the presence of the legacy data store on first run and
converts all existing data into the new GORM-managed schema. The migration must be automatic,
idempotent, and safe to run in production.

### User Story

As an **existing user upgrading the service**, I want my data to be automatically migrated to the
new database format on first start after upgrading, so that I do not lose any data and do not need
to perform manual migration steps.

### Acceptance Criteria

- [ ] **Detection**: On startup, the service checks for the presence of the legacy data store. If not found, migration is skipped silently.
- [ ] **Idempotent**: Running the migration multiple times produces no errors and no duplicate data.
- [ ] **Atomic**: The migration runs inside a transaction; if any step fails, all changes are rolled back and the service exits with a clear error.
- [ ] **Progress logging**: Migration progress is logged at the `info` level so operators can monitor long-running migrations.
- [ ] **Completion marker**: After a successful migration, a marker is written (e.g. a flag in the new schema or a sentinel file) so the migration does not re-run on subsequent startups.
- [ ] **Data integrity**: A post-migration validation step counts records in the legacy store vs. the new store and logs a warning if counts differ.
- [ ] **Dry-run mode**: A `--migrate-dry-run` flag (or equivalent env var) runs the migration logic and reports what would be changed without writing any data.

### Definition of Done

- [x] Code implemented following best practices.
- [x] Unit tests written and passing.
- [x] Code reviewed and approved.
- [x] Merged into the main branch.
- [x] Documentation updated (if applicable).
- [x] Deployed to staging/production environment.

### Assumptions and Constraints

- **Assumption**: The legacy data store and the new database can coexist on the same host during the migration window. ✅ Guaranteed by AD-3 (same folder, different filenames: `data.json` vs `prl-devops.db`).
- **Constraint**: Migration must complete without downtime for single-host deployments (runs as part of the normal startup sequence, not as a separate offline process).

### Dependencies

Unstated on GitHub; locked by [feature plan §8](../plan.md#8-delivery-order--dependency-graph):
**#419 (mapping table document) + #421 (query layer to write through) merged.** #420 provides the
`DB_AUTO_MIGRATE` gate.

## Repo context (verified 2026-11-03)

- **Migration source of truth**: the legacy `Data` struct (`src/data/main.go:33-51`) — 16
  collections including pointers (`Configuration`, `ReverseProxy`) and embedded relationships.
- **AD-5**: read the file through the legacy loader (`JsonDatabase.Load`/`loadFromFile` handles
  optional `EncryptionPrivateKey` decryption, residual `.save`/`.panic` recovery) — never hand-parse
  `data.json`. A constructor variant that loads without starting the save/backup goroutines is the
  clean way (legacy code is frozen except additive helpers approved in this PBI's tasks).
- **Detection**: file existence of `<db-folder>/data.json` **plus** non-empty content (empty file =
  fresh install, skip silently; AC 1). Note legacy quirk: `IsDataFileEmpty` *creates* the file when
  missing (`src/data/main.go:645-655`) — detection must not reuse that side-effecting helper.
- **Startup placement**: after legacy schema migrations and GORM AutoMigrate — current sequence
  `startup.Initialize` runs `migrations.Version0_6_0`, `migrations.AddApiKeyOwnClaims`
  (`src/startup/main.go:54-56,135-140`), then module init. New step slots between, before the API
  serves. **Never modify `data.json`** — legacy store stays intact for fallback/deprecation window
  (OQ-1: proposal is "frozen after migration"; fallback start logs a warn — decision at #426 round).
- **Marker**: a row in a `schema_migrations`-style table **in the new database** (AD-4), holding
  legacy source path + file size + completion timestamp; guards AC 2/6. Consider also a content
  fingerprint to detect re-pointing at a *different* legacy file (plan decision).
- **Validation counts (AC 6)**: legacy count = `len(Data.<Collection>)` in memory; new count =
  GORM `Count()` per table, join-table-aware (e.g. `user_roles` rows = Σ len(u.Roles)); the
  #419 mapping document defines the expected formula per collection — warn-only per AC.
- **Transaction (AC 3)**: one GORM `Transaction` per collection batch (whole-DB single tx risks
  SQLite lock/timeout on large stores) — the plan must pick and justify; idempotency via upsert-by-ID
  (`clause.OnConflict{ID}`), not delete-all-then-insert.
- **Dry-run flag**: project flag pattern = constants in `src/constants/main.go` read via
  `helper.GetFlagValue(...)` (see `CONFIG_FILE_FLAG` usage in `src/config/main.go:122`); mirror for
  `--migrate-dry-run` + `DB_MIGRATE_DRY_RUN` env.
- Large-store caution (risk R-7): log per-collection progress at info; keep memory bounded by
  chunked inserts; the legacy loader already materialises everything in RAM — acceptable since the
  legacy store's ceiling is the same one the service runs under today.

## Engineering-rule triggers (feature plan §6)

- New package `src/database/legacy_migrator/` (files ≤ ~200 lines: `detector.go`, `migrator.go`,
  `mapping.go`, `validation.go`, `marker.go`, `dryrun.go`).
- ER-1: `basecontext` + `errors.NewWithCode` + `ctx.LogInfof` progress per AC 4.

## Fit in the feature

- **Depends on**: #419 mapping document (schema truth), #421 write paths, #420 (`DB_AUTO_MIGRATE`).
- **Feeds**: #425 (migration edge-case tests: empty/partial/re-run/encrypted), #427 (upgrade guide).
- **Definition-of-done link**: first-run upgrade on a real legacy install (fixture generated by
  running an older release or checked-in golden `data.json`) must be exercised in CI
  (`//go:build integration`).
