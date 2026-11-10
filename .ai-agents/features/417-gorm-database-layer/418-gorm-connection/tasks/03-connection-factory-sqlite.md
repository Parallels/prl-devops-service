# Task 03 — Connection factory: SQLite branch + lifecycle

**PBI** [#418](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 02 · **PR size:** M
**Plan refs:** [D-2](./plan.md#d-2--factory-contract-ac-connection-factory-literal-newdatabaseconnection), [D-4](./plan.md#d-4--sqlite-pragmas--postgres-pool-defaults), [D-8](./plan.md#d-8--redaction-feeds-420-ac-secrets-safety) (error text only)

## Context

This is the AC's `NewDatabaseConnection` moment, SQLite half. It must create a *production-real*
handle — WAL pragmas, busy_timeout, parent-dir creation — because #421 will not revisit the
engine's boot parameters. In-memory mode (`:memory:`) is a first-class citizen from day one: the
whole test strategy (#425) and this task's own tests depend on it.

## Objective

`NewDatabaseConnection(ctx, cfg)` returns a live, pragma-verified `*gorm.DB` for `Type=sqlite`,
with `Close(ctx, db)` for teardown, unit-tested on disk and in memory. Postgres still returns an
explicit not-implemented error (this task ships the switch; task 04 fills its case).

## Files

| File | Op |
|---|---|
| `src/database/connection.go` | create |
| `src/database/sqlite.go` | create |
| `src/database/sqlite_test.go` | create |
| `src/database/connection_test.go` | create |

## File size guard

New files ≈ 90 / 110 / 200 / 60 lines. Verdict: **ok**.

## Implementation spec

### `connection.go`

```go
// NewDatabaseConnection opens the backend selected by cfg, applies engine settings,
// and returns a pool-safe handle. Caller owns Close. ctx is for logging.
// #418 scope: sqlite implemented; postgres case returns ErrUnsupportedDatabaseType
// until task 04 lands it (explicit, greppable — never silently falls back to sqlite).
func NewDatabaseConnection(ctx basecontext.ApiContext, cfg DatabaseConfig) (*gorm.DB, error) {
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    switch strings.ToLower(cfg.Type) {
    case DBTypeSQLite:
        return newSQLiteConnection(ctx, cfg)
    case DBTypePostgres:
        return nil, ErrPostgresNotWiredYet // errors.go addition; removed by task 04
    default:
        return nil, ErrUnsupportedDatabaseType
    }
}

// Close releases the underlying pool; nil-safe; logs (not errors) on failure.
func Close(ctx basecontext.ApiContext, db *gorm.DB) { ... sqlDB.Close() ... }
```

`gorm.Config` bits per D-2: logger from task 05 **not yet present** — until then use
`gorm.io/gorm/logger` in `Silent` mode constant kept in a small `gormConfig()` helper so task 05
replaces exactly one call site (leave `// TODO(#418 task 05)` marker).
`TranslateError: false`, default `NamingStrategy`.

### `sqlite.go`

```go
func newSQLiteConnection(ctx basecontext.ApiContext, cfg DatabaseConfig) (*gorm.DB, error) {
    if cfg.Path == MemoryPath { /* no dir creation, no pragmas */ } else {
        // filepath.Abs for relative paths (documented behaviour), then:
        if err := helpers.CreateDirIfNotExist(filepath.Dir(abs)); err != nil { ... }
    }
    dsn := buildSQLiteDSN(cfg.Path)
    db, err := gorm.Open(sqlite.Open(dsn), gormConfig())
    if err != nil {
        return nil, errors.NewWithCode(fmt.Sprintf("failed to open sqlite database at %s: %v", redactPath(cfg.Path), err), 500)
    }
    if cfg.Path != MemoryPath {
        if err := verifySQLitePragmas(ctx, db); err != nil { Close(ctx, db); return nil, err }
    }
    return db, nil
}
```

- `buildSQLiteDSN`: `":memory:"` for MemoryPath (glebarez expects exactly this for private
  memory DBs); else `file:<abs>?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)`
  — quote/escape the path with `url.URL{Scheme:"file", Opaque: path}` + manual pragma append so
  spaces in home dirs work (macOS reality: `/Users/Carlos Lapao/...` class paths happen).
- `verifySQLitePragmas`: `db.Raw("PRAGMA journal_mode").Scan(&mode)` expect `wal`;
  `PRAGMA foreign_keys` expect 1. Mismatch ⇒ error (e.g. WAL unsupported on some network mounts —
  fail fast at boot rather than discover at first write; message says exactly which pragma).
  On verify failure: Close first (no leaked `-wal`/`-shm` holders).
- **No pool caps for SQLite** (D-4) — comment explains why.
- Error messages may embed the **path** (not sensitive like a DSN, but still only via
  `redactPath` placeholder = plain passthrough today; task 04's redaction covers DSNs).

### Tests (both files; pattern: `basecontext.NewBaseContext()` + `DisableLog()`, `t.TempDir()`)

| Test | Assert |
|---|---|
| `TestNewDatabaseConnection_SQLite_CreatesFile` | file exists after open; `Close`; reopen → data survives (write a probe row via `db.Exec` raw) |
| `TestNewDatabaseConnection_SQLite_CreatesMissingParentDir` | path `tmp/a/b/c.db` opens; dir exists |
| `TestNewDatabaseConnection_SQLite_Pragmas` | journal_mode == wal; foreign_keys == 1 |
| `TestNewDatabaseConnection_SQLite_PathWithSpaces` | `tmp/My Docs/data.db` opens |
| `TestNewDatabaseConnection_Memory` | no file on disk; insert+select works; after `Close`, reopen is empty (private memory semantics) |
| `TestNewDatabaseConnection_RejectsUnwritablePath` | path under an existing *file* (e.g. `tmp/probe.db/sub/x.db`) → error, no panic, no leaked handle (subsequent identical call must fail the same way) |
| `TestNewDatabaseConnection_PostgresReturnsNotWired` | `ErrPostgresNotWiredYet` |
| `TestNewDatabaseConnection_InvalidType` | from `Validate()` (e.g. Type `"mongodb"`) |
| `TestClose_NilSafe` | no panic on nil |

Concurrency guard test (cheap, catches real WAL pain): 20 goroutines inserting concurrently into
one file DB with busy_timeout → all succeed (this is the AC-adjacent "concurrent access" seed for
#421).

## Out of scope

Postgres branch body, ping-with-retry + logger bridge (task 05), AutoMigrate, GormDatabase, any
serviceprovider touch.

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./database/... -count=1 -race` (race on the concurrency test matters)
- [ ] `cd src && go test ./... -count=1` — legacy suite untouched
- [ ] Manual: run one test with `-v`, confirm no GORM banner noise on stdout (Silent config)
- [ ] `grep -n "ErrPostgresNotWiredYet\|TODO(#418 task 05)" src/database/connection.go` — both markers present for task 04/05 to consume

## Rollback

Revert commit removes 4 files; `doc.go` keeps deps honest. No consumers yet.

## PBI acceptance mapping

Moves AC “**Connection factory**” (SQLite half + shape) and “**No regressions**”; pre-satisfies
#421's “File path respected” + “Concurrent access (WAL)” engine requirements.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `418-03`.** Cut `slice/418-03` from `sql-migration` HEAD
(`.worktrees/slice-418-03`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
