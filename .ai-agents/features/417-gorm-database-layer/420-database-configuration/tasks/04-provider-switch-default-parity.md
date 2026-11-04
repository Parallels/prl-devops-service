# Task 04 — Provider switch, default-parity proof, and startup matrix

**PBI** [#420](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 02 (hard); task 03 (soft — its message-contract tests make the matrix assertions meaningful, but no file overlap) · **PR size:** M
**Plan refs:** [D-1](./plan.md) (deviation protocol — this task executes it) · [D-7](./plan.md) (error semantics at runtime) · [D-8](./plan.md) (secrets safety, runtime evidence) · OQ-420-3 / OQ-420-4 (implemented resolutions to restate in PR)

## Context

Everything before this task is inert plumbing: the factory still opens with `DefaultDatabaseConfig()`
because #418's wiring calls it directly. This task performs the **one-identifier switch** that makes
#420 real at runtime — `FromApplicationConfig()` becomes the provider's source of truth — and then
*proves* the three ACs that only have runtime existence: backend selection for an operator, the
default-parity guarantee ("identical behaviour to the current application for users who make no
changes"), and startup-delivered actionable errors. It also discharges the D-1 deviation protocol
(PR note + #420 issue comment + #427 handover) so the flat-key deviation is as honestly tracked as
AD-2's.

## Objective

With nothing configured, boot behavior is byte-for-byte #418's; with `DB_TYPE=postgres` and a bad
DSN, the API command exits 1 with a redacted actionable message — and a build-tagged boot matrix +
unit tests document both.

## Files

| File | Op |
|---|---|
| `src/serviceprovider/database_connection.go` | modify — the flip (1 expression + 1 comment line) |
| `src/serviceprovider/database_connection_test.go` | modify — add switch tests + default-parity test |
| `src/integration_test/db_boot_matrix_test.go` | create (`//go:build integration`) |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `src/serviceprovider/database_connection.go` | *created by #418 task 05; projected ~75 at #420 start* | ~85 | ok (projection — re-run `wc -l` at implementation time; flip adds ≤10) |
| `src/serviceprovider/database_connection_test.go` | *#418 projection ~120* | ~260 | ok |
| `src/integration_test/db_boot_matrix_test.go` | new | ~150 | ok |

(`src/serviceprovider/main.go` untouched — the call sites and failure policy arrived in #418 task 05;
that is exactly why this PR is one identifier wide.)

## Implementation spec

### 1. The flip

```go
// database_connection.go — in initGormConnection:
- cfg := database.DefaultDatabaseConfig()
+ // PBI #420: operator settings (DB_TYPE / DB_PATH / DB_DSN / DB_AUTO_MIGRATE —
+ // config-file keys db_type/... flat, per D-1) now drive the connection.
+ cfg := database.FromApplicationConfig(ctx)
```

Error-handling shape unchanged (`initGormConnection` propagates; `mustInitGormConnection`'s
exit-1-under-`API_COMMAND` policy from #418 task 05 is the *existing* delivery mechanism for the
"clear error and exits gracefully" AC — this task proves it fires for config errors, not invents it).
Grep gate: after this PR, `grep -rn "DefaultDatabaseConfig()" src/serviceprovider/` → **zero hits**.

### 2. Unit tests (`database_connection_test.go`)

Pattern: `config.New(ctx)` reset + `t.Setenv` + recording-fake `ApiContext` (both exist in #418
task 05's test file — reuse, don't rebuild).

```go
func TestFromApplicationConfig_DefaultParity(t *testing.T)  // no DB_* env set at all:
// reflect.DeepEqual(database.FromApplicationConfig(ctx), database.DefaultDatabaseConfig(ctx))
// — this single assertion IS the "default configuration produces identical behaviour" AC.

func TestInitGormConnection_SwitchesByEnv(t *testing.T)     // table:
```

| # | env | expect |
|---|---|---|
| a | `DATABASE_FOLDER=tmp` only | nil error; file `prl-devops.db` in tmp (path default flows through — OQ-420-4 folder branch) |
| b | `DB_TYPE=postgres` (no DSN) | `ErrPostgresDSNRequired` surfaced from initGormConnection; `provider.SQLDatabase == nil`; warn (not exit) because test command ≠ `api` — policy already tested in #418, assert the error value, not os.Exit |
| c | `DB_PATH=tmp/x.db` + `DATABASE_FOLDER=tmp/y` | file lands at `tmp/x.db` (OQ-420-4 precedence implemented and pinned here) |
| d | `DB_AUTO_MIGRATE=maybe` | `ErrInvalidAutoMigrateValue` (task 02 sentinel) — config error, no connection attempted |

All with `ctx.DisableLog()`-style quiet contexts; assert on returned errors, not logs (logs are
task 03 + matrix territory).

### 3. Boot matrix (`src/integration_test/db_boot_matrix_test.go`)

Package conventions: `//go:build integration`, package `integration_test` (see `setup_test.go`).
**Opt-in guard first:** `if os.Getenv("PRLDEVOPS_BOOT_MATRIX") != "1" { t.Skip("opt-in boot matrix — see PR body") }` —
full API boot is heavy (module init, install config, ports); it runs deliberately, not accidentally.

```go
// bootAPI runs `<binary> api` with a minimal env (DATABASE_FOLDER=t.TempDir(), plus case env),
// returns combined stdout+stderr and exit code, killing the process after `deadline` (30s).
// Success = the "SQL database connection ready" info line appears within the window.
func runAPI(t *testing.T, env []string, deadline time.Duration) (output string, exited bool, code int)
```

| # | env (beyond DATABASE_FOLDER) | expect |
|---|---|---|
| a | *(none)* | success line present; `prl-devops.db` created in the temp folder (default parity at the **process** level) |
| b | `DB_TYPE=postgres` | exited, code 1, output contains `"required for database type \"postgres\""` and **no** canary (no DSN here, sanity row) |
| c | `DB_TYPE=postgres`, `DB_DSN=postgres://dbuser:sup3rs3cr3t@127.0.0.1:1/x?sslmode=disable` | exited, code 1; output contains `dbuser:***@127.0.0.1:1` and **must not** contain `sup3rs3cr3t` — the D-8 canary, at the real startup line |
| d | `DB_AUTO_MIGRATE=maybe` | exited, code 1, message quotes `maybe` |

Binary discovery at test top: `go build -o t.TempDir()/prldevops .` inside the test (module `src/`),
or reuse a `PRLDEVOPS_BIN` env pointer if CI already exports one (check `.github/workflows/pull_request.yml`
at implementation time — do not add a new CI step in this task; matrix is opt-in, same posture as
#418's manual boot evidence). Document the exact local invocation in the test file header:
`PRLDEVOPS_BOOT_MATRIX=1 go test -tags integration ./integration_test/... -run BootMatrix -v`.

If a row reveals the API boot needs more env than assumed (e.g. install config resolution), the
implementer adjusts the **minimal env** the harness injects — *never* relaxes the assertions — and
records the discovery in the PR body (it becomes #425/#427 knowledge).

### 4. Deviation protocol (D-1) — PR-body requirements

- [ ] PR description states the flat-key deviation verbatim from plan D-1 (one paragraph).
- [ ] Comment posted on issue #420 so the literal `database.path` AC boxes are checked with the
      known deviation (OQ-420-1) — mirror of #418's AD-2 protocol.
- [ ] Handover note to the #427 tracker (issue comment on #427 or its folder's open-questions, pick
      whichever exists): docs publish `db_type`/`db_path`/`db_dsn`/`db_auto_migrate`; dotted form must
      never appear in user docs.

## Out of scope

Nothing else touches `serviceprovider/main.go`; no pool env vars (`DB_PG_*` — deferred, plan Scope);
no nested-key loader work (D-1); AutoMigrate execution; CI changes (matrix is opt-in by env);
`src/config/main.go` (already split in task 01 — do not re-touch).

## Validation

- [ ] `cd src && go build ./... && go vet ./config/... ./database/... ./serviceprovider/...`
- [ ] `cd src && go test ./config/... ./database/... ./serviceprovider/... -count=1 -v` — new rows green
- [ ] `cd src && go test ./... -count=1` — full-suite parity, no regressions (ER-5)
- [ ] `grep -rn "DefaultDatabaseConfig()" src/serviceprovider/` → empty (flip gate)
- [ ] Boot matrix rows a–d run locally (`PRLDEVOPS_BOOT_MATRIX=1 …`), output pasted in PR —
      row c's canary-absence is the headline evidence
- [ ] Deviation protocol items (4 boxes) done
- [ ] AC parity quote in PR: "Default configuration produces identical behaviour" ↔ `TestFromApplicationConfig_DefaultParity` + matrix row a

## Rollback

Revert = switch back to `DefaultDatabaseConfig()` + delete the two test additions/files. One-
identifier blast radius; #418's runtime posture returns unchanged; every other #420 artifact
(constants, getters, seam, tests) stays valid and dormant.

## PBI acceptance mapping

Runtime completion of **Backend selection**, **SQLite path**, **PostgreSQL DSN**, **Auto-migrate
flag** (operator-visible behavior), final proof for **Validation and errors** (startup delivery)
and **Secrets safety** (matrix row c), and the whole of **Config file parity** + **Validation** at
process level; this is the task where AC "default configuration produces identical behaviour"
(constraint) becomes *demonstrated*, not asserted.
