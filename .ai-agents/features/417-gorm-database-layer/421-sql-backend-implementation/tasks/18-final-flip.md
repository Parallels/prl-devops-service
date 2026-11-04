# Task 18 — S3 final flip: provider selects by `DB_TYPE`; legacy bypassed

**PBI** [#421](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 17 (all domains ported; consumers interface-typed) · **PR size:** L
**Plan refs:** D-6 step S3 · D-7 (every final-column row) · D-8 (+OQ-421-4) · risk-table row "Flip-day regressions: rollback = env var, not revert" · AC map rows *file path / WAL / bypass*

## Context

The flip itself is small — a selection branch in the provider — but it *activates* everything: goroutine
ownership (save-queue / stale-job watcher / backup), the legacy schema-migrator gate, the panic-snapshot
proof on SQL, PostgreSQL enablement, and the bypass audit that closes the AC "Legacy layer bypassed when
SQL database configured". Post-flip default: **`DB_TYPE` unset ⇒ `sqlite` ⇒ `GormDatabase`** — with the
documented `DB_TYPE=legacy` escape hatch as the rollback lever (feature risk posture; #426 owns the
warn-and-deprecate story for that fallback).

## Objective

`FromApplicationConfig().Type` selects the engine; SQLite and Postgres both boot; legacy goroutines exist
**only** in legacy mode; panic-snapshot works on SQL via `VACUUM INTO`; bypass audit green in PR; one env
var away from rollback.

## Files

| File | Op |
|---|---|
| `src/serviceprovider/main.go` | modify — selection in init (530 lines, measured; selection block ≈35 lines added, ≤565) |
| `src/serviceprovider/database_connection.go` | modify — expose `buildGormDatabase(ctx, cfg)` helper if 01/04 haven't already |
| `src/startup/main.go` | modify — gate legacy schema-migrator loop (~:128 per D-7) + `migrations/service.go` DbService nil-guard + SQL-mode job hooks (Recover + watcher start) |
| `src/main.go` | modify — panic-snapshot path exercises `SaveAs` through the interface (17 prepared the type; 18 proves the behavior) |
| `src/database/lifecycle_sql_test.go`, `src/integration_test/db_flip_matrix_test.go` | create |

## File size guard

| File | Lines | After | Verdict |
|---|---|---|---|
| `src/serviceprovider/main.go` | 530 (measured) | ~565 | ok |
| `src/startup/main.go` | *(measure at task start — pre-edit 1,273 in early recon; if still >1,000: type-only/guard edits are ≤5 lines each — ER-3 exception posture like D-10, note it)* | — | watch |
| new test files | — | ~150 + ~180 | ok |

## Implementation spec

1. **Selection (provider init, after #420's connection init):**

```go
cfg := database.FromApplicationConfig(ctx)
switch cfg.Type {
case "", "legacy":                       // explicit escape hatch; "" = pre-flip default
    // #426 adds the warn here — for now behaviorally identical to today
case "sqlite", "postgres":
    gormDB := buildGormDatabase(ctx, cfg) // *GormDatabase over #418's *gorm.DB
    provider.SQLDatabase = gormDB
    provider.legacyDatabase = nil          // concrete handle stays nil ⇒ legacy goroutines never start
}
```

   `GetDatabaseService` (interface-typed since 17) returns `SQLDatabase` when non-nil, else legacy.
   **Default flips to `sqlite` via #420's default `DatabaseType`** — the "" case exists only for the
   escape hatch and pre-migration boots (#424 detection runs under SQL mode, it doesn't need legacy mode).
2. **Goroutine ownership (D-7/D-8):** legacy `Connect` (save-queue, stale-job, backup loops) runs only when
   legacy selected; SQL mode starts **only** the stale-job watcher (task 15's function, interval constant
   reused) and calls `RecoverOngoingJobs` once at startup (D-8/OQ-421-4 always-on; PR flags for PO). Stop
   hooks on provider teardown path mirroring #418's save-loop stop.
3. **Legacy schema-migrator gate (D-7 SchemaVersion row):** `src/startup/main.go` migrator loop skips when
   SQL selected (`SchemaVersion`/`UpdateSchemaVersion` served by `schema_metadata` row — task 08);
   `startup/migrations/service.go` DbService nil-guard so a nil legacy handle can't panic a SQL boot.
4. **Panic-snapshot proof:** with `DB_TYPE=sqlite`, force the `src/main.go:59` path in a subprocess test
   (build binary; trigger the crash condition the legacy snapshot test uses, or call the cleanup func with a
   non-nil error under `API_COMMAND`): assert `<path>.panic-snapshot-*` materialises, opens with the SQLite
   driver, contains migrated rows. `VACUUM INTO` semantics: absolute path required — the D-7 helper from
   task 04's lifecycle file handles it; Postgres branch = log-once warn + no-op (documented gap → #427,
   OQ-421-3).
5. **Flip matrix (`db_flip_matrix_test.go`, build tag `integration`, opt-in per #420's `PRLDEVOPS_BOOT_MATRIX`
   pattern or its own guard):** boot API with (a) default env → SQL mode lines, `prl-devops.db` created at
   #420-resolved path (**AC "file path respected"**), `PRAGMA journal_mode` returns `wal` **over the provider
   handle** (AC WAL); (b) `DB_TYPE=legacy` → today's log sequence + legacy goroutine lines present;
   (c) `DB_TYPE=sqlite` + `DB_PATH` → file at explicit path; (d) `DB_TYPE=postgres` + docker DSN (reuse #418's
   canonical one-liner) → connects, `AutoMigrate` green, **and** `sqlite_master`-equivalent introspection
   finds the tables (pg `information_schema` query).
6. **Bypass audit (PR body, closes the AC):** grep block + verdicts:

```bash
grep -rn "\.JsonDatabase\b" src --include=*.go | grep -v _test | grep -v '^src/data/' | grep -v serviceprovider/
grep -rn "SaveNow\|SaveAsync" src --include=*.go | grep -v _test | grep -v '^src/data/'   # now interface no-ops in SQL
grep -rn "\.Legacy()" src --include=*.go | grep -v _test                                   # = the 17-retained rows + startup fallbacks only
```

   Any survivor outside the 17 retained-rows list = blocking.

## Out of scope

Deprecation warnings & `data.json` freeze messaging (#426), docs (#427), migration runner (#424 — runs
before/parallel to this via its own startup slot; the matrix boots against an *empty* SQL store, which is
#424's clean-slate shape), test-suite conversions (#425).

## Validation

- [ ] `cd src && go build ./... && go vet ./... && go test ./... -count=1` (ER-5)
- [ ] Flip matrix rows a–d output pasted in PR (row d needs docker; mark skipped-in-CI status)
- [ ] Panic-snapshot subprocess evidence (file path + row-count smoke)
- [ ] Bypass audit greps in PR body, zero unexplained survivors
- [ ] Goroutine evidence: `DB_TYPE=sqlite` boot logs contain **no** legacy save-loop line ("Starting json
      file saving routine..."); SQL watcher line present; `DB_TYPE=legacy` boot shows the inverse
- [ ] Rollback documented: `DB_TYPE=legacy` row-b output IS the rollback proof

## Rollback

**Runtime:** `DB_TYPE=legacy` (no deploy). **Structural:** revert this commit — providers return legacy,
goroutines return; domains 05–16 stay ported but dormant (delegate-free code just unused — tasks 05–16 stay
shippable); S2 (17) unaffected. No schema rollback needed (tables additive).

## PBI acceptance mapping

The closing PR for every remaining box: **Backend selection** (sqlite/postgres/legacy rows), **SQLite file
path** (row a/c), **PostgreSQL** (row d), **Transactions** (D-9 fully live under selection), **Concurrent
access WAL** (row a, provider handle), **Legacy layer bypassed** (audit + goroutine evidence). Post-merge:
#424/#425/#426 unblock *for real* — SQL is now the shipping default path.
