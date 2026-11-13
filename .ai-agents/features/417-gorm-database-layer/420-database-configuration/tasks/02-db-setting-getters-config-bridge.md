# Task 02 — `DB_*` constants, typed config getters, and the `FromApplicationConfig` bridge

**PBI** [#420](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 01 (`config/database.go` must exist), #418 tasks 01–02 merged (`src/database/config.go` seam + sentinels) · **PR size:** M
**Plan refs:** [D-1](./plan.md#d-1--config-file-keys-are-flat) · [D-2](./plan.md#d-2--env-varconstant-names) · [D-3](./plan.md#d-3--typed-getters-on-config-dumb-by-design) · [D-4](./plan.md#d-4--seam-signature-grows-an-error) · [D-5](./plan.md#d-5--defaults-delegate-to-418) · [D-7](./plan.md#d-7--error-semantics)

## Context

`Config.GetKey` (`src/config/main.go:933` pre-split; post-task-01 still `config` package core)
already resolves *file key → CLI flag → env var, env wins* — the "config file parity + env
precedence" AC is a property of the mechanism. What's missing is vocabulary: named settings, typed
readers, and the bridge that turns them into the `DatabaseConfig` struct #418 defined. #418 left
`FromApplicationConfig()` in `src/database/config.go` returning `DefaultDatabaseConfig()` with a
`TODO(#420)` marker; this task fills it. The seam has **no callers yet** (the provider still calls
`DefaultDatabaseConfig()`), so this PR changes zero runtime behavior — the flip is task 04.

## Objective

`database.FromApplicationConfig() (DatabaseConfig, error)` resolves `DB_TYPE`/`DB_PATH`/`DB_DSN`/
`DB_AUTO_MIGRATE` (env, CLI flag, or flat config-file key) with defaults byte-identical to
`DefaultDatabaseConfig()`, strict-but-tolerant parsing, and actionable errors — fully unit-tested;
`go build`/`go test` green; **no call-site changes** (provider flip = task 04).

## Files

| File | Op |
|---|---|
| `src/constants/main.go` | modify — 4 new constants in the env-var block |
| `src/config/database.go` | modify — 4 new getters appended (task 01 created the file) |
| `src/config/database_getters_test.go` | create — getter + precedence table tests (package `config`) |
| `src/database/config.go` | modify — fill `FromApplicationConfig` (replace seam body; widen signature per D-4) |
| `src/database/from_application_config_test.go` | create — resolve/defaults matrix (package `database`) |

## File size guard

| File | wc -l at plan time (measured / predicted) | After | Verdict |
|---|---|---|---|
| `src/constants/main.go` | **560** (measured 2026-11-03) | ~565 | ok |
| `src/config/database.go` | ~70 (created by task 01) | ~140 | ok |
| `src/database/config.go` | ~120 per #418 task 02 estimate — **re-run `wc -l` after #418 merge** | ~200 | ok |
| both test files | new | < 260 each | ok |
| `src/config/main.go` | ~370 after task 01 | 370 (untouched here) | ok |

## Implementation spec

### 1. Constants — `src/constants/main.go`

Env-var const block spans lines 82–138 (verified; `DATABASE_FOLDER_ENV_VAR` at :101, `…_ENV_VAR`
naming pattern, **zero existing `DB_*` constants** — no collision). Append inside the block after
:104 (`DATABASE_SAVE_INTERVAL_ENV_VAR`), matching gofmt alignment:

```go
	DB_TYPE_ENV_VAR         = "DB_TYPE"
	DB_PATH_ENV_VAR         = "DB_PATH"
	DB_DSN_ENV_VAR          = "DB_DSN"
	DB_AUTO_MIGRATE_ENV_VAR = "DB_AUTO_MIGRATE"
```

Flat config-file keys need **no constants** — `GetKey` matches file keys case-insensitively against
the same name string (D-1); the env-var constant doubles as the file key (exactly how
`database_folder` works today, `src/config.local.yaml:5` ↔ `DATABASE_FOLDER_ENV_VAR`).

### 2. Getters — `src/config/database.go` (dumb by design, D-3)

One-liner idiom of `DatabaseFolder()` (`src/config/main.go:598`), plus the `(value, error)` getter
precedent of `ValidateTlsConfiguration() (bool, string)` (`src/config/main.go:338`):

```go
// DatabaseType returns the raw DB_TYPE setting ("" = unset → caller defaults to sqlite).
// Config file key db_type resolves through the same case-insensitive GetKey lookup.
func (c *Config) DatabaseType() string { return c.GetKey(constants.DB_TYPE_ENV_VAR) }

// DatabasePath returns the raw DB_PATH setting ("" = unset → AD-3 default resolution).
func (c *Config) DatabasePath() string { return c.GetKey(constants.DB_PATH_ENV_VAR) }

// DatabaseDSN returns the raw DB_DSN setting ("" = unset). Treat as a secret:
// callers must never embed it in error/log text unredacted (#420 plan D-8).
func (c *Config) DatabaseDSN() string { return c.GetKey(constants.DB_DSN_ENV_VAR) }

// DatabaseAutoMigrate parses DB_AUTO_MIGRATE. Unset → true (AC: "Defaults to true").
// Vocabulary extends the repo's tolerant-bool sets (helpers.StringToBool,
// src/helpers/strings.go:144; IsDatabaseAutoRecover, config main.go:525) with an
// explicit false-set — and, deliberately unlike both, returns a clear error for
// garbage instead of silently defaulting (AC "Validation and errors").
func (c *Config) DatabaseAutoMigrate() (bool, error) {
	raw := strings.ToLower(strings.TrimSpace(c.GetKey(constants.DB_AUTO_MIGRATE_ENV_VAR)))
	switch raw {
	case "":
		return true, nil
	case "true", "1", "t", "yes", "y", "on", "enabled":
		return true, nil
	case "false", "0", "f", "no", "n", "off", "disabled":
		return false, nil
	default:
		// Echo capped at 64 bytes (same guard idea as database.quoteCapped, D-7.1): a
		// mis-set value is echoed to its owner, but a paste-gone-wrong can't spam logs.
		echo := raw
		if len(echo) > 64 {
			echo = echo[:64] + "…"
		}
		return false, errors.NewWithCode(
			fmt.Sprintf("invalid DB_AUTO_MIGRATE %q (config key db_auto_migrate): expected true or false", echo), 400)
	}
}
```

`config/database.go` gains imports `fmt`, `strings` + `github.com/Parallels/prl-devops-service/errors`
(`errors.NewWithCode` = `src/errors/main.go:114`; `src/errors` imports only `fmt`/`strings` → no
cycle, checked against `src/errors/main.go:3-6`).

### 3. Bridge — `src/database/config.go` (replace seam body; D-4 signature)

```go
// FromApplicationConfig builds DatabaseConfig from DB_TYPE/DB_PATH/DB_DSN/DB_AUTO_MIGRATE
// resolved through config.Get() — env var > CLI flag > config-file key (Config.GetKey),
// config-file keys flat lowercase (db_type, db_path, db_dsn, db_auto_migrate; #420 D-1).
// Unset values keep DefaultDatabaseConfig()'s AD-3 defaults byte-identically (D-5).
// Errors name the offending setting + allowed values; the DSN never appears in messages (D-8).
func FromApplicationConfig() (DatabaseConfig, error) {
	c := config.Get()
	cfg := DefaultDatabaseConfig()

	if t := strings.ToLower(strings.TrimSpace(c.DatabaseType())); t != "" {
		cfg.Type = t
	}
	if p := strings.TrimSpace(c.DatabasePath()); p != "" {
		cfg.Path = p // DB_PATH beats DATABASE_FOLDER-derived default (OQ-420-4)
	}
	cfg.DSN = strings.TrimSpace(c.DatabaseDSN())

	autoMigrate, err := c.DatabaseAutoMigrate()
	if err != nil {
		return DatabaseConfig{}, err
	}
	cfg.AutoMigrate = autoMigrate

	switch strings.ToLower(cfg.Type) {
	case DBTypeSQLite, DBTypePostgres: // constants from #418 task 02, same file/package
	default:
		return DatabaseConfig{}, fmt.Errorf("invalid DB_TYPE %s (config key db_type): %w",
			quoteCapped(cfg.Type, 64), ErrUnsupportedDatabaseType) // sentinel text names sqlite|postgres
	}
	if strings.EqualFold(cfg.Type, DBTypePostgres) && strings.TrimSpace(cfg.DSN) == "" {
		return DatabaseConfig{}, fmt.Errorf("%w — set DB_DSN (env) or db_dsn (config key db_dsn)",
			ErrPostgresDSNRequired)
	}
	if err := cfg.Validate(); err != nil { // #418 checks stay the single source of truth
		return DatabaseConfig{}, err
	}
	return cfg, nil
}

// quoteCapped quotes v for an error message, truncated at n bytes so a huge typo
// can't spam logs. Echoing a user's own setting value back to them is safe; the
// DSN is never routed through this (D-8).
func quoteCapped(v string, n int) string {
	if len(v) > n {
		v = v[:n] + "…"
	}
	return strconv.Quote(v)
}
```

Notes: `%w` keeps `errors.Is(err, ErrUnsupportedDatabaseType)` working (`fmt.wrapError` implements
`Unwrap`); `config.go` imports grow `strconv` (`fmt`, `strings`, `config` already imported by #418).
`AutoMigrate` flag is only **stored** here — AutoMigrate *execution* is #419/#421 (AD-4); never act
on it in this PBI.

**Required PR-description note (D-1 deviation protocol):** state that config-file keys are flat
(`db_type`/`db_path`/`db_dsn`/`db_auto_migrate`), not the issue table's `database.type` dotted form,
with the one-line reason (loader shape `ConfigFile.Environment map[string]string`,
`src/config/models.go:9` + `GetKey` EqualFold scan) — and post the matching comment on issue #420.

### 4. Tests

`src/config/database_getters_test.go` (package `config`) — file-key + env matrix, direct
`cfg.config.Environment` seeding per the existing pattern (`src/config/main_test.go:31-40`), env
override via `t.Setenv` (auto-restores; never `os.Setenv`):

| Case | Expect |
|---|---|
| `db_type: "postgres"` file key only | `DatabaseType() == "postgres"` (AC: config-file parity) |
| `db_type: postgres` + `t.Setenv("DB_TYPE","sqlite")` | `"sqlite"` (AC: env precedence — env wins in `GetKey`, main.go:949-951) |
| key spelled `DB_TYPE` in file map | still matches (`EqualFold`, main.go:935) |
| auto_migrate unset / `"yes"` / `"OFF"` / `"maybe"` | `true,nil` / `true,nil` / `false,nil` / error message contains `DB_AUTO_MIGRATE` and `db_auto_migrate` |

`src/database/from_application_config_test.go` (package `database`) — reset idiom per
`src/data/main_test.go:16-36`: `ctx := basecontext.NewBaseContext(); ctx.DisableLog(); config.New(ctx)`.
Defaults touch `$HOME` via AD-3 fallback — always pin `DATABASE_FOLDER` (and `HOME` where the
fallback branch is the subject, unix CI only, mirroring #418 task 02's matrix):

| Case | Expect |
|---|---|
| nothing set (clean `DATABASE_FOLDER=tmp`) | `cfg == DefaultDatabaseConfig()` (field-wise; D-5 byte-identity) — proves AC "identical behaviour … for users who make no changes" at the contract level |
| `t.Setenv DB_TYPE sqlite/postgres/SQLITE/MySQL/""` | sqlite ok; postgres w/o DSN → `ErrPostgresDSNRequired` via `errors.Is`; case-insensitive; `MySQL` → `ErrUnsupportedDatabaseType` + message contains `DB_TYPE` and `db_type`; `""` → sqlite default |
| `t.Setenv DB_PATH tmp/x.db` | `cfg.Path == tmp/x.db` even with `DATABASE_FOLDER` also set (OQ-420-4 precedence) |
| file-key path: seed `config.Get().` `SetKey("db_path", …)` (writes into the in-memory Environment map — the file-key code path) + no env | resolves — proves flat-key resolution end-to-end through `GetKey` |
| `DB_AUTO_MIGRATE=false` / garbage | `false` / error with setting name (error text asserted, AC 6 seed) |
| postgres + `DB_DSN=" "` (whitespace) | `ErrPostgresDSNRequired` (trimmed) |
| postgres + `DB_DSN=postgres://u:sup3rs3cr3t@h/db` | no error; `cfg.DSN` verbatim (trimmed only) |

All tests: no `t.Parallel` (shared `globalConfig` — R-420-5), no disk writes outside `t.TempDir()`.

## Out of scope

Provider call-site flip (task 04). AutoMigrate execution / schema (AD-4: #419/#421). Nested
`database.*` YAML support (D-1). `DB_PG_*` pool overrides (OQ-420-2). Redaction test suite
(task 03). Any edit to `config/main.go`, `models.go`, legacy `src/data`.

## Validation

- [ ] `cd src && go build ./...` — seam signature change compiles (no callers exist yet; `grep -rn "FromApplicationConfig(" src --include='*.go'` shows decl + tests only)
- [ ] `cd src && go test ./config/... ./database/... -count=1 -v` — both matrices green
- [ ] `cd src && go test ./... -count=1` — repo-wide green (ER-5)
- [ ] `cd src && go vet ./config/... ./database/...`
- [ ] `git diff` touches exactly the 5 files listed
- [ ] Manual: grep the new test files for `os.Setenv` → none (only `t.Setenv`)
- [ ] PR description contains the D-1 flat-key deviation note; issue #420 comment posted (D-1 protocol)

## Rollback

Revert commit: seam body returns to `return DefaultDatabaseConfig(), nil`-equivalent — *if* a
follow-up needs to roll back just the fill, re-point the body to `DefaultDatabaseConfig()` (task 04
depends on the error-returning signature; rolling back task 02 after 04 requires rolling back 04
first — order stated for the revert-PR author). No runtime behavior existed before or after this PR
alone (seam unconsumed), so rollback is user-invisible.

## PBI acceptance mapping

Moves AC boxes: **Backend selection** (full resolution + default, contract-level), **SQLite path**
(D-5/AD-3 delegation + `DB_PATH` precedence), **PostgreSQL DSN** (read + required-when-postgres
error; startup exit proof lands in task 04), **Auto-migrate flag** (parse + default true + garbage
error), **Config file parity** (parity + env precedence proven per key). Contributes AC "Validation
and errors" (message authoring, finalized in task 03).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `420-02`.** Cut `slice/420-02` from `sql-migration` HEAD
(`.worktrees/slice-420-02`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
