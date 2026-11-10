# Task 03 — Validation UX matrix and secrets-safety regressions

**PBI** [#420](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 02 · **PR size:** S (test-dominant)
**Plan refs:** [D-7](./plan.md#d-7--error-semantics) · [D-8](./plan.md#d-8--secrets-safety) · feature plan ER-1 (idioms)

## Context

Task 02 authors the resolution logic; this task *proves and hardens the user-facing contract*: the
two pure-text ACs — "Missing or invalid configuration produces a clear, actionable error message at
startup" and "DSN values are never logged in plain text". The plan's audit (§D-8) verified there is
**no pre-existing config-dump path** to mask (grep evidence: only `src/install/main.go:181` writes
config values, and it writes the *user's own file*, never a log; `GetConfigContent`
(`src/config/main.go:1004`) feeds `install.saveConfig`'s file write at `src/install/main.go:683`,
never a log call). With no masking change to make, the durable artifact is a **regression suite**
that fails the moment anyone introduces a leak or an unhelpful message — plus surgical message fixes
if the matrix exposes gaps.

## Objective

Table-driven regressions pin, forever: (a) every `FromApplicationConfig`/`DatabaseAutoMigrate` error
names the offending setting (`DB_*` **and** flat file key) and — where applicable — the allowed
values; (b) no error message from #420's code paths ever contains a raw DSN or a canary password,
both DSN forms covered using #418's `RedactDSN`; (c) resolution mutates neither config state nor
the process environment. Build + suite green; zero behavior change unless a test exposes a gap.

## Files

| File | Op |
|---|---|
| `src/database/db_config_validation_test.go` | create (package `database`) |
| `src/database/db_secrets_test.go` | create (package `database`) |
| `src/config/database.go` | modify **only if** the matrix exposes a getter-message gap (e.g. `DatabaseAutoMigrate` error must name both notations) |
| `src/database/config.go` | modify **only if** a message gap surfaces (surgical, one-line-class fixes) |

## File size guard

| File | wc -l at plan time | After | Verdict |
|---|---|---|---|
| `src/database/db_config_validation_test.go` | new | ~180 | ok |
| `src/database/db_secrets_test.go` | new | ~150 | ok |
| `src/database/config.go` | ~200 after task 02 (**re-measure**) | ~210 | ok |
| `src/config/database.go` | ~140 after task 02 | ~145 | ok |

## Implementation spec

### `db_config_validation_test.go` — actionable-message matrix (AC "Validation and errors")

Helper first:

```go
func mustFailWith(t *testing.T, err error, wantIs error, mustContain ...string) // errors.Is + substring asserts
```

| Given (via `t.Setenv`; reset idiom: `config.New(basecontext.NewBaseContext())` w/ `DisableLog()`, per `src/data/main_test.go:16-36`) | Expect error |
|---|---|
| `DB_TYPE=mysql` | `errors.Is(err, ErrUnsupportedDatabaseType)`; message contains `DB_TYPE`, `db_type`, `"mysql"`, `sqlite`, `postgres` |
| `DB_TYPE` = 4 KB string | error message truncated (≤ ~200 chars per `quoteCapped`) — log-spam guard |
| `DB_TYPE=postgres`, DSN unset / DSN=`"   "` | `ErrPostgresDSNRequired`; message contains `DB_DSN` and `db_dsn` |
| `DB_AUTO_MIGRATE=maybe` | message contains `DB_AUTO_MIGRATE`, `db_auto_migrate`, `true`, `false` |
| file-key-only variants (`SetKey("db_type","mysql")` etc., no env) | same actionable text (parity AC) |
| sqlite + `DB_PATH=""` (impossible via defaults, guard future changes) | reaches `Validate()` → `ErrSQLitePathRequired` passed through unwrapped |

### `db_secrets_test.go` — no-leak regressions (AC "Secrets safety")

Canaries: `pwCanary := "sup3rs3cr3t"`; DSN forms `postgres://ops:`+pwCanary+`@db.internal:5432/app?sslmode=disable`
(URL style) and `host=db.internal user=ops password=`+pwCanary+` dbname=app` (key/value style) —
matching #418 `RedactDSN`'s two handled shapes (plan D-8).

| Test | Assert |
|---|---|
| postgres + DSN set (valid-looking, both forms) → `FromApplicationConfig` succeeds | no error to inspect; `cfg.DSN` verbatim (in-memory only) |
| postgres, DSN **blank** → error | `err.Error()` must not contain `password`, `://`, or any DSN remnant (D-7.3 never echoes; guard locks it) |
| garbage `DB_AUTO_MIGRATE` whose value *is* a DSN-looking string (`DB_AUTO_MIGRATE=postgres://u:sup3rs3cr3t@h/d`) | error echoes the value **truncated at 64 bytes** (task 02's cap; assert `len(msg) < 200` and the `…` marker) — spec decision: config-validation errors echo the user's own mis-set value back to *them*, bounded; the AC's "never logged" guarantee binds to **DSN-adjacent paths** (postgres errors, factory errors via #418 `RedactDSN`, provider log lines pinned in task 04). This test asserts **our additions never inject `cfg.DSN`** into any text — the canary here sits in the wrong setting, so it may legitimately appear truncated; the `cfg.DSN` injection check stays canary-clean |
| `FromApplicationConfig` (any valid case) mutates nothing | snapshot `config.Get().GetConfigContent()` bytes + `os.Environ()` before/after → equal (guards the D-8 rule "never `SetKey`/`Setenv` a DSN"; `GetConfigContent` is the marshal used by every config writer — `src/config/main.go:1004-1011`, `src/install/main.go:683`) |
| `RedactDSN` sanity (integration with #418) | both canary forms → returned string contains `***`/`[redacted` (per #418 D-8 contract) and never the canary |

If any test exposes a leak/gap, the **surgical fix** goes in `src/database/config.go` /
`src/config/database.go` (same PR, separate commit from the tests, message `fix(database): …`).
No defensive masking helpers are invented here — `RedactDSN` from #418 is the single funnel.

### Audit record (documentation-in-code)

Add one comment block atop `db_secrets_test.go` recording the §D-8 audit greps (file:line of
`Save`/`GetConfigContent`/`install.saveConfig`/installer env-copy) and the negative result — so a
future reader knows the masking question was answered, not skipped, and knows where to look first
if a new dump path ever appears.

## Out of scope

Startup exit-code/boot evidence (task 04 — the policy lives in #418 task 05's
`mustInitGormConnection`). Loader diagnostics for nested `database:` YAML (OQ-420-3). Log-level
wiring (GORM logger = #418). Any change to `Save`/`GetConfigContent` semantics. Any masking of
user-file contents (a config file holding the operator's own DSN is secret *storage*, not a leak).

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./database/... ./config/... -count=1 -v` — matrices green
- [ ] `cd src && go test ./... -count=1` (ER-5)
- [ ] `grep -rn "sup3rs3cr3t" src/database/*_test.go` — canary appears **only** in test literals, nowhere in non-test code
- [ ] `grep -rn "cfg.DSN\|DatabaseDSN()" src/database/config.go src/config/database.go` — every hit inspected: no DSN value in any error/log string construction (messages use sentinels/settings only)
- [ ] `git diff` — if no gaps surfaced, the diff is **two new test files only** (that is a success state, not a miss)

## Rollback

Revert commit — pure test additions (plus at most two surgical one-liners). Zero runtime surface;
no dependency from later tasks on the test files themselves (task 04 depends on task 02's API).

## PBI acceptance mapping

Moves AC boxes: **Validation and errors** (the actionable-message contract is *this* AC's final
word; startup delivery proven in task 04), **Secrets safety** (regressions + audit record; runtime
startup-line evidence in task 04). Hardens **Config file parity** (file-key-only error matrix).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `420-03`.** Cut `slice/420-03` from `sql-migration` HEAD
(`.worktrees/slice-420-03`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
