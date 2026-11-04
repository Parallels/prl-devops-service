# PBI Brief #421 — SQL backend implementation

| Field | Value |
|---|---|
| PBI | [#421](https://github.com/Parallels/prl-devops-service/issues/421) — `[PBI] SQL backend implementation` |
| Labels | `triage`, `pbi` · Assignee: unassigned |
| Round | Deep-dive round **4** — **the largest PBI**; expect a task chain of 10+ PRs |
| Upstream feature docs | [`brief.md`](../brief.md) · [`plan.md`](../plan.md) — required reading, esp. AD-1, AD-2, ER-3 |
| Planning status | **Fully planned** — [plan.md](./plan.md) · chain [01](./tasks/01-interface-capture.md) → [02](./tasks/02-query-language-filter-order-index.md) → [03](./tasks/03-query-language-authorize-harness.md) → [04](./tasks/04-gormdatabase-bootstrap.md) → [05](./tasks/05-domain-roles-claims.md) → [06](./tasks/06-domain-users.md) · [07](./tasks/07-domain-api-keys.md) · [08](./tasks/08-domain-config-schema.md) · [09](./tasks/09-domain-user-configs.md) · [10](./tasks/10-domain-packer-templates.md) · [11](./tasks/11-domain-catalog.md) · [12](./tasks/12-domain-reverse-proxy.md) · [13](./tasks/13-domain-orchestrator-hosts.md) · [14](./tasks/14-domain-orchestrator-vms-snapshots.md) · [15](./tasks/15-domain-jobs.md) · [16](./tasks/16-domain-enrollment-tokens.md) · [17](./tasks/17-call-site-conversion.md) · [18](./tasks/18-final-flip.md) |

## Description (verbatim from GitHub)

Implement the full GORM-based data access layer for the SQLite backend. This replaces all custom
read/write operations currently targeting the legacy store with GORM queries, using SQL as the
engine.

### User Story

As a **user running the service on a single host**, I want the service to use SQL as its database by
default, so that I get the simplicity of a file-based setup without needing to run a separate
database server.

### Acceptance Criteria

- [ ] **All CRUD operations**: Create, read, update, and delete operations for every entity are implemented using GORM and verified against SQLite.
- [ ] **File path respected**: The SQLite database file is created at the path specified in configuration (defaulting to the existing legacy path).
- [ ] **Transactions**: Operations that must be atomic are wrapped in GORM transactions.
- [ ] **Concurrent access**: SQLite WAL mode is enabled to handle concurrent reads without locking issues.
- [ ] **Error handling**: GORM errors are wrapped and surfaced consistently through the existing error handling conventions.
- [ ] **Legacy layer bypassed**: No code paths in scope reach the old data layer when SQLite is the configured backend.

### Definition of Done

- [x] Code implemented following best practices.
- [x] Unit tests written and passing.
- [x] Code reviewed and approved.
- [x] Merged into the main branch.
- [x] Documentation updated (if applicable).
- [x] Deployed to staging/production environment.

### Assumptions and Constraints

- **Assumption**: SQLite 3.x is available in all target runtime environments.
- **Constraint**: The SQLite driver must be the pure-Go implementation (`modernc.org/sqlite`) to avoid CGO requirements → **AD-2: use `github.com/glebarez/sqlite` GORM dialect.**

### Dependencies

Unstated on GitHub; locked by [feature plan §8](../plan.md#8-delivery-order--dependency-graph):
**#418 + #419 + #420 merged.**

## Repo context (verified 2026-11-03)

- **The API surface to implement: 157 methods.** Enumerate at task time with:
  `grep -rn "func (j \*JsonDatabase) [A-Z]" src/data/*.go | grep -v _test | sort`.
  Grouping per domain (from file sizes in `src/data/`): users 422, orchestrator 933, catalog 714,
  catalog_managers, jobs, claims, roles (+role_claims), api_key, config (user/system config),
  user_config, packer_template, reverse_proxy, vm_snapshots, orchestrator_enrollment_token,
  claims, backup, main (connect/save/lock helpers — **not** part of the SQL engine; AD-1's
  `Legacy()` escape hatch covers `SaveAs/SaveNow/Backup` callers: `src/main.go:59`, `src/data/backup.go`).
- **Query-language parity is the hidden iceberg**: `ParseFilter`/`FilterByProperty`
  (`filter_layer.go`), `order_layer.go`, `index_layer.go`, `authorize_layer.go` implement a
  Mongo-flavoured filter/orderBy/index/authorization layer consumed by every `Get*` method with a
  `filter string` parameter (public API — see `GET /api/v1/...?filter=` usage in controllers).
  SQL translation must be byte-compatible on the same fixtures (risk R-3). Foundation task:
  differential tests running both engines over captured fixtures before any domain flips.
- **Strangler switch**: `src/serviceprovider/data.go:8` flips from `*data.JsonDatabase` to
  `DatabaseService` selection by `DB_TYPE`. Before flipping, enumerate concrete-type usages
  (`grep -n "\.SaveAs\|\.SaveNow\|\.Backup\|JsonDatabase\b" src --include=*.go`).
- Atomic multi-record flows needing `db.Transaction`: role/claim assignment in
  `CreateUser/UpdateUser`, orchestrator host + nested VM records, catalog manifest versions —
  inventory during the round (AC "Transactions").
- WAL + busy_timeout are applied by #418's factory already (task 03); verify pragmas in tests.
- Errors: keep `ErrXxx = errors.NewWithCode(...)` sentinels — controllers branch on codes
  (`src/data/users.go:14-26` archetype). Map `gorm.ErrRecordNotFound` → existing `Err*NotFound`.

## Engineering-rule triggers (feature plan §6)

- ER-2/ER-3: per-domain implementation files in `src/database/` must stay < 1,000; `src/data/orchestrator.go` (933) is
  only touched by *deprecation comments* in #426, so breach risk is mainly the new files — split
  orchestrator-domain SQL if needed.
- ER-5: each domain flips to GORM only after its differential tests pass; during the chain,
  mixed engines coexist (interface + `Legacy()`), so `DB_TYPE=postgres` **cannot ship as default**
  until every domain is ported — default stays `sqlite`+legacy until the final flip task.

## Fit in the feature

- **Depends on**: #418 (factory), #419 (tagged models), #420 (config).
- **Feeds**: #424 (migration writes through this layer), #425 (suite parity), #426 (this is what
  makes legacy "reachable only via fallback"), #427.
- **Recommended task chain shape** (validate in round): F1 interface capture + compile assertions →
  F2 filter/order/index/authorize port + differential harness → F3 switch mechanism (default still
  legacy) → domain chains users→roles/claims→api_keys→user_config→config→packer→catalog→
  catalog_managers→orchestrator→jobs→snapshots→reverse_proxy→enrollment → S1 final flip +
  `Legacy()` audit.
