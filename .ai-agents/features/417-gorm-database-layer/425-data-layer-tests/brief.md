# PBI Brief #425 — Update data layer unit tests

| Field | Value |
|---|---|
| PBI | [#425](https://github.com/Parallels/prl-devops-service/issues/425) — `[PBI] Update data layer unit tests` |
| Labels | `triage`, `pbi` · Assignee: unassigned |
| Round | Deep-dive round **6** |
| Upstream feature docs | [`brief.md`](../brief.md) · [`plan.md`](../plan.md) — required reading, esp. §12 test strategy |
| Planning status | **Fully planned** — [plan.md](./plan.md) · tasks [01](./tasks/01-testutil-harness-dsn-isolation.md) · [02](./tasks/02-parity-wave-auth.md) · [03](./tasks/03-parity-wave-infra.md) · [04](./tasks/04-parity-wave-orchestrator-jobs-config.md) · [05](./tasks/05-postgres-offline-integration-ci-job.md) · [06](./tasks/06-migration-edge-case-suite.md) · [07](./tasks/07-flake-perf-sweep-make-ci-audit.md) |

## Description (verbatim from GitHub)

Update all existing unit tests that exercise the data layer to work with the new GORM-based
implementation. Add new tests to cover both backends, edge cases in the migration path, and any
data access patterns not currently covered.

### User Story

As a **backend engineer**, I want comprehensive unit and integration tests for the new data layer,
so that I can refactor and ship with confidence and catch regressions early.

### Acceptance Criteria

- [ ] **Existing tests pass**: All tests that previously tested the legacy data layer have been updated and pass against the GORM implementation.
- [ ] **In-memory SQLite for unit tests**: Unit tests use an in-memory SQLite database (`file::memory:?cache=shared`) for speed and isolation — no test writes to disk.
- [ ] **Backend coverage**: Tests exist for both SQLite and PostgreSQL code paths where behaviour may differ (e.g. type handling, constraint errors).
- [ ] **Migration tests**: The migration script is tested with synthetic legacy data, including edge cases (empty store, partial migration, re-run after success).
- [ ] **Error path coverage**: Tests verify that GORM errors (constraint violations, not-found, connection failure) are correctly wrapped and returned.
- [ ] **No flaky tests**: All new tests are deterministic and do not depend on test execution order.

### Definition of Done

- [x] Code implemented following best practices.
- [x] Unit tests written and passing.
- [x] Code reviewed and approved.
- [x] Merged into the main branch.
- [x] Documentation updated (if applicable).
- [x] Deployed to staging/production environment.

### Assumptions and Constraints

- **Assumption**: The project uses the standard Go testing package; testify or a similar assertion library may be used if already present. ✅ `stretchr/testify v1.11.1` is in `src/go.mod`.
- **Constraint**: Test execution time for the unit test suite must not increase by more than 30 seconds overall (in-memory SQLite should keep this well within range).

### Dependencies

Runs **after** #421 has ported at least most domains (suite-parity work), and after #424 for
migration tests. Earlier PBIs already carry their own task-level tests (ER-5) — this PBI is about
**completeness**, not first coverage.

## Repo context (verified 2026-11-03)

- Existing data-layer tests to migrate/adapt: `src/data/*_test.go` — `main_test.go` (archetype:
  `setupTestDB`/`cleanupTestDB` with temp files + `memoryDatabase = nil`), `users_test.go`,
  `claims_test.go`, `roles_test.go`, `api_key_test.go`, `catalog_test.go`, `catalog_managers_test.go`,
  `concurrency_test.go`, `jobs_timeout_test.go`, `packer_template_test.go`, `reverse_proxy_test.go`,
  `vm_snapshots_test.go`, `orchestrator_test.go`, `config_test.go`, `roles_claims_test.go`,
  `main_test.go`, plus `src/startup/migrations/common_test.go`.
- These tests will run **in parallel** with new GORM tests (both engines live during the feature).
  Plan for two harness families, no shared globals: legacy keeps `setupTestDB`; GORM gets
  `setupGormDB(t)` in `src/database/testutil_test.go` (or exported `testutil` pkg if migrations
  tests need it) built on #418's factory with `:memory:` / `file::memory:?cache=shared`.
- `src/data/testing_reset.go` is gated `//go:build integration` yet `main_test.go` resets
  `memoryDatabase` directly — document/decide the convention in the round (don't "fix" silently).
- **In-memory SQLite nuance (AC 2)**: shared-cache `file::memory:?cache=shared` is one logical DB
  per process — tests sharing it leak into each other; the harness must use unique DSN names
  (`file:test-<name>?mode=memory&cache=shared`) and/or truncate-per-test; `t.Cleanup(close)`
  everything (AC "no flaky").
- **PostgreSQL backend coverage**: unit-offline scope (DSN validation, pool via
  `db.SQLDB().Stats()`, redaction) + real-server scope behind `//go:build integration` with
  `POSTGRES_TEST_DSN`; wire a `postgres` service into `docker-compose.yml` (exists in repo root)
  or a documented `docker run` in `Makefile` target — pick one in the round.
- **Migration edge cases** (AC 4): fixtures under `src/database/legacy_migrator/testdata/` —
  empty store, partial (some collections only), encrypted (key fixture), corrupt-JSON (must not
  crash), re-run after success (marker present).
- Baseline the suite runtime **before** this PBI's PRs land (record `go test ./... ` duration in
  the PR description) to prove the +30 s budget.
- `Makefile` (root) and `.github/` workflows define CI test invocation — new build tags must be
  added there explicitly (integration job or documented manual step; decision in round).

## Engineering-rule triggers (feature plan §6)

- ER-4 guard applies to test files too (< 1,000 lines each; per-domain test files mirror per-domain
  implementation files).
- ER-1: testify already used — assert/require consistently with `src/data/main_test.go` style.

## Fit in the feature

- **Consumes**: everything (#418 harness, #419 schema, #421 domains, #424 migration).
- **Feeds**: #426 (deprecation PRs want green proof), #427 (documented `make test` behaviour stays honest).
- **Also owns**: the differential-harness promotion from #421's foundation task into permanent
  CI-runnable suites.
