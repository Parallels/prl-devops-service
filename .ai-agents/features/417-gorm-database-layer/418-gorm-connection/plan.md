# PBI Plan #418 — GORM dependency setup and database connection management

> Reading order: [feature brief](../brief.md) → [feature plan](../plan.md) (AD-2, AD-3, ER-1…ER-5) →
> [PBI brief](./brief.md) → this plan → your task spec.

## Scope

**In:** GORM dependencies · `DatabaseConfig` contract + defaults · connection factory (SQLite +
PostgreSQL) · ping health check with retry · PostgreSQL pool config · DSN redaction helper · GORM
logger bridged to `basecontext`/`LOG_LEVEL` · additive service-provider wiring of the shared
connection · unit tests.
**Out:** AutoMigrate execution (no models until #419; flag is stored, not acted on), `DB_*` env
parsing (#420 fills task 02's seam), any legacy-layer change, any controller change.

## Design decisions

### D-1 — New package `src/database`, one concern per file

All new code lands in `src/database/` (package `database`). Planned file sizes ≤ ~200 lines —
comfortably under ER-2 caps:

```
src/database/
├── doc.go            # package doc + blank imports pinning drivers (task 01)
├── errors.go         # sentinels: ErrUnsupportedDatabaseType, ErrDatabaseUnreachable, … (task 02)
├── config.go         # DatabaseConfig, constants, defaults, Validate (task 02)
├── connection.go     # NewDatabaseConnection, Close (task 03, extended by 04)
├── sqlite.go         # SQLite branch: pragmas, dir creation, :memory: (task 03)
├── postgres.go       # Postgres branch + pool application (task 04)
├── redact.go         # RedactDSN (task 04)
├── health.go         # PingWithRetry (task 05)
├── logger.go         # gorm logger.Interface adapter (task 05)
└── *_test.go         # colocated; tests use t.TempDir()/":memory:" only
```

Later PBIs add more files to the same package (`gorm_database.go` etc.) — keep it flat.

### D-2 — Factory contract (AC "Connection factory", literal `NewDatabaseConnection`)

```go
// NewDatabaseConnection opens the configured backend, applies pragmas/pool settings,
// and verifies it with PingWithRetry. The caller owns Close. Safe for concurrent use
// (GORM/sql.DB are pool-safe). ctx is used for logging only.
func NewDatabaseConnection(ctx basecontext.ApiContext, cfg DatabaseConfig) (*gorm.DB, error)
```

- Opens → applies settings → pings (D-5) → returns the live handle. On any failure after open, the
  half-open handle is closed before returning the error (no leaked file locks / sockets).
- `gorm.Config` shared bits: `Logger: NewGormLogger(ctx)` (D-6); `NamingStrategy`: GORM default
  (OQ-3 stance); `TranslateError: false` — mapping to the repo's `Err*` sentinels happens in #421
  where domain errors live; here we surface `gorm`/driver errors wrapped with our sentinels.

### D-3 — `DatabaseConfig` is a neutral struct, not `*config.Config`

```go
type DatabaseConfig struct {
    Type      string        // DBTypeSQLite (default) | DBTypePostgres
    Path      string        // SQLite file; ":memory:" allowed; default resolved below
    DSN       string        // Postgres only; required when Type=postgres
    AutoMigrate bool        // stored for #421/#424; NOT acted on in #418
    Postgres  PostgresPoolConfig
}
type PostgresPoolConfig struct {
    MaxOpenConns    int           // default 25
    MaxIdleConns    int           // default 5
    ConnMaxLifetime time.Duration // default 30m
    ConnMaxIdleTime time.Duration // default 5m (0 = driver default)
}
```

- `Validate()` — pure function, table-tested: unknown type; postgres w/o DSN; sqlite w/o Path
  (empty after default resolution is invalid). Returns `errors.NewWithCode`-style sentinels from
  `errors.go` so #420 can build its user-facing messages on them.
- `DefaultDatabaseConfig()` implements AD-3 folder resolution *without* importing `serviceprovider`
  (avoid cycles/globals): honour explicit override first; then `DATABASE_FOLDER` via
  `config.Get().DatabaseFolder()`; else uid-0 → `/etc/parallels-devops`, otherwise
  `os.UserHomeDir() + "/.parallels-devops"`. File name `prl-devops.db`.
  **Seam for #420:** `FromApplicationConfig() (DatabaseConfig, error)` is *declared here* (returns
  defaults today, documented `TODO(#420)`) so #420 changes one file, not the factory.
- The uid/home heuristic intentionally duplicates the *convention* (documented in AD-3); the
  authoritative `System.GetUserHome` is unreachable at this layer without the provider — noted, not
  a bug. Windows: `os.UserHomeDir()` covers it; the service already normalises folders elsewhere.

### D-4 — SQLite pragmas & Postgres pool defaults

- SQLite DSN (glebarez syntax): `<path>?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)`
  — WAL satisfies #421's concurrency AC from day one; `busy_timeout` keeps concurrent writers from
  `SQLITE_BUSY` flakes. Parent dir created with `helpers.CreateDirIfNotExist`. `":memory:"` skips
  pragmas that don't apply (journal stays default) and never creates files.
- Pool defaults (D-3 struct) applied **only** for postgres via `sqlDB.Set*` (SQLite: single-file,
  glebarez handles serialisation; setting caps there is a footgun — skip, documented).
  `MaxOpenConns` verifiable offline via `db.SQLDB().Stats().MaxOpenConnections`.

### D-5 — Health ping with retry

```go
type Pinger interface{ Ping() error } // *sql.DB satisfies it
func PingWithRetry(ctx basecontext.ApiContext, p Pinger, attempts int, delay time.Duration) error
```
Defaults from the factory: 3 attempts, 500 ms delay. Failure wraps `ErrDatabaseUnreachable` with
the **redacted** DSN/path in the message. The interface keeps it unit-testable with a fake (no
network). "Clear error and exits gracefully" lands in the wiring (D-7), not here.

### D-6 — GORM logger bridge

`logger.go` implements `gorm logger.Interface` (all 8 methods incl. `Trace` for slow queries):

| GORM level | BaseContext method | Gate |
|---|---|---|
| Silent / test | — | `logger.Default.LogMode(logger.Silent)` when `ctx` logging disabled |
| Info | `ctx.LogInfof` | always |
| Warn | `ctx.LogWarnf` | always |
| Error | `ctx.LogErrorf` | always |
| Trace (params/rows) | `ctx.LogDebugf` | only when effective level is `debug`/`trace` |

Effective level = `config.Get().GetKey(constants.LOG_LEVEL_ENV_VAR)` (values seen: `trace`,
`debug`, `info`, `warn`, `error`). `Trace()` additionally logs any statement over 200 ms at warn
(slow-query observability, cheap and bounded). GORM's default noisy `info` mode is thus demoted to
our debug — boot logs stay clean (regression AC).

### D-7 — Service-provider wiring & failure policy (AC "exits gracefully")

- New file `src/serviceprovider/database_connection.go`:
  ```go
  // initGormConnection builds DatabaseConfig via DefaultDatabaseConfig(), opens the
  // connection (D-2) and stores it on the provider. Called after the legacy
  // JsonDatabase is connected in each serviceprovider init branch.
  func (s *ServiceProvider) initGormConnection(ctx basecontext.ApiContext) error
  func GetSQLDatabase(ctx basecontext.ApiContext) (*gorm.DB, error) // mirrors GetDatabaseService
  ```
  Provider gains field `SQLDatabase *gorm.DB` (`src/serviceprovider/main.go` struct; +≤2 lines per
  init branch × 4 branches — each adds `if err := globalProvider.initGormConnection(ctx); err != nil { ... }`).
- **Failure policy (deliberate, reviewable):**
  - `config.Get().GetRunningCommand() == constants.API_COMMAND` → `ctx.LogErrorf` (message
    includes redacted DSN) + `os.Exit(1)` — matches the AC and the `src/main.go` precedent.
  - Any other command (CLI tools, migrations run standalone) → warn and continue: CLI must remain
    usable to *fix* a bad DB config (e.g. run a dry-run migration), and unit tests construct
    providers freely (exit(1) there would poison the suite).
- Nothing consumes `SQLDatabase` yet; #421 flips selection. Risk of double-handle on the SQLite dir
  is nil (different file, `data.json` untouched).

### D-8 — Redaction (feeds #420 AC "Secrets safety")

`RedactDSN(dsn string) string` handles both forms: URL style `postgres://u:pw@host/db` →
`postgres://u:***@host/db` (keep host/port/params), and key/value style
`host=… password=… user=…` → `password=***` (case-insensitive key match). Never panics; unknown
forms return `"[redacted-dsn]"` rather than leak. Used by every error/log path in this package that
could include a DSN.

## Task index (each task = one PR, merge in order)

| Task | Goal | Depends on | Touches | Size |
|---|---|---|---|---|
| [01](./tasks/01-add-gorm-dependencies.md) | GORM + drivers in `go.mod`, pinned by `doc.go` | — | `src/go.mod`, `src/go.sum`, `src/database/doc.go` | S |
| [02](./tasks/02-database-config-contract.md) | `DatabaseConfig` + defaults + validation | 01 | `errors.go`, `config.go`, tests | S |
| [03](./tasks/03-connection-factory-sqlite.md) | `NewDatabaseConnection` SQLite branch + Close | 02 | `connection.go`, `sqlite.go`, tests | M |
| [04](./tasks/04-connection-factory-postgres-and-pool.md) | Postgres branch + pool + `RedactDSN` | 03 | `postgres.go`, `redact.go`, `connection.go` (switch case), tests | M |
| [05](./tasks/05-startup-wiring-healthcheck-logging.md) | ping retry, logger bridge, provider wiring + failure policy | 04 | `health.go`, `logger.go`, `serviceprovider/database_connection.go`, `serviceprovider/main.go` (+~12 lines), tests | M |

Parallelism note: 03 and 04 both edit `connection.go` — keep them **sequential** (04's change is
one switch-case + imports). No other file overlap.

## Verification commands (all tasks)

```bash
cd src
go build ./...                       # must stay green at every task
go vet ./database/... ./serviceprovider/...
go test ./database/... -count=1 -v   # new package's tests
go test ./... -count=1               # no regressions (legacy suite untouched)
```

## Risks specific to #418

| Risk | Mitigation |
|---|---|
| `go mod tidy` churn unrelated deps in task 01 | Review `go.mod` diff; `doc.go` blank imports keep exactly the 3 additions + transitive graph; revert unrelated bumps |
| glebarez/sqlite version drift vs gorm core | Pin exact versions in 01; #419's AutoMigrate smoke is the real integration proof (early-warning list handed over) |
| `os.Exit(1)` in provider init breaking tests | Policy D-7 gates on `API_COMMAND`; unit tests never set it; integration test starts the binary against a dead Postgres DSN and asserts exit code + log line |
| WAL mode changes file layout (`-wal`/`-shm` sidecar files) | Documented; backup logic (legacy `backup.go`) only ever covered `data.json` — new-DB backups are #421+ concerns, noted for its round |

## Handover to next round (#420)

- `FromApplicationConfig()` seam exists in `config.go` (single fill-in point).
- `RedactDSN` ready to reuse.
- Pool/pragmas are config-plumbed (`PostgresPoolConfig`); #420 may surface env overrides
  (`DB_PG_*`) — decide then, not here.
