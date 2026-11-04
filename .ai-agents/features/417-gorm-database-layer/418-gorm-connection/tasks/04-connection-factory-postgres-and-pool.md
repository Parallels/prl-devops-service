# Task 04 — PostgreSQL branch, connection pool, DSN redaction

**PBI** [#418](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 03 (sequential — edits `connection.go`) · **PR size:** M
**Plan refs:** [D-4](./plan.md#d-4--sqlite-pragmas--postgres-pool-defaults), [D-8](./plan.md#d-8--redaction-feeds-420-ac-secrets-safety)

## Context

Postgres is the backend the whole scalability story (#417's raison d'être) rides on, and the only
backend where pooling and DSN secrets are real concerns. Both must be right at this layer: #421
will run every query through this handle, #420 will surface config errors that quote it, and the
operator-facing guide (#427) documents these very defaults.

## Objective

`NewDatabaseConnection` with `Type=postgres` returns a live pooled handle with configured limits;
`RedactDSN` scrubs secrets from any string containing a DSN; pool values are verifiable offline
(no real server needed in unit tests).

## Files

| File | Op |
|---|---|
| `src/database/postgres.go` | create |
| `src/database/redact.go` | create |
| `src/database/connection.go` | modify — replace `ErrPostgresNotWiredYet` case with `newPostgresConnection`, delete the sentinel from `errors.go` |
| `src/database/postgres_test.go` | create |
| `src/database/redact_test.go` | create |
| `src/database/postgres_integration_test.go` | create, `//go:build integration` |

## File size guard

New ≈ 90 / 70 / 120 / 90 / 60 lines; `connection.go` grows ~5 (90 → ~95). Verdict: **ok**.

## Implementation spec

### `postgres.go`

```go
func newPostgresConnection(ctx basecontext.ApiContext, cfg DatabaseConfig) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(cfg.DSN), gormConfig()) // postgres.Open does NOT dial — lazy
    if err != nil {
        return nil, errors.NewWithCode(fmt.Sprintf("failed to configure postgres connection (%s): %v", RedactDSN(cfg.DSN), err), 500)
    }
    sqlDB, err := db.DB()
    if err != nil { Close(ctx, db); return nil, errors.NewFromError(err) }
    pool := cfg.Postgres
    sqlDB.SetMaxOpenConns(pool.MaxOpen())   // defaults via getters (task 02)
    sqlDB.SetMaxIdleConns(pool.MaxIdle())
    sqlDB.SetConnMaxLifetime(pool.Lifetime())
    if d := pool.IdleTime(); d > 0 { sqlDB.SetConnMaxIdleTime(d) }
    // PingWithRetry arrives in task 05; until then a single sqlDB.Ping() here so the
    // factory still fails fast (replaced by the retry helper — leave TODO(#418 task 05)).
    if err := sqlDB.Ping(); err != nil {
        Close(ctx, db)
        return nil, fmt.Errorf("%w (%s): %v", ErrDatabaseUnreachable, RedactDSN(cfg.DSN), err)
    }
    ctx.LogInfof("PostgreSQL connection established (pool: maxOpen=%d maxIdle=%d lifetime=%s)",
        pool.MaxOpen(), pool.MaxIdle(), pool.Lifetime())   // no DSN, ever
    return db, nil
}
```

Notes: the `%w` wrap keeps `errors.Is(err, ErrDatabaseUnreachable)` true (task 05 + #421 branch on
it). Driver import is `gorm.io/driver/postgres` (canonical — the pure-Go constraint is
SQLite-only). `gormConfig()` stays the single logger seam from task 03.

### `redact.go`

```go
// RedactDSN removes credentials from a Postgres connection string for safe logging.
// Handles libpq keyword form and URL form; anything unrecognised collapses to
// "[redacted-dsn]" — fail-closed, never echo the input.
func RedactDSN(dsn string) string
```

Rules (each table-tested):
- URL form (`scheme://user:pass@host` or env-var `PG*` style): replace password segment with
  `***`; keep scheme/host/port/params. `postgres://u:p%40ss@h:5432/db?sslmode=require` →
  `postgres://u:***@h:5432/db?sslmode=require`. Usernames are kept (needed for ops triage; they
  are not secrets here — call it out in the doc-comment so reviewers see it was chosen).
- Keyword form: case-insensitive `password=` / `passwd=` value up to next space/quote → `***`.
  Quoted values with embedded spaces handled.
- Bare `sslkey=`, `sslpassword=` scrubbed too (same treatment).
- No delimiter found and string is non-trivial → `"[redacted-dsn]"`.
- Empty → empty (callers may legitimately have none).

### `postgres_integration_test.go` (`//go:build integration`)

Guarded by `POSTGRES_TEST_DSN` env (`t.Skip` when unset): open → `AutoMigrate` a scratch
struct (declared *inside* the test file — no model code leaks from #419) → insert/select roundtrip
→ bad `search_path` DSN variant yields `ErrDatabaseUnreachable`-wrapped error → pool stats sane
under 10 concurrent opens. Document in-file how to run it (`docker run --rm -e POSTGRES_PASSWORD=… -p 5432:5432 postgres:16`,
matching #425's later compose service — one canonical snippet, copied there).

### Offline tests (`postgres_test.go`, `redact_test.go`)

| Test | Assert |
|---|---|
| Factory: DSN → unreachable host `postgres://u:p@127.0.0.1:1/db` | error wraps `ErrDatabaseUnreachable`; **message contains `u:***` not `p`** |
| Factory: malformed URL DSN | error, no panic (pgx parse) |
| Pool applied offline (no Ping): build `sqlDB` from `postgres.Open` handle directly or refactor `applyPool(db, cfg)` into a pure helper — **preferred**: helper so stats assertions need no server; assert `Stats().MaxOpenConnections == 25` default & custom values |
| Redact matrix | the 8 rule cases above incl. unicode/password with `@` |
| `fmt.Errorf("%w")` chain | `errors.Is` true through the factory error |

`applyPool` refactor counts as this task's only touch of task-03 code beyond the switch — keep it
a pure function `(applyPool(db *gorm.DB, cfg DatabaseConfig) error)`.

## Out of scope

Retry loop + logger bridge (task 05), config env plumbing (#420), per-env pool overrides, PgBouncer
docs (#427), TLS option modelling beyond what the raw DSN string carries (deliberate: DSN passthrough).

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./database/... -count=1` — all offline tests green, no network attempted (verify via `go test -run TestPostgres -count=1` speed <1s)
- [ ] `cd src && go test ./database/... -count=1 -tags integration` on a machine with the docker Postgres up (paste output in PR; CI hook is #425's job)
- [ ] Grep audit: `grep -rn "cfg.DSN" src/database/*.go | grep -v _test` — every use either passed to the driver or routed through `RedactDSN` (paste in PR)
- [ ] `cd src && go test ./... -count=1` regression green; `ErrPostgresNotWiredYet` no longer exists

## Rollback

Revert commit; `connection.go` switch returns to the sentinel state only if the whole task is
reverted (task 03's markers are removed — note this in the revert PR so #421's audit grep stays honest).

## PBI acceptance mapping

Moves AC “**Connection pool**” (settings + applied + offline-proven), completes “**Connection
factory**” (both backends), and pre-arms #420's AC “**Secrets safety**”.
