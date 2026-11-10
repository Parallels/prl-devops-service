# Task 04 — GormDatabase bootstrap: hybrid type, all-domain delegate stubs, error mapper

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 03 · **PR size:** M
**Plan refs:** [D-1](../plan.md#d-1--hybrid-delegate-gormdatabase-adopts-the-recommended-architecture-refined), [D-7](../plan.md#d-7--lifecycle-semantics-in-sql-mode-escape-hatch-table), [D-11](../plan.md#d-11--error-mapping-ac-error-handling)

## Context

This PR makes `GormDatabase` a *complete* `DatabaseService` on day one — every method one line of
`return g.legacy.X(...)`. Interface satisfaction (R-1 assertion for the SQL side) is proven before any domain
flips, the parity harness (task 03) gets its SQL engine target, and the error mapper that every domain task
reuses lands here. From here to task 16, domain tasks **replace** stubs (delete lines), never reshape the type.

## Objective

`database.GormDatabase` implements `interfaces.DatabaseService` (compile assertion green), delegates all 157
methods to an injected legacy handle, exposes `NewGormDatabase`/`DB()`/`Close`, and ships
`mapGormError` + lifecycle semantics per D-7. Unit-tested against `:memory:` + a seeded legacy delegate.

## Files

| File | Op |
|---|---|
| `src/database/gorm_database.go` | create (~110 lines) — struct, constructor, `DB()`, `Close`, `IsConnected/Connect/Disconnect` |
| `src/database/errors.go` | create (~60) — `mapGormError`, constraint helpers |
| `src/database/delegate_users.go` … `delegate_lifecycle.go` | create — 16 files, one per task-01 interface file, stub bodies |
| `src/database/gorm_database_test.go` | create |
| `src/database/errors_test.go` | create |

## File size guard

| File | Lines | Verdict |
|---|---|---|
| `gorm_database.go` new | ~110 | ok |
| `delegate_orchestrator.go` (largest stub file, 31 one-liners) | ~120 | ok |
| all other delegate files | ≤ 100 | ok |
| (`src/serviceprovider/main.go` untouched here) | 530 | ok |

## Implementation spec

```go
// gorm_database.go
type GormDatabase struct {
    gormDB   *gorm.DB
    legacy   *data.JsonDatabase      // hybrid delegate; deleted in task 18
    onJobTimeout func(models.Job)
    timeoutMu    sync.RWMutex
    connected    atomic.Bool
    closeOnce    sync.Once
}

func NewGormDatabase(ctx basecontext.ApiContext, gormDB *gorm.DB, legacy *data.JsonDatabase) *GormDatabase {
    g := &GormDatabase{gormDB: gormDB, legacy: legacy}
    g.connected.Store(gormDB != nil)
    return g
}
func (g *GormDatabase) DB() *gorm.DB { return g.gormDB }            // WAL audit + #424 write path
func (g *GormDatabase) Connect(ctx basecontext.ApiContext) error    // ping gormDB (cheap), set connected
func (g *GormDatabase) Disconnect(ctx basecontext.ApiContext) error // stopWatch if watcher wired (t15/t18), closeOnce Close
func (g *GormDatabase) IsConnected() bool                           // connected.Load()
```

**Every domain method in this PR is exactly** (example file `delegate_users.go`):

```go
func (g *GormDatabase) GetUsers(ctx basecontext.ApiContext, filter string) ([]models.User, error) {
    return g.legacy.GetUsers(ctx, filter)
}
// … all 11 of data/users.go per the task-06 inventory
```

Method lists per delegate file = the inventories embedded in domain tasks 05–16 (copy signatures verbatim from
those tables; task 01 already froze them in the interfaces).

**Lifecycle overrides — these are NOT legacy delegates, implement per D-7 now** (`delegate_lifecycle.go`):
`SaveNow`→`nil`; `SaveAsync`→`nil`; `ProcessSaveQueue`→ no-op; `Load`→`nil`; `IsDataFileEmpty`→`false,nil`;
`Filename`→ `g.dbFilename()` (from `gormDB` DSN / config; `":memory:"` verbatim); `Backup`/`RunBackup`→ log-once
warn + nil; `SaveAs`→ **task 04 delegates to legacy** with `// TODO(#421 task 18): VACUUM INTO` marker (SQL
snapshot needs the flip-day wiring/verification); `GetSchemaVersion`/`UpdateSchemaVersion`→ task 08 replaces.

`errors.go`:

```go
// mapGormError maps GORM/driver failures onto legacy sentinel identity (D-3).
// notFound/dup may be nil (legacy has no sentinel for that domain op → generic 500/400 wrap).
func mapGormError(err error, notFound, dup *errors.SystemError) error {
    if err == nil { return nil }
    if errors.Is(err, gorm.ErrRecordNotFound) && notFound != nil { return notFound }
    if isUniqueViolation(err) && dup != nil { return dup }            // ErrDuplicatedKey || "UNIQUE constraint failed" || pg 23505 (string match now, pg code in task 18)
    return errors.NewWithCode(fmt.Sprintf("database operation failed: %v", err), 500)
}
```

**Compile assertion** (`gorm_database_test.go`): `var _ interfaces.DatabaseService = (*GormDatabase)(nil)`.

Tests:

| Test | Assert |
|---|---|
| `TestNewGormDatabase_ImplementsInterface` | assertion + `Connect` on `:memory:` handle nil-err; `IsConnected` true |
| `TestDelegates_ToLegacy` | table: one representative per domain — seed legacy engine (task-03 harness), assert identical output through `GormDatabase` |
| `TestLifecycleSemantics` | `SaveNow`/`Load`/`SaveAsync` return nil; `IsDataFileEmpty`=false; `Filename`=`:memory:`; `Backup` logs warn once (recording ctx) and returns nil |
| `TestMapGormError` | `ErrRecordNotFound`→sentinel identity (`==`), unique-violation→dup sentinel, unknown→`*errors.SystemError` code 500, nil→nil |
| `TestClose_Idempotent` | double `Disconnect` no panic (closeOnce) |

## Out of scope

Any real SQL implementation; watcher (15); provider wiring (18); `VACUUM INTO` (18); touching `src/data` code.

## Validation

- [ ] `cd src && go build ./... && go vet ./database/...`
- [ ] `cd src && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./... -count=1`
- [ ] `grep -c "return g.legacy\." src/database/delegate_*.go | …` sum = 157 − (12 lifecycle/schema overrides) — count reconciled in PR description against the interface file totals (plan D-2 table)
- [ ] `grep -rn "func (g \*GormDatabase)" src/database/*.go | wc -l` → 157 + helpers

## Rollback

Delete `gorm_database.go`, `errors.go`, `delegate_*.go` + tests. No importers outside `database`/`parity`.

## PBI acceptance mapping

Structural half of AC-1 ("every entity … GORM" target type exists and type-checks); AC-5 groundwork (mapper);
pre-satisfies harness's SQL engine so AC-1 parity tests can start flowing from task 05.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-04`.** Cut `slice/421-04` from `sql-migration` HEAD
(`.worktrees/slice-421-04`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
