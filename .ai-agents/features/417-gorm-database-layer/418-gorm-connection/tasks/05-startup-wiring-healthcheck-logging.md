# Task 05 — Startup wiring, ping health check, GORM logger bridge

**PBI** [#418](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 04 · **PR size:** M
**Plan refs:** [D-5](./plan.md#d-5--health-ping-with-retry), [D-6](./plan.md#d-6--gorm-logger-bridge), [D-7](./plan.md#d-7--service-provider-wiring--failure-policy-ac-exits-gracefully)

## Context

The last of the four ACs that need runtime presence: the service must *own* a verified GORM
connection from boot ("the shared `DB` instance used by all subsequent data access work") with
clear failure semantics, and GORM's own chatter must speak our logging dialect. This is the only
task in the PBI that can affect boot behaviour — the failure policy in D-7 exists so that risk is
one explicit, reviewable decision instead of an accident.

## Objective

On startup (API command), the provider opens the SQLite connection via the factory, verifies it
with `PingWithRetry`, stores it (`SQLDatabase`), logs via the bridged logger, and on failure logs a
clear redacted error and exits 1. Everything else (CLI paths, unit tests of other packages) is
unaffected.

## Files

| File | Op |
|---|---|
| `src/database/health.go` | create |
| `src/database/health_test.go` | create |
| `src/database/logger.go` | create |
| `src/database/logger_test.go` | create |
| `src/database/connection.go` | modify — replace `sqlDB.Ping()`/`gormConfig()` internals (task 03/04 TODOs) |
| `src/serviceprovider/database_connection.go` | create (~70 lines) |
| `src/serviceprovider/main.go` | modify — struct field + 4 call sites, ≤ ~12 lines total (530 → ~542) |
| `src/serviceprovider/database_connection_test.go` | create |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `src/serviceprovider/main.go` | 530 | ~542 | ok |
| `src/database/connection.go` | ~95 | ~100 | ok |
| all others | new | <200 each | ok |

## Implementation spec

### `health.go` — D-5 verbatim

```go
type Pinger interface{ Ping() error }
const (
    defaultPingAttempts = 3
    defaultPingDelay    = 500 * time.Millisecond // flat; slow-start DBs are a #420/#421 tuning topic
)
func PingWithRetry(ctx basecontext.ApiContext, p Pinger, attempts int, delay time.Duration) error {
    // loop; attempt>1 logs ctx.LogWarnf("database ping attempt %d/%d failed: %v", ...);
    // final failure: fmt.Errorf("%w: %d attempts", ErrDatabaseUnreachable, attempts)
}
```

Factory integration: SQLite & Postgres branches replace their plain `Ping()` with
`PingWithRetry(ctx, sqlDB, …)` (removes task 04's `TODO(#418 task 05)` marker; the error-wrap
with redacted DSN stays).

### `logger.go` — `gorm logger.Interface` adapter (D-6)

```go
type baseContextLogger struct { ctx basecontext.ApiContext; level string; slowThreshold time.Duration }
func NewGormLogger(ctx basecontext.ApiContext) gormlogger.Interface // reads config.Get().GetKey(constants.LOG_LEVEL_ENV_VAR)
```

- Implement `LogMode(logger.LogLevel)` storing an internal mode (only `Silent` suppresses output —
  mode mapping: `Silent`→nothing, `Error`→errors only, `Warn`→warn+errors, `Info`→all of D-6 table).
- `Trace(ctx context.Context, begin time.Time, fc func() (string,int64), err error)`: elapsed >
  200ms → `LogWarnf("slow query (%s): %s", elapsed, sql)`; else when effective level is debug/trace
  → `LogDebugf(sql, rows, elapsed)`; `err != nil` always → `LogErrorf`. Note the ctx here is
  **GORM's** `context.Context`, not our ApiContext — never route through it; use the struct's.
- `LogMode(Silent)` must be returned when the caller passes a `DisableLog()`'d context — expose
  constructor variant `NewGormLoggerSilent(ctx)` for tests to opt out deterministically.
- `connection.go`'s `gormConfig()` now returns `&gorm.Config{Logger: NewGormLogger(ctx)…}` —
  signature grows a ctx param; update task-03/04 call sites (their tests already use a disabled
  context → expect silent).

### `src/serviceprovider/database_connection.go` — D-7 verbatim

```go
func (s *ServiceProvider) initGormConnection(ctx basecontext.ApiContext) error {
    cfg := database.DefaultDatabaseConfig() // seam: #420 swaps to FromApplicationConfig()
    db, err := database.NewDatabaseConnection(ctx, cfg)
    if err != nil { s.SQLDatabase = nil; return err }
    s.SQLDatabase = db
    ctx.LogInfof("SQL database connection ready (type: %s)", cfg.Type)
    return nil
}
func GetSQLDatabase(ctx basecontext.ApiContext) (*gorm.DB, error) // provider-nil / SQLDatabase-nil → database.ErrDatabaseUnreachable
```

Failure policy lives at the **call sites**, not inside the method (testability):

```go
// helper in the same file
func (s *ServiceProvider) mustInitGormConnection(ctx basecontext.ApiContext) {
    if err := s.initGormConnection(ctx); err != nil {
        ctx.LogErrorf("SQL database unavailable: %v", err)          // message carries redacted DSN (factory did it)
        if config.Get().GetRunningCommand() == constants.API_COMMAND { os.Exit(1) } // D-7
        ctx.LogWarnf("continuing without SQL connection (non-API command)")
    }
}
```

Call sites: after each of the 4 `_ = globalProvider.JsonDatabase.Connect(ctx)` lines in
`src/serviceprovider/main.go` (`grep -n "JsonDatabase.Connect" src/serviceprovider/main.go` →
≈ lines 82, 101, 217, 238 — re-grep, don't trust line numbers): one added line
`globalProvider.mustInitGormConnection(ctx)`. Struct: add `SQLDatabase *gorm.DB` next to
`JsonDatabase` (line ~41). Close: wire `database.Close` into the same shutdown path that today
closes `memoryDatabase.cancel` if one exists in `src/main.go`/`src/startup` — if none exists,
document with `// TODO(#421): graceful shutdown hook` (the OS reaps sqlite/postgres handles;
#421 owns proper shutdown).

### Tests

| Test | Assert |
|---|---|
| `PingWithRetry` with fake Pinger (fails k times) | k<attempts → nil + k warn logs captured? (use a recording fake ctx implementing `basecontext.ApiContext` — interface exists, build a 15-method fake, this fake also serves logger tests) |
| fake always fails, attempts=3 | `ErrDatabaseUnreachable`, exactly 3 attempts observed, total elapsed ≥ 2×delay |
| Logger level matrix | debug-level ctx logs Trace statements; info-level doesn't; slow-query (fake `begin` 300ms old) → warn; `LogMode(Silent)` → zero output |
| `initGormConnection` happy path | `t.Setenv("DATABASE_FOLDER", t.TempDir())` + `config.New(ctx)` (reset global, pattern `src/data/main_test.go`) → field non-nil, file exists |
| `initGormConnection` failure | point `DATABASE_FOLDER` under an existing file → error returned, field nil, **no exit** (method-level; policy is call-site tested separately) |
| `GetSQLDatabase` | provider absent → `ErrDatabaseUnreachable` |
| Regression boot check (manual, PR step) | `cd src && go run . api --help`-style boot of the API command against temp folder, observe "SQL database connection ready" log + WAL files, Ctrl-C exits — record output in PR. **Plus the exit-1 case:** `DB_TYPE` can't be forced to postgres pre-#420; simulate by making the sqlite path unwritable (`chmod 500` dir) and assert exit code 1 + the error line. |

## Out of scope

Replacing legacy selection (nothing reads `SQLDatabase` yet — that's #421's switch), AutoMigrate,
`DB_*` env parsing (#420 fills `FromApplicationConfig` + the one call-site swap to it), Helm values
(#427/OQ-4), proper shutdown hook (noted above).

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./database/... ./serviceprovider/... -count=1` (new tests; `-race` for health)
- [ ] `cd src && go test ./... -count=1` — **the PBI's “No regressions” AC proof**: full suite green
- [ ] Manual boot matrix (both cases above) pasted in PR
- [ ] `grep -rn "os.Exit" src/serviceprovider/` — exactly one, gated on `constants.API_COMMAND`, visible in PR diff
- [ ] `grep -c "mustInitGormConnection(ctx)" src/serviceprovider/main.go` ≥ 4 (one per legacy Connect branch)

## Rollback

Revert commit: provider field/call sites + three new files disappear; factory/tasks 01–04 remain
unaffected (they have no runtime presence). Zero user impact since nothing consumes the handle.

## PBI acceptance mapping

Moves AC “**Health check**” (ping on startup + clear error + graceful exit per repo precedent),
“**Logging**” (bridge + silent-in-tests), finishes “**No regressions**”. After this PR merges, all
six #418 ACs are demonstrably addressed → close #418 with the AD-2 deviation comment linked.
