# PBI Brief #420 — Database configuration via environment variables and config file

| Field | Value |
|---|---|
| PBI | [#420](https://github.com/Parallels/prl-devops-service/issues/420) — `[PBI] Database configuration via environment variables and config file` |
| Labels | `triage`, `pbi` · Assignee: unassigned |
| Round | Deep-dive round **2** (order rationale: [feature plan §8](../plan.md#8-delivery-order--dependency-graph)) |
| Upstream feature docs | [`brief.md`](../brief.md) · [`plan.md`](../plan.md) — both required reading |
| Planning status | **Fully planned** — [plan.md](./plan.md) · tasks [01](./tasks/01-split-config-main-pure-move.md) · [02](./tasks/02-db-setting-getters-config-bridge.md) · [03](./tasks/03-validation-ux-secrets-safety.md) · [04](./tasks/04-provider-switch-default-parity.md) |

## Description (verbatim from GitHub)

Implement the configuration layer that allows users to select the database backend and supply
connection parameters through either environment variables or the application config file. SQLite
must be the default with zero additional configuration required.

### User Story

As a **service operator**, I want to configure the database backend and its connection settings
using environment variables or a config file, so that I can deploy the service in any environment —
from a local laptop to a production Kubernetes cluster — without modifying application code.

### Acceptance Criteria

- [ ] **Backend selection**: `DB_TYPE` env var (values: `sqlite`, `postgres`) selects the backend. Defaults to `sqlite` if unset.
- [ ] **SQLite path**: `DB_PATH` env var (and `database.path` config key) sets the SQLite file path. Defaults to the current legacy database file path.
- [ ] **PostgreSQL DSN**: `DB_DSN` env var (and `database.dsn` config key) supplies the PostgreSQL connection string. Required when `DB_TYPE=postgres`; validated on startup.
- [ ] **Auto-migrate flag**: `DB_AUTO_MIGRATE` env var (and `database.auto_migrate` config key) controls whether schema migration runs on startup. Defaults to `true`.
- [ ] **Config file parity**: All env vars have equivalent config file keys; env vars take precedence over config file values.
- [ ] **Validation and errors**: Missing or invalid configuration produces a clear, actionable error message at startup.
- [ ] **Secrets safety**: DSN values are never logged in plain text.

### Definition of Done

- [x] Code implemented following best practices.
- [x] Unit tests written and passing.
- [x] Code reviewed and approved.
- [x] Merged into the main branch.
- [x] Documentation updated (if applicable).
- [x] Deployed to staging/production environment.

### Assumptions and Constraints

- **Assumption**: The project already has a configuration loading mechanism (e.g. Viper or similar); this PBI extends it rather than replacing it. ✅ **True** — see Repo context.
- **Constraint**: The default configuration must produce identical behaviour to the current application for users who make no changes.

## Repo context (verified 2026-11-03)

- The loader is **custom, not Viper**: `Config.GetKey(key)` at `src/config/main.go:933` resolves
  config-file value → CLI flag → `os.LookupEnv` (**env wins**) — the "config file parity + env
  precedence" AC is already satisfied by the mechanism itself; this PBI mostly adds typed getters.
  Config file formats: YAML (`.local.yaml`/`.yaml` discovery + `--config-file` flag), loaded in
  `src/config/main.go`.
- Typed getters follow an established one-liner pattern, e.g.
  `func (c *Config) DatabaseFolder() string { return c.GetKey(constants.DATABASE_FOLDER_ENV_VAR) }`
  (`src/config/main.go:598`). Boolean/int reads: `GetIntKey` and siblings.
- Env-var names are constants in `src/constants/main.go` (existing DB set at lines ~101-122).
  Add `DB_TYPE`/`DB_PATH`/`DB_DSN`/`DB_AUTO_MIGRATE` constants there.
- **Output target**: #418 task 02 creates `DatabaseConfig` in `src/database/config.go` with a
  `FromLegacyConfig()`-style bridge; this PBI **fills that bridge** (reads `DB_*` keys, falls back
  to legacy conventions) — single source of truth, no second resolver.
- DSN secret safety: reuse `RedactDSN` from #418 (`src/database/redact.go`); never log raw DSN.
- Defaults interplay: `DB_PATH` unset ⇒ AD-3 (`<legacy folder>/prl-devops.db`, folder honouring
  `DATABASE_FOLDER`); "Defaults to the current legacy database file path" in the AC means the
  legacy **folder** convention — the file itself gets the new name (AD-3, human-approved).

## Engineering-rule triggers (feature plan §6)

- **ER-3 REFACTOR-FIRST**: `src/config/main.go` is **1,013 lines** — any edit here requires the
  split task first (suggested cohesive split: `database.go`, `modules.go`, `logging.go`).
  `src/config/models.go` is 1,085 — same rule if touched.
- ER-1: getters on `*Config`, constants file, `GetKey` reuse — no new config frameworks.

## Fit in the feature

- **Depends on**: #418 (`DatabaseConfig` contract + redaction helper exist).
- **Feeds**: #419 (AutoMigrate path), #421 (backend selection at serviceprovider switch), #424
  (`auto_migrate` flag), #427 (docs).
- Sequencing note: merge **before** #419's AutoMigrate validation and #421's selection flip.
