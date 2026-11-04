# Task 02 — `DatabaseConfig` contract, defaults and validation

**PBI** [#418](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 01 · **PR size:** S
**Plan refs:** [D-3](./plan.md#d-3--databaseconfig-is-a-neutral-struct-not-configconfig), AD-3 (feature)

## Context

The connection factory (tasks 03/04) must not read env vars itself — otherwise config semantics
would splinter across packages before #420 exists. This task defines the neutral contract struct,
its AD-3 default resolution, and pure validation. #420 later fills the declared
`FromApplicationConfig` seam; nothing else changes for it.

## Objective

`database.DatabaseConfig` exists with typed defaults + `Validate()`, fully unit-tested, exported
sentinel errors, and the #420 seam declared. Compiles and tests green; **still no connection code**.

## Files

| File | Op |
|---|---|
| `src/database/errors.go` | create |
| `src/database/config.go` | create |
| `src/database/config_test.go` | create |

## File size guard

All three files are new; expected sizes `errors.go` ~15, `config.go` ~120, `config_test.go` ~180.
Verdict: **ok**.

## Implementation spec

### `src/database/errors.go`

```go
package database

import "github.com/Parallels/prl-devops-service/errors"

var (
    ErrUnsupportedDatabaseType = errors.NewWithCode("unsupported database type, expected 'sqlite' or 'postgres'", 400)
    ErrPostgresDSNRequired     = errors.NewWithCode("database type 'postgres' requires a connection string (DSN)", 400)
    ErrSQLitePathRequired      = errors.NewWithCode("database type 'sqlite' requires a file path", 400)
    ErrDatabaseUnreachable     = errors.NewWithCode("database is unreachable", 503)
)
```

(`errors.NewWithCode` = `src/errors` — the repo idiom, http-code carries through to API errors in
#421. Match the real signature in `src/errors` before coding.)

### `src/database/config.go`

```go
package database

const (
    DBTypeSQLite   = "sqlite"
    DBTypePostgres = "postgres"
    // DefaultDatabaseFilename preserves the legacy *folder* convention (AD-3)
    // with a distinct filename so data.json and prl-devops.db coexist.
    DefaultDatabaseFilename = "prl-devops.db"
    MemoryPath              = ":memory:"
)

type PostgresPoolConfig struct {
    MaxOpenConns    int           // 0 → DefaultPostgresMaxOpenConns
    MaxIdleConns    int           // 0 → DefaultPostgresMaxIdleConns
    ConnMaxLifetime time.Duration // 0 → DefaultPostgresConnMaxLifetime
    ConnMaxIdleTime time.Duration // 0 → leave driver default
}

type DatabaseConfig struct {
    Type        string
    Path        string
    DSN         string
    AutoMigrate bool // stored for #421/#424; never read in #418
    Postgres    PostgresPoolConfig
}

const (
    DefaultPostgresMaxOpenConns    = 25
    DefaultPostgresMaxIdleConns    = 5
    DefaultPostgresConnMaxLifetime = 30 * time.Minute
)
```

Behaviour (each unit-tested):

- `DefaultDatabaseConfig() DatabaseConfig` — `Type: DBTypeSQLite`, `AutoMigrate: true`, `Path` =
  `defaultDatabasePath()`, empty DSN, zero pool (defaults applied at use site, task 04).
- `defaultDatabasePath() string`:
  1. `config.Get().DatabaseFolder()` non-empty → `filepath.Join(folder, DefaultDatabaseFilename)`
  2. else `os.Getuid() == 0` → `"/etc/parallels-devops/" + name`
  3. else `os.UserHomeDir()` succeeds → `filepath.Join(home, ".parallels-devops", name)`
  4. home lookup failure → fall back to case 2 path and `ctx`-free `fmt`-level warning is NOT
     allowed here (no ctx) — return the path; callers log. Mirror the comment in
     `src/serviceprovider/main.go:69-104` so the duplication is *documented* (AD-3; #420 may
     centralise).
- `(DatabaseConfig) Validate() error`:
  - `Type == ""` → treated as `DBTypeSQLite` before checks (defaults-forward-compatible);
  - `DBTypePostgres` + trimmed-empty DSN → `ErrPostgresDSNRequired`;
  - `DBTypeSQLite` + trimmed-empty Path → `ErrSQLitePathRequired`;
  - other types → `ErrUnsupportedDatabaseType` (message quotes allowed values; AC #420
    “actionable error” seed);
  - type comparison case-insensitive (`strings.EqualFold`), stored normalised lowercase by callers.
- Pool getters `(PostgresPoolConfig) MaxOpen()/MaxIdle()/Lifetime()/IdleTime()` returning
  defaults when zero — task 04 consumes them; tests assert both branches.
- **The seam:**
  ```go
  // FromApplicationConfig builds the config from the application configuration
  // (DB_TYPE/DB_PATH/DB_DSN/DB_AUTO_MIGRATE). PBI #420 owns that implementation;
  // until then it MUST return DefaultDatabaseConfig() unchanged.
  func FromApplicationConfig() DatabaseConfig { return DefaultDatabaseConfig() }
  ```

### `src/database/config_test.go` — matrix

| Case | Expect |
|---|---|
| `DefaultDatabaseConfig()` with `t.Setenv("DATABASE_FOLDER", tmp)` | Path == `tmp/prl-devops.db`, Type sqlite, AutoMigrate true |
| `DATABASE_FOLDER` unset, HOME forced (`t.Setenv("HOME", tmp)` — unix CI) | Path under `tmp/.parallels-devops/prl-devops.db` |
| Validate: empty type | nil (defaults to sqlite) |
| Validate: `"SQLITE"` type | nil (case-insensitive) |
| Validate: `"mysql"` | `ErrUnsupportedDatabaseType` |
| Validate: postgres, DSN `" "` | `ErrPostgresDSNRequired` |
| Validate: sqlite, Path `"  "` | `ErrSQLitePathRequired` |
| Pool getters zero vs set | defaults vs given values |

Use `basecontext.NewBaseContext()` + `ctx.DisableLog()` for anything needing a ctx (pattern:
`src/data/main_test.go:16-36`); config global: `config.New(ctx)` where needed — same pattern.

## Out of scope

No GORM import beyond what 01 pinned. No env parsing for `DB_*`. No YAML `database:` section
(#420). No AutoMigrate logic anywhere.

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./database/... -count=1 -v` — matrix above all green
- [ ] `cd src && go vet ./database/...`
- [ ] `git diff` shows no changes outside `src/database/`
- [ ] Manual: `grep -n "TODO(#420)\|MUST return" src/database/config.go` — seam documented at `FromApplicationConfig`

## Rollback

Revert commit; package still just `doc.go` — no consumers.

## PBI acceptance mapping

Foundation for AC “Connection factory” (config input shape) and, later, #420's ACs through the seam.
