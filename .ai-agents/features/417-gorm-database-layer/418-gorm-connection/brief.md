# PBI Brief #418 — GORM dependency setup and database connection management

| Field | Value |
|---|---|
| PBI | [#418](https://github.com/Parallels/prl-devops-service/issues/418) — `[PBI] GORM dependency setup and database connection management` |
| Labels | `pbi` · Assignee: sai-kumar-peddireddy |
| Round | **Round 1 — first PBI to implement; fully planned** |
| Upstream feature docs | [`brief.md`](../brief.md) · [`plan.md`](../plan.md) — required reading (esp. AD-2, AD-3) |
| Planning docs | [`plan.md`](./plan.md) · tasks [01](./tasks/01-add-gorm-dependencies.md) · [02](./tasks/02-database-config-contract.md) · [03](./tasks/03-connection-factory-sqlite.md) · [04](./tasks/04-connection-factory-postgres-and-pool.md) · [05](./tasks/05-startup-wiring-healthcheck-logging.md) |

## Description (verbatim from GitHub)

Introduce GORM as the project's ORM framework and establish the foundational database connection
infrastructure. This PBI delivers the shared `DB` instance used by all subsequent data access work,
and sets up the connection lifecycle (open, ping, close) for both supported backends.

### User Story

As a **backend engineer**, I want a centralised, well-configured GORM database connection that
supports both SQLite and PostgreSQL, so that all data layer components can build on a consistent and
reliable foundation.

### Acceptance Criteria

- [ ] **GORM added**: `gorm.io/gorm`, `gorm.io/driver/sqlite`, and `gorm.io/driver/postgres` are added to `go.mod` and `go.sum`. ⚠️ *Approved deviation:* SQLite driver is `github.com/glebarez/sqlite` (pure-Go `modernc.org/sqlite` engine) instead of CGO `gorm.io/driver/sqlite` — feature AD-2; see OQ-5, post the deviation comment on #418 during task 01's PR.
- [ ] **Connection factory**: A `NewDatabaseConnection(cfg Config) (*gorm.DB, error)` function (or equivalent) initialises the correct driver based on the configured backend.
- [ ] **Health check**: The connection is verified with a ping on startup; the service logs a clear error and exits gracefully if the database is unreachable.
- [ ] **Connection pool**: Pool settings (max open connections, max idle connections, connection lifetime) are configurable and applied on startup for PostgreSQL.
- [ ] **Logging**: GORM's logger is configured to use the project's existing log level settings (silent in tests, info/warn in production).
- [ ] **No regressions**: The existing application boots and passes all current tests with the new dependency present.

### Definition of Done

- [ ] Code implemented following best practices.
- [ ] Unit tests written and passing.
- [ ] Code reviewed and approved.
- [ ] Merged into the main branch.
- [ ] Documentation updated (if applicable).
- [ ] Deployed to staging/production environment.

### Assumptions and Constraints

None stated on GitHub. Inherited from the feature: ER-1…ER-5, AD-2 (driver), AD-3 (path).

### Dependencies

None on GitHub. In-feature position: **first**; downstream consumers #419 (AutoMigrate needs the
factory), #420 (will *fill in* the config contract task 02 leaves as a seam), #421 (pool/logger/
health are the foundation of `GormDatabase`), #425 (harness builds on `:memory:` support here).

## Repo context (verified 2026-11-03 — anchors for the five tasks)

| Anchor | Value for this PBI |
|---|---|
| `src/go.mod` | Module `github.com/Parallels/prl-devops-service`; Go 1.25.0; GORM absent; `go.work` at repo root points at `src` — run Go commands in `src/`. |
| `src/sql/` | 58-line **unused** stub (`DatabaseService` interface + mysql blank import) — do not extend, do not import; the new package is `src/database/` (AD layout). |
| DB folder convention | root → `/etc/parallels-devops/`; else `~/.parallels-devops/`; `DATABASE_FOLDER` overrides — `src/serviceprovider/main.go:69-104`. AD-3: new SQLite default = `<folder>/prl-devops.db`. |
| Idioms to copy | ctx-first `basecontext.ApiContext` (methods `LogInfof/LogWarnf/LogErrorf/LogDebugf/LogTracef`, `DisableLog()`); errors via `errors.NewWithCode` (`src/errors`); helpers like `helpers.CreateDirIfNotExist` (used at `src/serviceprovider/main.go:71`). |
| Log level | `LOG_LEVEL` env → `config.Get().GetKey(constants.LOG_LEVEL_ENV_VAR)` (`src/constants/main.go:86`, `src/config/main.go:229`). |
| Startup failure precedent | `src/main.go:59-63` snapshot+`os.Exit(1)`; `src/startup/main.go:125` `panic(err)` — "exits gracefully" in this repo means *clear log then non-zero exit*, not silent survival. |
| Provider init sites | `src/serviceprovider/main.go` — 4 branches create the legacy DB and call `Connect` (≈ lines 75-104, 210-239): the GORM init call slots once per branch via one helper (task 05). |
| Tests | testify available; `src/data/main_test.go` pattern; #418's tests live in `src/database/*_test.go`, must not touch disk except via `t.TempDir()`. |

## Engineering-rule triggers for this PBI

- ER-2/ER-3: **no file touched is near the cap** — `src/serviceprovider/main.go` is 530 lines
  (task 05 adds ≤ ~12 lines total), everything else is new and small. Guard tables completed in
  each task spec.
- ER-5: after **each** task PR, `cd src && go build ./... && go test ./... ` must be green — task
  05 is the only one that can affect boot behaviour, and its failure policy is explicit.
- OQ-5: deviation comment on #418 (AD-2) posted with task 01's PR — the human closing the PBI must
  see why the `gorm.io/driver/sqlite` box can be ticked as "deviation approved".

## Out of scope for this PBI (do not sneak in)

- AutoMigrate execution (no models exist yet — flag stored, execution lands with #421/#424; AD-4).
- `DB_TYPE`/`DB_PATH`/`DB_DSN` env reading and validation UX (**#420's** job; task 02 only builds
  the neutral contract + legacy-default resolution that #420 will feed).
- Anything touching `src/data/**` (legacy engine is untouched until #421/#426), `src/sql/**`,
  controllers, or `serviceprovider`'s legacy selection (task 05 only *adds* a parallel GORM
  connection; it does not switch anything).
