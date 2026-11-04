# Feature Brief — Database Layer Migration to GORM (SQLite / PostgreSQL)

> **Read me first.** This is the canonical reference for every agent working on any task in this
> feature. Read this brief, then [`plan.md`](./plan.md), then your PBI's own `brief.md` and
> `plan.md`, then your task spec. Never start implementation without all four.

| Field | Value |
|---|---|
| Source issue | [#417 — [FEATURE] Upgrade data layer to SQL](https://github.com/Parallels/prl-devops-service/issues/417) |
| Feature folder | `.ai-agents/features/417-gorm-database-layer/` |
| Status | Planning complete for PBI #418 · deep-dive rounds pending for the rest |
| Repository | `Parallels/prl-devops-service` — Go service, module `github.com/Parallels/prl-devops-service`, sources under `src/` (Go 1.25) |
| Last verified against code | 2026-11-03 |

---

## 1. Why (problem statement, from #417)

The current persistence layer is a bespoke implementation that mixes a MongoDB-style in-memory
document model with direct whole-file disk writes. It blocks product evolution:

- **Scalability** — cannot support horizontal scaling; the service cannot run reliably as a
  microservice in orchestrated environments (Kubernetes, Docker Swarm).
- **Maintainability** — a custom database abstraction increases bug surface, needs specialised
  knowledge, and lacks the hardening an established ORM provides.
- **Flexibility** — no supported path today for production deployments against a robust
  relational database.

## 2. What we build

Replace the custom data layer with a **GORM-based abstraction** supporting two backends:

- **SQLite (default)** — file-based storage with a structured SQL schema; keeps single-host
  simplicity. The existing database *folder* path convention is preserved (see §5).
- **PostgreSQL** — production-grade, stateless-microservice deployments.

Delivery properties required by the issue:

1. All persisted structs get GORM tags; the data-access layer is refactored to GORM queries.
2. Backend selected via env vars or config file; SQLite default requires zero configuration.
3. On first run after upgrade, legacy data store is auto-detected and migrated **once**,
   **idempotently**, with no manual steps.
4. Legacy data layer **stays** in the codebase, formally deprecated, removed in a future release.
5. Unit tests updated for the new layer; coverage for both backends and the migration path.
6. User-facing docs updated (config, migration behaviour, backends, upgrade guidance).

## 3. Configuration reference (target state — owned by PBI #420)

| Setting | Env var | Config key | Default |
|---|---|---|---|
| Database backend | `DB_TYPE` | `database.type` | `sqlite` |
| SQLite file path | `DB_PATH` | `database.path` | `<current default path>` (see §5) |
| PostgreSQL DSN | `DB_DSN` | `database.dsn` | _(empty)_ |
| Run migration on start | `DB_AUTO_MIGRATE` | `database.auto_migrate` | `true` |

Existing keys that keep working unchanged: `DATABASE_FOLDER`, `DATABASE_NUMBER_BACKUP_FILES`,
`DATABASE_BACKUP_INTERVAL_MINUTES`, `DATABASE_SAVE_INTERVAL_MINUTES`,
`SYSTEM_AUTO_RECOVER_DATABASE` (legacy-store controls), and `LOG_LEVEL` (drives GORM logging).

## 4. Glossary

| Term | Meaning |
|---|---|
| **Legacy store** | The `<db-folder>/data.json` file — one encrypted (optionally) JSON document holding every collection; whole-file atomic rewrites on a save interval. |
| **`JsonDatabase`** | The legacy implementation type in `src/data` (package `data`). 157 public methods, in-memory `Data` struct + file writes. **Will be deprecated, not removed.** |
| **`GormDatabase`** | The new GORM-backed implementation of the data-access API (introduced by PBI #421). |
| **`DatabaseService`** | New interface capturing the data-access API so both implementations are interchangeable (strangler pattern; see plan AD-1). |
| **Legacy migration** | One-time data conversion `data.json → SQL schema`, PBI #424. Runs at startup, idempotent, transactional. |
| **Completion marker** | Row in a table in the **new** database recording that the legacy migration ran; makes re-runs a no-op. |
| **Strangler selection** | Runtime choice at service-provider init of which `DatabaseService` implementation is live, driven by `DB_TYPE`. |

## 5. Verified repo context map (as measured 2026-11-03)

These facts are already verified — trust them; re-verify only if git log shows movement.

| Fact | Where |
|---|---|
| Go module root, `go.mod` (no GORM yet; `go-sql-driver/mysql` present only via an unused stub interface) | `src/` (`src/go.mod`, `src/sql/database_service.go`) |
| Legacy engine: in-memory `Data` (16 collections), encrypted atomic saves, backup/restore, save-interval goroutine, cross-process file locks, stale-job watcher | `src/data/main.go` (686 lines) |
| Public method surface: **157 methods** `func (j *JsonDatabase) …` across `src/data/*.go` (~9,332 lines incl. tests) | `src/data/` |
| Mongo-style query helpers the GORM layer must re-implement as SQL queries: `ParseFilter`, `FilterByProperty`, ordering, indexes, authorization | `src/data/filter_layer.go`, `order_layer.go`, `index_layer.go`, `authorize_layer.go` |
| Models: plain structs, **string dates**, embedded slices (e.g. `User.Roles`, `User.Claims`), JSON tags only | `src/data/models/` (~2,462 lines, 24 files) |
| DB wiring — `serviceprovider.GetDatabaseService()` returns the **concrete** `*data.JsonDatabase`; ~143 non-test call sites | `src/serviceprovider/data.go` |
| DB file path convention: root user → `/etc/parallels-devops/`, otherwise `~/.parallels-devops/`, both overridable by `DATABASE_FOLDER`; file is `<folder>/data.json` | `src/serviceprovider/main.go:69-104` |
| Config resolution `Config.GetKey`: config file → CLI flag → **env var wins**; access via `config.Get()` | `src/config/main.go:933` |
| Log level env var `LOG_LEVEL`; context logging API `LogInfof/LogWarnf/LogErrorf/LogDebugf/LogTracef`, `DisableLog()` | `src/constants/main.go:86`, `src/basecontext/api_context.go` |
| Error idiom `errors.NewWithCode(msg, httpCode)`; `ErrDatabaseNotConnected` convention | `src/data/main.go:21-24` |
| Legacy *schema* migrations (versioned, operate on the JSON store, tracked in `Data.Schema`) — **different concept** from GORM AutoMigrate; must not collide | `src/startup/migrations/` |
| Test harness: `setupTestDB` (temp dir + `memoryDatabase = nil` reset), testify, integration tests behind `//go:build integration` | `src/data/main_test.go`, `src/data/testing_reset.go` |
| Startup fatal-error precedent (panic/`os.Exit(1)` on unrecoverable init errors) | `src/main.go:59-63`, `src/startup/main.go:125` |

## 6. PBI index (delivery order — rationale in plan.md §8)

| Order | PBI | Folder | Scope in one line | Status |
|---|---|---|---|---|
| 1 | [#418](https://github.com/Parallels/prl-devops-service/issues/418) | [`418-gorm-connection/`](./418-gorm-connection/brief.md) | GORM deps + connection factory, ping, pool, logging, startup wiring | **Fully planned** (brief + plan + 5 tasks) |
| 2 | [#420](https://github.com/Parallels/prl-devops-service/issues/420) | `420-database-configuration/` | `DB_TYPE`/`DB_PATH`/`DB_DSN`/`DB_AUTO_MIGRATE` via env + config file | Brief only (deep-dive round 2) |
| 3 | [#419](https://github.com/Parallels/prl-devops-service/issues/419) | `419-gorm-model-tags/` | GORM tags, relationships, AutoMigrate schema | Brief only (round 3) |
| 4 | [#421](https://github.com/Parallels/prl-devops-service/issues/421) | `421-sql-backend-implementation/` | `GormDatabase` implementing all 157 methods on SQLite + filter/order/index layers | Brief only (round 4) |
| 5 | [#424](https://github.com/Parallels/prl-devops-service/issues/424) | `424-legacy-data-migration/` | Idempotent startup migration `data.json → SQL`, dry-run, validation | Brief only (round 5) |
| 6 | [#425](https://github.com/Parallels/prl-devops-service/issues/425) | `425-data-layer-tests/` | Suite-level test migration + coverage for backends & migration | Brief only (round 6) |
| 7 | [#426](https://github.com/Parallels/prl-devops-service/issues/426) | `426-legacy-deprecation/` | `// Deprecated:` annotations, runtime warning, lint guard, changelog | Brief only (round 7) |
| 8 | [#427](https://github.com/Parallels/prl-devops-service/issues/427) | `427-documentation-update/` | README/docs config reference, upgrade + Postgres guides, deprecation notice | Brief only (round 8) |

## 7. Feature-level success criteria

- [ ] Fresh install with zero configuration runs on SQLite at the preserved path convention.
- [ ] `DB_TYPE=postgres` + `DB_DSN` runs a fully stateless instance (no local writable state needed).
- [ ] Upgrade over an existing install: data migrated automatically on first start, idempotent,
      validated, and the service serves the same API behaviour afterwards.
- [ ] Legacy layer exists, is annotated deprecated, logs a warn when selected, gains no new callers.
- [ ] `go build ./...` and `go test ./...` green at every merged PR boundary.

## 8. Non-goals

- No MySQL/MariaDB support surface (the mysql stub in `src/sql` is unused — leave it alone).
- No API/behaviour changes visible to end users; this is an internal persistence swap.
- No removal of the legacy layer in this release — deprecation only (removal is a future issue per #426).
- No multi-tenancy / sharding work; PostgreSQL support means *one* standard relational DB.
