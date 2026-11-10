# PBI Plan #420 — Database configuration via environment variables and config file

> Reading order: [feature brief](../brief.md) → [feature plan](../plan.md) (AD-3, AD-4, ER-1…ER-5) →
> [PBI brief](./brief.md) → this plan → your task spec.
>
> Prerequisite state: **#418 fully merged** (its tasks 01–05). All line counts below re-measured
> 2026-11-03 against HEAD (see each task's `File size guard`).

## Scope

**In:** `DB_TYPE` / `DB_PATH` / `DB_DSN` / `DB_AUTO_MIGRATE` env-var + config-file settings · constants
in `src/constants/main.go` · typed getters on `*Config` · **filling the `FromApplicationConfig()`
seam** #418 left in `src/database/config.go` (defaults per AD-3, tolerant-but-strict parsing,
actionable errors built on #418's sentinels) · flipping `src/serviceprovider/database_connection.go`
from `DefaultDatabaseConfig()` to `FromApplicationConfig()` (the runtime moment #420 becomes real) ·
secrets-safety audit + regression tests · startup failure-message checks · ER-3 split of
`src/config/main.go` · unit tests throughout.

**Out:** anything the loader can't do without redesign — nested/dotted config keys (`D-1` explains
the deviation and why nested `database.*` support is deliberately **not** added); PostgreSQL pool
tuning env vars (`DB_PG_*` — #418 handover explicitly defers the call to us, we defer it further:
no AC requires them; see Open questions); AutoMigrate execution (#419/#421); legacy-store selection
flip (#421); user documentation (#427 — we only hand over the deviation notes); Helm values (#427/OQ-4);
graceful shutdown hook (#418 task 05 already parked it).

## Design decisions

### D-1 — Config-file keys are FLAT (`db_type`, `db_path`, `db_dsn`, `db_auto_migrate`); the issue's `database.*` notation is a documented deviation

The loader's file shape is **verified**: `ConfigFile.Environment` is a `map[string]string` under the
YAML top-level `environment:` key (`src/config/models.go:9`, file is 72 lines), and `Config.GetKey`
scans that flat map with `strings.EqualFold` (`src/config/main.go:933-941`), then CLI flag, then
`os.LookupEnv` — **env wins** (`src/config/main.go:943-951`). Existing config files use lowercase
snake keys matching env names — `src/config.local.yaml:1-17` (`api_port`, `database_folder`, …).

Consequence: the issue/feature-brief §3 notation `database.type` **cannot resolve** through `GetKey`
as-is. Worse, writing literal YAML

```yaml
database:
  type: sqlite
```

into a config file makes `yaml.Unmarshal` into `map[string]string` fail for the **whole file**
(`Load` logs "Error reading configuration file" and returns false, `src/config/main.go:147-160`).

**Decision:** the config-file key *is* the env-var name, lower-snake — `db_type`, `db_path`,
`db_dsn`, `db_auto_migrate` — matched case-insensitively because `GetKey` compares with
`strings.EqualFold` (a file key `DB_TYPE` would also match; the docs use lowercase for
consistency with `database_folder`). Zero loader changes; exact repo convention; satisfies the
"all env vars have equivalent config file keys; env vars take precedence" AC *for free* through the
mechanism itself.

**Required (mirrors AD-2's deviation protocol):** every PR touching this states the deviation in
its description; a comment goes on #420 so the literal AC boxes ("`database.path` config key") are
checked with the known deviation; a handover note goes to #427 (docs must publish the flat keys,
never the dotted ones). No nested-key loader support is added (ER-1: "no unrelated re-plumbing";
it would also change behavior for every other setting). See Risks R-420-1 and Open question OQ-420-1.

### D-2 — Env-var/constant names: `DB_TYPE` / `DB_PATH` / `DB_DSN` / `DB_AUTO_MIGRATE`

Constants appended to the env-var block in `src/constants/main.go` (block spans lines 82–138; the
existing DB set `DATABASE_FOLDER_ENV_VAR` sits at line 101, `…_ENV_VAR` naming is the pattern —
verified there are **no** existing `DB_*` constants, so no collision). Values exactly as the AC
names them: `"DB_TYPE"`, `"DB_PATH"`, `"DB_DSN"`, `"DB_AUTO_MIGRATE"`. Allowed `DB_TYPE` values
reuse #418's `database.DBTypeSQLite` / `DBTypePostgres` constants (defined in #418 task 02 at
`src/database/config.go`) — one source of truth; `config` must not import `database`, so the
*validation* against those constants lives in `database.FromApplicationConfig` (task 02), while the
`config` getters stay dumb strings (D-3).

### D-3 — Typed getters on `*Config`, dumb by design

Following the `DatabaseFolder()` one-liner pattern (`src/config/main.go:598-600`), four new methods
on `*Config`, placed in the new `src/config/database.go` (created by task 01's ER-3 split — the
cohesive home for DB-related getters, next to `DbSaveInterval`, `IsDatabaseAutoRecover`, …):

| Getter | Returns | Semantics |
|---|---|---|
| `DatabaseType() string` | raw `GetKey(DB_TYPE)` | `""` = unset → caller defaults |
| `DatabasePath() string` | raw `GetKey(DB_PATH)` | `""` = unset → AD-3 resolution |
| `DatabaseDSN() string` | raw `GetKey(DB_DSN)` | `""` = unset |
| `DatabaseAutoMigrate() (bool, error)` | parsed `GetKey(DB_AUTO_MIGRATE)` | `""` → `true, nil`; parse rules below |

Why dumb (raw strings, no defaults): import direction — `database` imports `config` (already true
in #418 task 02, `defaultDatabasePath` calls `config.Get().DatabaseFolder()`), never the reverse;
defaults and validation therefore live in `database.FromApplicationConfig` — the single bridge
the brief's "single source of truth, no second resolver" line requires. The `(bool, error)` shape
has a same-package precedent for errorful getters: `ValidateTlsConfiguration() (bool, string)`
(`src/config/main.go:338`). `src/errors` imports only `fmt`/`strings` (`src/errors/main.go:3-6`),
so `config` importing it introduces **no cycle** (config already imports
`serviceprovider/system`, `src/config/main.go:20` — a heavier edge than this).

`DatabaseAutoMigrate` parse rules (task 02 spec): trim + lowercase; true-set
`true/1/t/yes/y/on/enabled`, false-set `false/0/f/no/n/off/disabled` — the repo's tolerant-bool
vocabulary is `helpers.StringToBool` (`src/helpers/strings.go:144-158`) and the `IsDatabaseAutoRecover`
list (`src/config/main.go:525-538`), extended with an explicit false-set and — deliberately
differently from both — **an error instead of silent `false` for garbage**, because AC "Validation
and errors" demands it. Empty/unset → `true` (AC: "Defaults to `true`"; AD-4 confirms the flag
gates AutoMigrate default-on).

### D-4 — Seam signature grows an error: `FromApplicationConfig() (DatabaseConfig, error)`

#418 task 02 declared `func FromApplicationConfig() DatabaseConfig { return DefaultDatabaseConfig() }`
(no error, documented `TODO(#420)`), and **#418 task 05's provider calls `DefaultDatabaseConfig()`,
not the seam** (`src/serviceprovider/database_connection.go`, `cfg := database.DefaultDatabaseConfig()
// seam: #420 swaps to FromApplicationConfig()`). Because nothing calls the seam yet, widening it in
task 02 breaks no build (ER-5 safe); task 04 then writes the *first and only* consumer. Rationale for
the error return: AC "Missing or invalid configuration produces a clear, actionable error message at
startup" needs a typed failure path from parsing (garbage `db_auto_migrate`, unknown `db_type`) up to
`mustInitGormConnection`, whose exit policy already exists in #418 task 05:
`ctx.LogErrorf(...)` + `os.Exit(1)` when `config.Get().GetRunningCommand() == constants.API_COMMAND`
(`API_COMMAND = "api"`, `src/constants/main.go:142`; set at `src/cmd/main.go:24`), warn-and-continue
otherwise (CLI stays usable to *fix* a bad DB config). `DefaultDatabaseConfig()` keeps its signature
and remains the reference for default-parity tests (D-5).

### D-5 — Defaults delegate to #418; zero duplication; byte-identical no-config behavior

`FromApplicationConfig()` on unset inputs must produce a `DatabaseConfig` **equal to**
`DefaultDatabaseConfig()`: `Type = database.DBTypeSQLite`; `Path` = the same AD-3 resolution
(`DB_PATH` → else `config.Get().DatabaseFolder()` join `prl-devops.db` → else uid-0
`/etc/parallels-devops` → else `~/.parallels-devops`) — achieved by *delegating to*
`defaultDatabasePath()` inside #418's `config.go`, never re-implementing it (AD-3: "do not duplicate
the logic in a second place"); `AutoMigrate = true`. The AC phrase "Defaults to the current legacy
database file path" means the legacy **folder** convention with the new file name `prl-devops.db`
(AD-3, human-approved — brief §"Repo context" restates this). The default-parity property
(no `DB_*` set anywhere ⇒ runtime behavior byte-identical to today) is pinned by a unit test in
task 04 (`reflect.DeepEqual(FromApplicationConfig(), DefaultDatabaseConfig())` under a clean env) —
and remains true for users until #421 flips the store, because nothing consumes `SQLDatabase` yet.
Precedence chain per key: **env var > CLI flag > config-file key > built-in default** — the first
three are enforced by `GetKey` itself (`src/config/main.go:933-951`, verified order: file map
scanned first, flag overrides, env overrides last).

### D-6 — ER-3 split of `src/config/main.go`: four cohesive files, `main.go` core (measured)

`src/config/main.go` is **1,013 lines** — over the ER-2 cap; every #420 edit is blocked until the
pure-move split lands (task 01). The feature plan's file list (`database.go`, `modules.go`,
`logging.go`) is a *suggestion*; measured function-group sizes (awk over real extents, 2026-11-03)
justify a slightly different cohesive split, and one thin `logging.go` (30 lines around the single
`LogLevel` function, main.go:228-253) is dropped in favor of keeping `LogLevel` in core next to
`New`/`Load` which call it (main.go:96, :163):

| New file (same package `config`) | Functions moved | Measured body lines |
|---|---|---|
| `src/config/database.go` | `DbNumberBackupFiles`, `DbBackupInterval`, `DbSaveInterval`, `DbGhostJobTimeoutMinutes`, `IsDatabaseAutoRecover`, `DatabaseFolder` | 58 → ~70 with header; **grows to ~140 in task 02** with the D-3 getters |
| `src/config/modules.go` | `GetEnabledModules` (137 lines itself, main.go:760), `IsModuleEnabled`, `DisableModule`, `EnableModule`, `IsOrchestrator`, `IsCatalog`, `IsCatalogManager`, `IsHost`, `IsApi`, `UseOrchestratorResources` | 205 → ~215 |
| `src/config/network.go` | `ApiPort`, `ApiPrefix`, `IsCorsEnabled`, all `Tls*`, `ValidateTlsConfiguration`, `ShouldDisableHttpWhenTls`, `DisableTlsValidation`, all `ReverseProxy*`, `Localhost` | 186 → ~196 |
| `src/config/platform.go` | `RootFolder`, cache family (`CatalogCacheFolder`…`ForceCacheRefreshInterval`), `Mode`, `SystemReserved*`, `OrchestratorPullFrequency`, `OrchestratorPublicUrl`, `IsCanaryEnabled`, `IsBetaEnabled`, `IsRemoteProviderStreamEnabled`, `EnableCredentialsObfuscation`, `CatalogCompress*` | 208 → ~220 |
| `src/config/main.go` (core, stays) | struct + `New`/`Get`/`Load`/`Save`/`Refresh`, `GetKey`/`SetKey` family, `GetConfig`/`GetFilename`/`GetConfigContent`, `EncryptionPrivateKey`, `RootPassword`, `Source`, `LogLevel`, command accessors, file discovery | 657 lines move out → **~370** incl. imports |

Pure move only — no renames, no signature edits, no test edits (`main_test.go`/`modules_test.go`/
`job_timeout_test.go` are same-package, compile unchanged); import blocks redistribute
(`crypto/tls`, `crypto/x509`, `pem` → mostly `network.go`; `runtime`, `filepath` → mostly
`platform.go`). One commit, build + full suite green ⇒ ER-3 satisfied before any feature line is
written. `src/config/models.go` measures **72 lines** (not 1,085 as the feature plan's pre-seeded
table claims — measured, reported upward); #420 does not touch it, so no split is needed there.

### D-7 — Error semantics: sentinels + setting-naming wraps; validation at resolve time

`FromApplicationConfig` normalizes (`strings.ToLower(strings.TrimSpace(...))`), then:

1. Unknown `DB_TYPE` → `fmt.Errorf("invalid DB_TYPE %q (config key db_type): %w", v, ErrUnsupportedDatabaseType)`
   — the sentinel (#418 `src/database/errors.go`) already names the allowed values in its message;
   wrapping with `%w` keeps `errors.Is` working (`fmt`'s wrapError implements `Unwrap`), while the
   prefix names the **offending setting** exactly as the AC demands. Echoes the user's own value
   (capped at 64 bytes via `strconv.Quote` on a truncated copy — a value the user set is echoed back
   to *them*, in their logs; capped so a 10 KB typo can't spam).
2. Garbage `DB_AUTO_MIGRATE` → `errors.NewWithCode` (`src/errors/main.go:114`, repo idiom per ER-1;
   400 code like the other DB config sentinels): message names `DB_AUTO_MIGRATE` *and* `db_auto_migrate`
   *and* lists accepted spellings.
3. `Type=postgres` + blank DSN → `fmt.Errorf("%w — set DB_DSN (env) or db_dsn (config key db_dsn)", ErrPostgresDSNRequired)`
   — sentinel reused per brief guidance; **the DSN value is never echoed** (there is none, by definition
   of this branch).
4. Final `cfg.Validate()` pass (#418) is returned unwrapped — it is already sentinel-backed and
   factory-side checks stay single-sourced.

All failure surfaces at startup through #418 task 05's `mustInitGormConnection`
(`src/serviceprovider/database_connection.go`): `ctx.LogErrorf("SQL database unavailable: %v", err)`
then `os.Exit(1)` **only** under `API_COMMAND` — the exit policy is *not* duplicated in #420.

### D-8 — Secrets safety: never route DSN through `SetKey`; message-level redaction; audit is negative (verified)

AC "DSN values are never logged in plain text" + brief's "reuse `RedactDSN` from #418". Audit result
(this plan, verified by grep):

- `Config.GetKey` reads env/file values; **#420 code never calls `SetKey`/`Setenv`** with a DSN — so
  the config-file writers (`Config.Save`, `src/config/main.go:170-188` marshals `c.config`;
  `GetConfigContent`, main.go:1004-1011; `install.saveConfig` `src/install/main.go:683-691` writing
  0644; installer's env-copy loop `src/install/main.go:181-183`) never gain a runtime DSN they don't
  already have from *the user's own file* (a file is expected secret storage, not a log). Task 02
  pins this with an assertion that `FromApplicationConfig` mutates neither config state nor
  `os.Environ()`.
- No controller/diagnostic dumps `ConfigFile`/`Environment` over the API or into logs: grep for
  `GetConfig().Environment` hits only `src/install/main.go:625` (key enumeration, no values) — no
  masking change needed anywhere. #418's `RedactDSN` (`src/database/redact.go`, plan D-8) remains
  the only DSN→text funnel inside `database`'s own error paths.
- Regression tests (task 03): table tests across both DSN forms asserting no error string returned
  by `FromApplicationConfig` (or by `Validate`) contains the raw DSN/password; plus boot-matrix
  output assertions (task 04) proving the *startup line* itself is clean.

## Target files & layout (cumulative, per task)

```
src/constants/main.go                      modify (task 02)  — 4 new *_ENV_VAR constants (block lines 82–138)
src/config/
├── main.go                                modify (task 01)  — 1,013 → ~370 (pure move OUT only)
├── database.go                            create (task 01), modify (task 02) — moved DB getters + 4 new ones
├── modules.go    network.go  platform.go  create (task 01) — pure moves only
├── database_getters_test.go               create (task 02) — getter + precedence table tests (package config)
src/database/
├── config.go                              modify (task 02) — FromApplicationConfig fill + signature (D-4)
├── from_application_config_test.go        create (task 02) — resolve/defaults matrix (t.Setenv + config.New reset)
├── db_config_validation_test.go           create (task 03) — actionable-message matrix
├── db_secrets_test.go                     create (task 03) — redaction/no-mutation regressions
src/serviceprovider/
├── database_connection.go                 modify (task 04) — DefaultDatabaseConfig() → FromApplicationConfig()
├── database_connection_test.go            modify (task 04) — switch + default-parity tests
src/integration_test/
└── db_boot_matrix_test.go                 create (task 04) — `//go:build integration` exec matrix
```

## Contract sketch (final state after tasks 02+04)

```go
// src/database/config.go  (replaces #418's seam body; signature widened per D-4)
// FromApplicationConfig resolves DB_TYPE/DB_PATH/DB_DSN/DB_AUTO_MIGRATE through
// config.Get() (env > flag > config file, D-1 flat keys), defaults per AD-3,
// and validates. Errors name the offending setting and allowed values (D-7);
// DSN values never appear in messages (D-8).
func FromApplicationConfig() (DatabaseConfig, error) {
    c := config.Get()
    cfg := DefaultDatabaseConfig()          // baseline = AD-3 defaults (D-5: no duplication)
    if v := normalize(c.DatabaseType()); v != "" {
        cfg.Type = v
    }
    if v := strings.TrimSpace(c.DatabasePath()); v != "" {
        cfg.Path = v
    }
    cfg.DSN = strings.TrimSpace(c.DatabaseDSN())
    auto, err := c.DatabaseAutoMigrate()    // (bool, error), D-3
    if err != nil {
        return DatabaseConfig{}, err
    }
    cfg.AutoMigrate = auto
    if !isAllowedType(cfg.Type) {           // {sqlite, postgres}, case-insensitive
        return DatabaseConfig{}, fmt.Errorf("invalid DB_TYPE %s (config key db_type): %w", quoteCapped(cfg.Type), ErrUnsupportedDatabaseType)
    }
    if cfg.Type == DBTypePostgres && cfg.DSN == "" {
        return DatabaseConfig{}, fmt.Errorf("%w — set DB_DSN (env) or db_dsn (config key db_dsn)", ErrPostgresDSNRequired)
    }
    if err := cfg.Validate(); err != nil {  // #418's checks remain the single source
        return DatabaseConfig{}, err
    }
    return cfg, nil
}
```

`src/serviceprovider/database_connection.go` `initGormConnection` becomes:

```go
cfg, err := database.FromApplicationConfig()   // was: database.DefaultDatabaseConfig() — the #418 seam comment
if err != nil { s.SQLDatabase = nil; return err }  // mustInitGormConnection: LogErrorf + exit(1) iff API_COMMAND (policy already merged in #418 t05)
```

## Task index (each task = one PR, merge in order)

| Task | Goal | Depends on | Touches | PR size |
|---|---|---|---|---|
| [01](./tasks/01-split-config-main-pure-move.md) | ER-3: pure-move split of `config/main.go` (1,013 → ~370) — unlocks all later edits | — (plus #418 merged) | `src/config/{main,database,modules,network,platform}.go` | S (mechanical, wide diff) |
| [02](./tasks/02-db-setting-getters-config-bridge.md) | Constants + `*Config` getters + `FromApplicationConfig` filled (D-1…D-5, D-7) + unit tests | 01 | `constants/main.go`, `config/database.go`, `config/database_getters_test.go`, `database/config.go`, `database/from_application_config_test.go` | M |
| [03](./tasks/03-validation-ux-secrets-safety.md) | AC "Validation and errors" + "Secrets safety" hardened: message matrix, redaction regressions, no-mutation guard; message polish only if a test fails | 02 | `database/db_config_validation_test.go`, `database/db_secrets_test.go` (+ surgical fixes in `database/config.go` if needed) | S |
| [04](./tasks/04-provider-switch-default-parity.md) | Flip provider to `FromApplicationConfig()`; default-parity tests; integration boot matrix; deviation notes for PR/issue/#427 | 03 (runtime chain) + **#418 task 05 merged** (provider file exists) | `serviceprovider/database_connection.go`, its test file, `integration_test/db_boot_matrix_test.go` | S–M |

**Sequencing/ER-5 notes.** Partial completion never breaks main: after 01 the repo is unchanged
behaviorally; after 02 the seam is still uncalled (`DefaultDatabaseConfig()` still wired), so
runtime behavior is still byte-identical; after 03 ditto (tests + maybe message polish only);
only 04 changes runtime, and its default-parity test proves no-config boots identically. File
overlap is always behind a declared dependency (01→02 on `config/database.go`; 02→03→04 on
`database/config.go` consumers). No two *unfinished* tasks touch the same file.

## AC coverage map (every [brief](./brief.md) AC → tasks)

| AC checkbox | Task(s) | Where proven |
|---|---|---|
| Backend selection (`DB_TYPE`, default `sqlite`) | 02, 04 | `from_application_config_test.go` (unset→sqlite, case-insensitive, unknown→err); boot matrix (04) |
| SQLite path (`DB_PATH` + file key; AD-3 default) | 02, 04 | `database_getters_test.go`, `from_application_config_test.go` (explicit / `DATABASE_FOLDER` / HOME fallback); boot matrix |
| PostgreSQL DSN (`DB_DSN` + file key; required + validated at startup) | 02, 03, 04 | D-7.3 error + matrix tests (03); startup exit-1 with actionable message (04 boot matrix, policy from #418 t05) |
| Auto-migrate flag (`DB_AUTO_MIGRATE`, default `true`) | 02, 03 | tolerant-bool table + garbage→error (02 tests; messages hardened in 03) |
| Config file parity + env precedence | 02 | `database_getters_test.go` sets `cfg.config.Environment` (pattern `src/config/main_test.go:31-40`) + `t.Setenv` and asserts env>file per key; mechanism is `GetKey` itself (main.go:933) |
| Validation and errors (clear, actionable, at startup) | 02, 03, 04 | D-7 wraps name the setting + allowed values; message matrix (03); exits via #418's policy — exit-code evidence in 04's matrix |
| Secrets safety (DSN never logged plain) | 03, 04 | audit negative-result (this plan §D-8) + redaction regressions (03) + clean startup-line assertions incl. `sup3rs3cr3t`-canary check (04) |

## Risks specific to #420

| # | Risk | Mitigation |
|---|---|---|
| R-420-1 | A user copies the issue table literally (`database:` YAML section) → whole config file fails to parse (`map[string]string`, main.go:147-160), confusing UX | Deviation documented + PR/issue note (D-1 protocol); handover to #427 with explicit "flat keys only" example; OQ-420-3 records a possible friendly loader diagnostic as follow-up |
| R-420-2 | Task 01's split conflicts with parallel work on `config/main.go` by other features | ER-3 makes the split mandatory *before* #420 edits; land 01 first, announce rebase window in PR; pure-move conflicts are trivial |
| R-420-3 | Seam signature widening (D-4) surprises #421 (planned in parallel) which may have assumed `FromApplicationConfig() DatabaseConfig` | Documented here + in task 02 PR note; #421 consumes the seam *after* #420 merges (feature plan §8 order 420→421); parity with `Validate()` unchanged keeps factory-side checks intact |
| R-420-4 | Flip (task 04) changes boot behavior on misconfigured installs that silently worked before (bad `DB_TYPE` typo today is ignored — nobody read it; after #420 it exits) | That *is* the AC (validated on startup); the exit only happens under `API_COMMAND` (#418 D-7), CLI keeps working to repair config; release note goes to #427 |
| R-420-5 | `t.Setenv` + global `config` singleton leak across tests | Follow `src/data/main_test.go:16-36` reset idiom (`config.New(ctx)` + `ctx.DisableLog()`), never touch global env outside `t.Setenv`; `t.Parallel` forbidden in these files (documented in test headers) |
| R-420-6 | Boot matrix needs a built binary + real exits — flaky under CI runners | Kept behind `//go:build integration` (repo pattern: `src/integration_test/*.go`) + manual PR-step matrix following #418 task 05's precedent; default `go test ./...` never execs binaries |

## Open questions (human in the loop; none block tasks 01–02)

| ID | Question | Proposal |
|---|---|---|
| OQ-420-1 | Adopt flat keys `db_type`/`db_path`/`db_dsn`/`db_auto_migrate` — deviation from the issue's `database.*` notation (D-1). Confirm before task 02's PR is posted (issue comment) | Adopt flat keys; comment on #420 mirroring AD-2's protocol |
| OQ-420-2 | Expose `DB_PG_*` pool overrides now (offer from #418 handover)? | **No** in #420 — no AC requires them; add later under #421/#425 only if operationally requested |
| OQ-420-3 | Should `config.Load()` gain a friendly "nested `database:` section unsupported" diagnostic (behavior change in shared loader)? | Out of #420 scope (ER-1); file as follow-up idea with #427's doc note as the stopgap |
| OQ-420-4 | `DB_PATH` vs `DATABASE_FOLDER` precedence when both set (legacy vs new) | Proposed: `DB_PATH` wins (explicit newer setting beats generic folder); implemented that way in task 02, flagged for docs (#427) |
| OQ-420-5 | Feature plan's pre-seeded table lists `src/config/models.go` at 1,085 lines; measured 72 (2026-11-03). Table is stale (only affects other PBIs' guards) | Report upward at feature-plan touch; #420 does not touch models.go, no action here |

## Verification commands (all tasks)

```bash
cd src
go build ./...                        # green at every merged boundary (ER-5)
go vet ./config/... ./database/... ./serviceprovider/...
go test ./config/... ./database/... -count=1 -v
go test ./... -count=1                # full suite, no regressions
go test -tags integration ./integration_test/... -count=1   # task 04 matrix (needs built binary; see task)
```

## Execution model — worktree slices + human gate (feature plan §14, ER-6/ER-7)

Every task here executes as an isolated worktree slice `slice/<pbi>-<nn>` cut from **`sql-migration`**,
passes its spec's Validation inside the worktree, closes with a **test plan + results entry** in
[`../VERIFICATION.md`](../VERIFICATION.md) and the **human verdict** — only then merges `--no-ff` into
`sql-migration` and the next slice may start. ER-5's merge boundary = the merge into `sql-migration`.
