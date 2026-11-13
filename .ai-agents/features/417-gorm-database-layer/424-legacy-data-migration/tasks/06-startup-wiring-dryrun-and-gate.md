# Task 06 — Startup wiring, `DB_AUTO_MIGRATE` gate, dry-run flag/env, exit policy

**PBI** [#424](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** tasks 01–05; **#420
merged** (`DatabaseAutoMigrate()` accessor over `DB_AUTO_MIGRATE`/`database.auto_migrate`) · **PR size:** M
**Plan refs:** [D-9](../plan.md#d-9--dry-run-ac-7--startup-hook--failure-policy), locked **AD-4** (ordering)

## Context

First PR with runtime effect: from here, an upgrade boot performs the migration. The hook slots exactly
where AD-4 orders it — after the legacy schema-migrations loop (`src/startup/main.go:135-140`,
`schemaMigrations` applied at `:136-140`), before serving; GORM AutoMigrate has already run at
provider-init (feature plan §8). Failure policy mirrors #418 task 05 `mustInitGormConnection`: API
command ⇒ log clearly + `os.Exit(1)` (transaction already rolled back by `db.Transaction`); any other
command ⇒ warn and continue so CLI/tests stay usable (and so a `data` command can still run a dry-run).
Dry-run flag/env uses the proven idiom: constants in `src/constants/main.go` read via
`helper.GetFlagValue(constants.CONFIG_FILE_FLAG, "")` (`src/config/main.go:122`; env-last-wins ordering
like `Config.GetKey`, `src/config/main.go:933-947`).

## Objective

Boot sequence gains one step: gated, detected, migrated-or-skipped, validated, logged — with
`--migrate-dry-run=true` / `DB_MIGRATE_DRY_RUN=true` short-circuiting all writes.

## Files

| File | Op |
|---|---|
| `src/constants/main.go` | modify — 2 constants (flag block `:160-175`, env block `:82-98`) |
| `src/database/legacy_migrator/dryrun.go` | create |
| `src/database/legacy_migrator/dryrun_test.go` | create |
| `src/startup/legacy_migration.go` | create — `runLegacyDataMigration(ctx) error` (new file, package `startup`) |
| `src/startup/main.go` | modify — call site after the migrations loop |
| `src/startup/legacy_migration_test.go` | create |

## File size guard

Real `wc -l` (2026-11-03): `constants/main.go` **560** → ~563 ok · `startup/main.go` **345** → ~352
call-site only ok · `startup/legacy_migration.go` new ≈ 55 ok · `dryrun.go` ≈ 60 / test ≈ 90 ok.
**Not touched:** `config/main.go` **1,013** — over ER-2 cap; this task must NOT add accessors there
(#420 owns its ER-3 split); flag/env read directly. Verdict: **ok**.

## Implementation spec

### `constants/main.go`

```go
// env block (next to DB_* siblings landing via #420):
MIGRATE_DRY_RUN_ENV_VAR = "DB_MIGRATE_DRY_RUN"
// flag block (next to CONFIG_FILE_FLAG at :171):
MIGRATE_DRY_RUN_FLAG    = "migrate-dry-run"
```

### `dryrun.go`

```go
// IsDryRunRequested resolves AC-7's input. Precedence mirrors Config.GetKey (src/config/main.go:933-947):
// env var beats CLI flag. Accepted truthy values: "true","1","yes" (case-insensitive) — helper.GetFlagValue
// needs a VALUE, so the operator spellings are --migrate-dry-run=true / --migrate-dry-run true;
// a value-less flag is inert, exactly like every other repo flag (documented for #427).
func IsDryRunRequested() bool {
        flag := strings.TrimSpace(helper.GetFlagValue(constants.MIGRATE_DRY_RUN_FLAG, ""))
        if v, ok := os.LookupEnv(constants.MIGRATE_DRY_RUN_ENV_VAR); ok { flag = v }
        switch strings.ToLower(strings.TrimSpace(flag)) {
        case "true", "1", "yes": return true
        default: return false
        }
}

// ReportDryRun prints the AC-7 report; called for report.DryRun.
func ReportDryRun(ctx basecontext.ApiContext, r Report) {
        ctx.LogInfof("[dry-run] legacy migration NOT executed; %d collection(s) would be written:", len(r.Collections))
        for _, c := range r.Collections {
                ctx.LogInfof("[dry-run] would write %d record(s) into %s", c.WrittenCount, c.Name)
        }
        ctx.LogInfof("[dry-run] no data was written (no marker; data.json untouched); migration did NOT run") // OQ-C wording
}
```

### `startup/legacy_migration.go`

```go
// runLegacyDataMigration executes PBI #424 between legacy schema migrations and serving (AD-4 order).
// Gate lives at the call site (cfg.DatabaseAutoMigrate(), #420). Never modifies data.json.
// Test seams: package-level vars (overridden in _test.go, restored via t.Cleanup) so tests never
// construct the real ServiceProvider — JsonDatabase.filename is unexported (src/data/main.go:59), so
// the path is reached ONLY through Filename() (`:188-190`) behind this seam.
var (
        getSQLDatabase = serviceprovider.GetSQLDatabase // #418 task 05
        legacyStoreSource = func() (string, error) {    // resolves <folder>/data.json per provider convention
                provider := serviceprovider.Get()
                if provider == nil || provider.JsonDatabase == nil {
                        return "", errors.NewWithCode("legacy database provider not initialized; cannot locate data.json", 500)
                }
                return provider.JsonDatabase.Filename(), nil
        }
)

func runLegacyDataMigration(ctx basecontext.ApiContext) error {
        gdb, err := getSQLDatabase(ctx) // #418 task 05; nil/err ⇒ #418 policy already ruled
        if err != nil { return err }
        sourcePath, err := legacyStoreSource()
        if err != nil { return err }
        opts := legacy_migrator.Options{
                SourcePath: sourcePath,               // <folder>/data.json, resolved by the provider (AD-3)
                DryRun:     legacy_migrator.IsDryRunRequested(),
        }
        report, err := legacy_migrator.RunMigration(ctx, gdb, opts)
        if err != nil { return err }
        if report.DryRun { legacy_migrator.ReportDryRun(ctx, report) }
        return nil
}
```

### `startup/main.go` call site (immediately after the loop at `:136-140`)

```go
ctx.LogInfof("Applying migrations")
for _, migration := range schemaMigrations { … }        // existing code — unchanged

// Legacy data migration (PBI #424): legacy JSON store → GORM schema, once, idempotent.
if cfg.DatabaseAutoMigrate() {                            // #420 accessor: DB_AUTO_MIGRATE, default true
        if err := runLegacyDataMigration(ctx); err != nil {
                ctx.LogErrorf("Legacy data migration failed (all changes rolled back): %v", err)
                if cfg.GetRunningCommand() == constants.API_COMMAND {
                        os.Exit(1)                        // AC-3 exit policy; precedents: #418 task 05,
                }                                         // src/main.go:59-63
        }
}
```

`os` is already imported in `startup/main.go`? — it imports `encoding/base64`, `time`, … : **verify and
add `"os"` if absent** (trivial; noted in task body so the implementer doesn't get bitten).
`cfg.DatabaseAutoMigrate()` comes from #420; if #420 shipped a differently-spelled accessor, use theirs
(grep `auto_migrate` in `src/config` at implementation time) — the gate, not its spelling, is the spec.

### Tests

| Test | Assert |
|---|---|
| `IsDryRunRequested` table | unset⇒false; `DB_MIGRATE_DRY_RUN=true/1/yes/TRUE`⇒true; `false`⇒false; **env overrides flag** (flag=false+env=true ⇒ true) — flag rows inject `os.Args` via a package-var seam `osArgs()` if needed (document; common-go reads `os.Args` directly) |
| `runLegacyDataMigration` fresh store (temp `data.json` + `:memory:` gdb via provider stub) | rows migrated; marker row present |
| gate skip | with AutoMigrate-false the call site is skipped — tested at the helper level by asserting nothing happens preconditions-wise; call-site gate is review-verified (paste `startup/main.go` diff in PR) |
| failure ⇒ rollback ⇒ error propagated | `Options.FailAfter` injected via stub; `os.Exit` NOT reached outside API_COMMAND (test never sets it; policy is one `if` reviewed + task 07 subprocess proof) |
| dry-run report | info lines captured; DB still empty; marker absent |

## Out of scope

`config/main.go` accessors (frozen/over-cap, #420's), #426's frozen-legacy warn, Helm/docs (#427).

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go vet ./startup/... ./database/... ./constants/...`
- [ ] `cd src && go test ./database/legacy_migrator/... ./startup/... -count=1` green
- [ ] `cd src && go test ./... -count=1` regression green (provider tests must be unaffected: gate + non-API command ⇒ never exits)
- [ ] Manual boot matrix on a dev box (paste logs): (a) no data.json ⇒ silent skip; (b) golden data.json ⇒ migrate + validate + marker; (c) restart ⇒ `already-migrated` skip; (d) `DB_MIGRATE_DRY_RUN=true` + fresh ⇒ report only, `data.json` sha unchanged, new DB empty; (e) corrupt data.json under API ⇒ clear error + exit 1
- [ ] `grep -n "config/main.go" <PR diff>` ⇒ empty (over-cap file untouched)

## Rollback

Revert PR removes the call site (startup reverts to pre-#424 boot); already-migrated installs keep
their marker + data harmlessly (GORM path keeps working; legacy file untouched by design).

## PBI acceptance mapping

Completes **AC-1 Detection** (wired, silent skip), **AC-3** exit semantics, **AC-7 Dry-run** (flag +
env + report), and realises the "automatic, no manual steps" user story.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `424-06`.** Cut `slice/424-06` from `sql-migration` HEAD
(`.worktrees/slice-424-06`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
