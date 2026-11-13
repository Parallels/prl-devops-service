# Task 01 — `testutil` harness: unique-DSN in-memory scheme, isolation rules, recording context, engine factories

**PBI** [#425](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** #418 + #419 + #420 + #421 merged (factories, schema, GormDatabase) · **PR size:** M
**Plan refs:** [D-2](./plan.md#d-2--srdatabasetestutil-as-an-exported-packageunique-dsn-scheme418s-sqlitego-uri-guard), [D-3](./plan.md#d-3--global-state-hazards--isolation-rulesdocumented-as-package-doc-comment-not-a-testingmd), [D-9](./plan.md#d-9--baseline-numbers-measured-2026-11-03-warm-build-cache)

## Context

Everything in this PBI stands on one harness. The AC names `file::memory:?cache=shared`, which is
*one logical DB per process* — unguarded, every parity test silently shares state with every other,
and the suite becomes the flakiness machine AC 6 forbids. This task builds the guarded version
(unique named memory DSNs, harness-owned cleanup), the recording-`ApiContext` fake reused from #418
task 05's pattern, and the two engine factories that let one scenario run on both engines. It also
front-loads the budget risk: measuring per-test open+AutoMigrate cost *now*, before ~100 scenarios
exist, so a blown +30 s budget is discovered for the price of one PR.

## Objective

`go test ./database/... -count=1` is green and shows one parity scenario running as
`TestParityUsersSmoke/legacy` and `TestParityUsersSmoke/gorm` against isolated in-memory DBs;
`testutil`'s own leak tests prove same-name DSNs share and distinct DSNs don't; the AutoMigrate cost
number is pasted in the PR description.

## Files

| File | Op |
|---|---|
| `src/database/testutil/doc.go` | create — package doc = isolation rules (D-3 list verbatim), DSN scheme, usage snippets |
| `src/database/testutil/harness.go` | create — `OpenMemoryDB`, `MigrateSchema`, `WithCleanConfig`, `memoryDSN()` sanitizer + `atomic.Uint64` seq |
| `src/database/testutil/context.go` | create — `RecordingContext()` + `RecordingApiContext` (14-method fake, `src/basecontext/api_context.go:10-25`) + matchers |
| `src/database/testutil/engines.go` | create — `Engine`/`Scenario`/`EngineCapabilities`/`RunOnEngines`, `LegacyEngine()`, `GormEngine()` |
| `src/database/testutil/harness_test.go` | create — DSN leak/isolation tests, sanitizer table test, recording-ctx test, **timing test** |
| `src/database/parity/doc.go` | create — suite doc + "how to add a scenario" + *draft* exclusion ledger (task 04 finalizes) |
| `src/database/parity/parity_test.go` | create — bootstrap + `TestParityUsersSmoke` (seed→get→update→filter→delete, both engines) |
| `src/database/sqlite.go` | **maybe** modify +≤5 lines — `file:`-URI pass-through guard (only if #418's branch mangles an existing query string; verify first) |

## File size guard (ER-4)

| File | `wc -l` now | After | Verdict |
|---|---|---|---|
| `src/database/testutil/*` | new | ≤ 140 each | ok |
| `src/database/parity/doc.go` / `parity_test.go` | new | ≤ 80 / ≤ 140 | ok |
| `src/database/sqlite.go` | (#418 planned ~100) | +5 | ok |
| touched legacy files | — | none | ok (Tier 1 untouched) |

## Implementation spec

### `harness.go` — D-2 verbatim

```go
var dsnSeq atomic.Uint64

func memoryDSN(t *testing.T) string {
    name := strings.ToLower(regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(t.Name(), "-"))
    if len(name) > 48 { name = name[:48] }
    return fmt.Sprintf("file:t425-%s-%d?mode=memory&cache=shared", name, dsnSeq.Add(1))
}

// OpenMemoryDB opens a uniquely-named in-memory SQLite DB through #418's factory.
// t.Cleanup closes it — a named shared-cache memory DB otherwise lives until every conn closes.
func OpenMemoryDB(t *testing.T) *gorm.DB {
    t.Helper()
    ctx, _ := RecordingContext()
    db, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
        Type: database.DBTypeSQLite, Path: memoryDSN(t),
    })
    require.NoError(t, err)
    t.Cleanup(func() { database.Close(ctx, db) }) // follow #418's Close signature
    return db
}

func MigrateSchema(t *testing.T, db *gorm.DB) // #419 entrypoint — CONFIRM NAME at task time:
    // grep -n "func AutoMigrate\|func Migrate" src/database/*.go; fallback: local model registry here.
func WithCleanConfig(t *testing.T) basecontext.ApiContext // t.Setenv("DATABASE_FOLDER", t.TempDir()); config.New(ctx) — pattern src/data/main_test.go:16-36
```

`sqlite.go` guard (only if needed — verify by opening `file:x?mode=memory&cache=shared` through the
factory *before* editing): in the SQLite branch, if `strings.HasPrefix(cfg.Path, "file:")`, use the
path verbatim as DSN and skip the `?_pragma=journal_mode(WAL)` append (WAL is meaningless for memory;
`busy_timeout` may be appended with `&`). One commit, comment `// #425 task 01: named-memory URIs from testutil pass through untouched`.

### `context.go` — recording fake (the #418 task-05 pattern, exported)

```go
type RecordingApiContext struct {
    basecontext.ApiContext        // embed real BaseContext for the 9 non-log methods
    mu sync.Mutex
    lines []string                // "warn: …" per LogWarnf etc.
}
func RecordingContext() (*RecordingApiContext, basecontext.ApiContext) // DisableLog() on the real ctx; LogXf append + no output
func (r *RecordingApiContext) Lines() []string
func (r *RecordingApiContext) Containing(substr string) []string
func (r *RecordingApiContext) HasLevel(level, substr string) bool
```
Embed-and-override, not a 14-stub fake — but **verify the compiler accepts it** (interface satisfaction
via the embedded value works for value-receiver methods; #418 built a full fake, mirror whichever fits).
Count the methods first: `grep -c '^\t[A-Z]' src/basecontext/api_context.go` → plan says 14, #418's
spec said 15; record the true number in the doc-comment.

### `engines.go`

```go
type EngineCapabilities struct{ HasNotConnectedSemantics, IsFileEngine bool }
type Engine struct { Name string; New func(t *testing.T) (interfaces.DatabaseService, func()); Capabilities EngineCapabilities }
type Scenario struct { Name string; Run func(t *testing.T, ctx basecontext.ApiContext, db interfaces.DatabaseService) }

func LegacyEngine() Engine  // New: dir := t.TempDir(); db := data.NewJsonDatabase(ctx, dir+"/test_db.json"); require Connect;
                            //   cleanup mirrors cleanupTestDB (close db.cancel where reachable — same-package field NOT exported:
                            //   parity lives OUTSIDE package data → use db.Disconnect(ctx) if public, else document that the
                            //   save-goroutine lives until process exit and t.TempDir() deletion still succeeds on macOS/Linux;
                            //   verify: run one legacy scenario 20× in a loop, watch fd count. Decision recorded in PR.
func GormEngine() Engine    // OpenMemoryDB + MigrateSchema + database.NewGormDatabase(...) — CONFIRM constructor name (OQ-425-2)
func RunOnEngines(t *testing.T, scens []Scenario, engines ...Engine) // t.Run(scen.Name+"/"+eng.Name); Capability skip → t.Skip(rationale from doc.go ledger)
```
**LegacyEngine honesty note (decide in-place, don't guess):** `setupTestDB`'s twin in package
`data` pokes `memoryDatabase = nil` and `db.cancel` — both unexported. If `NewJsonDatabase` + a
public disconnect path can't fully settle the goroutine, wrap the legacy leg as "construct fresh
instance per scenario, no global reset needed (we never touch the singleton)" — which is exactly D-3
rule 2. The save-interval goroutine writing to a `t.TempDir()` file until process exit is acceptable
(test process teardown; verify it never delays `go test` exit beyond ~1 s — else add `// TODO` and a
public `Stop()` request to #426).

### `parity/parity_test.go` — smoke scenario

```go
func TestParityUsersSmoke(t *testing.T) {
    testutil.RunOnEngines(t, []testutil.Scenario{{
        Name: "users_crud_filter",
        Run: func(t *testing.T, ctx basecontext.ApiContext, db interfaces.DatabaseService) {
            // seed 3 users (one root, like users_test.go) → GetUsers("") == 3
            // GetUsers("username = parity-bob") == 1 → update block status → re-read → delete → count 2
            // assert NOT-found: Get(nonexistent) → data.ErrUserNotFound
        },
    }}, testutil.LegacyEngine(), testutil.GormEngine())
}
```
Seed helper conventions shared by all waves (users seeded via `CreateUser`, root user idiom copied
from `src/data/users_test.go:14`): keep a `parity/seed_test.go`-style helper file *only if* it grows
in wave 02 — don't pre-build empty abstractions (CLAUDE.md: nothing more than asked).

### `harness_test.go` — the trust foundation

| Test | Assert |
|---|---|
| `TestSameNamedDSNShares` | open DSN manually twice (two `NewDatabaseConnection` with same `Path`), table created via 1st visible through 2nd |
| `TestDistinctDSNsIsolated` | `OpenMemoryDB` in two subtests; table/rows in one invisible in the other |
| `TestMemoryDSNShape` | sanitizer table: spaces/slashes/subtest names → `[a-z0-9-]` only, prefix `t425-`, ≤ full length cap, seq differs under repetition |
| `TestNoDiskWrites` | run `OpenMemoryDB`+migrate+insert with `DATABASE_FOLDER` pointed at a temp dir → dir stays empty (no `.db`/`-wal` files) |
| `TestOpenMigrateCost` | `t.Logf` only, never fails: time 20 iterations of OpenMemoryDB+MigrateSchema; **paste into PR** — if median > 40 ms, escalate D-9 fallback (a) *now*, before waves exist |
| `TestRecordingContextCaptures` | all 4 levels + trace captured with level prefix; ordering preserved |

## Out of scope

Any wave scenarios beyond the smoke (tasks 02–04) · Postgres engine (task 05) · error matrix (04) ·
deleting/changing anything under `src/data` (Tier 1) · `Makefile`/CI (05/07) · TESTING.md (#427 —
rules live in `doc.go`).

## Validation

- [ ] `cd src && go build ./...` && `go vet ./database/...`
- [ ] `cd src && go test ./database/testutil/... -count=1 -race -v` — leak/isolation/cost tests green
- [ ] `cd src && go test ./database/parity/... -count=1 -v` — `TestParityUsersSmoke/legacy` **and** `/gorm` both execute (grep CI log: both subtest lines present)
- [ ] `cd src && go test ./database/... -count=5 -shuffle=on` — isolation proof (D-3.4)
- [ ] `cd src && go test ./... -count=1` — suite-wide green (ER-5), record wall time in PR
- [ ] `TestOpenMigrateCost` number pasted in PR description (D-9 gate)
- [ ] If `sqlite.go` was touched: `git diff` shows ≤5 added lines, one commit, `#425 task 01` comment
- [ ] Ledger check: `grep -rn "serviceprovider\." src/database/testutil src/database/parity` → empty (D-3.2)

## Rollback

Revert commit: `testutil/` + `parity/` disappear (nothing imports them yet), `sqlite.go` guard
reverts too — zero blast radius on shipped code paths.

## PBI acceptance mapping

Moves **AC 2** (in-memory SQLite, unique-DSN isolation, no disk — harness + `TestNoDiskWrites`),
**AC 6** (isolation rules + `-count=5 -shuffle=on -race` evidence), starts **AC 1** (first legacy
function's parity twin executing), seeds the **≤ +30 s** evidence chain (cost measurement).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `425-01`.** Cut `slice/425-01` from `sql-migration` HEAD
(`.worktrees/slice-425-01`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
