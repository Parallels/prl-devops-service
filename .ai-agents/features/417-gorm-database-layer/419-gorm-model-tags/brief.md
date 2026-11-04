# PBI Brief #419 — Data model conversion with GORM struct tags

| Field | Value |
|---|---|
| PBI | [#419](https://github.com/Parallels/prl-devops-service/issues/419) — `[PBI] Data model conversion with GORM struct tags` |
| Labels | `triage`, `pbi` · Assignee: SivaParvathi3112 |
| Round | Deep-dive round **3** |
| Upstream feature docs | [`brief.md`](../brief.md) · [`plan.md`](../plan.md) — required reading |
| Planning status | **Fully planned** — [plan.md](./plan.md) + [01-split-virtual-machine.md](./tasks/01-split-virtual-machine.md), [02-tagging-wave-a-auth.md](./tasks/02-tagging-wave-a-auth.md), [03-tagging-wave-b-infra.md](./tasks/03-tagging-wave-b-infra.md), [04-tagging-wave-c-orchestrator-snapshots.md](./tasks/04-tagging-wave-c-orchestrator-snapshots.md), [05-relationships-embed-fidelity.md](./tasks/05-relationships-embed-fidelity.md), [06-automigrate-verification-mapping-final.md](./tasks/06-automigrate-verification-mapping-final.md) |

## Description (verbatim from GitHub)

Annotate all existing domain structs with GORM struct tags so that GORM can map them to relational
tables. This includes defining primary keys, indexes, relationships, and any custom column names or
constraints required to faithfully represent the existing data model.

### User Story

As a **backend engineer**, I want all data models to carry GORM annotations, so that the ORM can
automatically manage schema creation and data access without hand-written SQL.

### Acceptance Criteria

- [ ] **All models annotated**: Every struct that is persisted has been updated with `gorm:"..."` tags for primary key, column names, indexes, and relationships.
- [ ] **Relationships defined**: `HasOne`, `HasMany`, `BelongsTo`, and `ManyToMany` associations are declared using GORM conventions where applicable.
- [ ] **Auto-migrate verified**: Running `db.AutoMigrate(...)` against a fresh SQLite file produces the expected schema with no errors.
- [ ] **No data loss**: The annotated models produce a schema that can round-trip all fields present in the legacy data store.
- [ ] **Soft deletes**: Models that support logical deletion use GORM's `gorm.DeletedAt` field where appropriate.

### Definition of Done

- [x] Code implemented following best practices.
- [x] Unit tests written and passing.
- [x] Code reviewed and approved.
- [x] Merged into the main branch.
- [x] Documentation updated (if applicable).
- [x] Deployed to staging/production environment.

### Assumptions and Constraints

- **Assumption**: The existing field names are sufficiently descriptive; column name overrides will only be added where the default GORM convention produces a conflict.
- **Constraint**: Struct tags must not break JSON serialisation — `json:"..."` and `gorm:"..."` tags must coexist correctly. ⚠️ These structs are **API wire formats** (used directly in REST responses) — a JSON tag change is a breaking API change. Round-trip tests must assert byte-level JSON compatibility.

### Dependencies

- **PBI-1**: GORM must be present and configured before model annotations can be validated. → #418 (done: connection factory + `:memory:` support).

## Repo context (verified 2026-11-03)

- 24 model files, ~2,462 lines in `src/data/models/` — all plain structs, JSON tags only
  (`src/data/models/user.go` is the archetype: `ID string` PK-style, `CreatedAt string`).
- **String dates everywhere** (`CreatedAt`/`UpdatedAt`/`BlockedSince`…). Decision needed in the
  PBI plan: keep `string` columns (round-trip-safe, ugly typing) vs `time.Time` + custom scan
  (GORM-native but touches JSON shape). Recommendation: keep `string` in v1 of the schema, revisit later.
- **Embedded slices → relationships**: `User.Roles []Role`, `User.Claims []Claim`
  (`user.go:11-12`); `Role.Claims`; `OrchestratorHost` nested VM lists; `CatalogManifest` nested
  structures. These become join tables (`user_roles`, …) or serialised JSON columns — the #419
  plan must decide **per field**, with the constraint "no data loss on round-trip" (#424 depends
  on the mapping table it produces).
- `DbRecord` (`db_common.go`, `IsLocked/LockedBy/LockedAt`) is embedding metadata used with
  `LockRecord/UnlockRecord` (`src/data/main.go:399-426`) — decide: columns or dropped in SQL
  layer (it's transient locking state, likely **dropped** from SQL models but kept for legacy compat).
- **AutoMigrate verification** uses #418's factory with a temp-dir SQLite path — no mocks.
- Soft-delete candidates: models with `IsDeleted`-ish semantics — inventory them in the plan
  (`catalog_managers.go`, `catalog.go`, `orchestrator.go`… grep for delete flags) before applying
  `gorm.DeletedAt`; blanket application would change list semantics (#421 parity).

## Engineering-rule triggers (feature plan §6)

- **ER-3 REFACTOR-FIRST**: `src/data/models/virtual_machine.go` is **1,018 lines** — split before
  adding tags (this task owns the split).
- ER-2: model files after tagging stay < 1,000 (split by aggregate if a domain outgrows one file).
- OQ-3 (table naming) resolved during this round — proposal in feature plan: GORM defaults
  (snake_case plurals) + `TableName()` only where the default collides.

## Fit in the feature

- **Depends on**: #418 (connection), #420 (path/type config to run AutoMigrate against).
- **Feeds**: #421 (all queries run on these tagged models), #424 (field→column mapping IS the migration spec).
- Deliverable that outlives this PBI: a **mapping table document** (legacy field → column / join
  table / dropped) produced in `plan.md` — #424's spec derives from it verbatim.
