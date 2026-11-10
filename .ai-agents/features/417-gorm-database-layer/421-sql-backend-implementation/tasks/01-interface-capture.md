# Task 01 — DatabaseService interface capture + compile assertions + provider field (S1)

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** #418 task 05, #420, #419 (merged) · **PR size:** M
**Plan refs:** [D-2](../plan.md#d-2--interface-capture-all-157-methods-one-file-per-domain-exact-signatures), [D-3](../plan.md#d-3--error-sentinels-move-to-interfaces-aliases-stay-in-data-identity-is-load-bearing), [D-6](../plan.md#d-6--three-step-provider-cutover-3-steps-01-interface-capture--17-api--18-selection) (S1)

## Context

R-1's mitigation is to freeze the whole 157-method surface in one PR *before* any domain port starts, so
signatures can't drift mid-chain. Interface capture touches nothing at runtime: `JsonDatabase` already satisfies
the interface by construction, the provider still hands out the legacy instance, and every existing call site
compiles unchanged. The load-bearing subtlety is D-3: several startup files compare errors by **pointer identity**
(`src/startup/seeds/role_claims.go:39,43`, `src/startup/migrations/add_api_key_own_claims.go:61`), so the `Err*`
sentinels must become shared values reachable without `database → data` imports.

## Objective

`interfaces.DatabaseService` exists as 17 composed per-domain interface files matching the 157-method inventory
verbatim; `var _ interfaces.DatabaseService = (*data.JsonDatabase)(nil)` compiles in CI; every `Err*` sentinel
lives in `interfaces` with identity-preserving aliases in `data`; provider gains a `Legacy()` escape hatch and its
`JsonDatabase` field is typed `interfaces.DatabaseService` while `GetDatabaseService` keeps returning the concrete
legacy type.

## Files

| File | Op |
|---|---|
| `src/data/interfaces/database_service.go` | create — aggregate composition |
| `src/data/interfaces/{users,roles,claims,api_keys,catalog,catalog_managers,orchestrator,snapshots,jobs,reverse_proxy,user_config,config,schema,packer_templates,enrollment_tokens,lifecycle}.go` | create — 16 domain files |
| `src/data/interfaces/errors_{users,roles,claims,api_keys,jobs,orchestrator,catalog,reverse_proxy,user_config,packer_template,enrollment_token,config,main}.go` | create — relocated sentinel blocks |
| `src/data/{users,roles,claims,api_key,jobs,orchestrator,catalog,reverse_proxy,user_config,packer_template,orchestrator_enrollment_token,main}.go` | modify — replace `var (…)` sentinel blocks with alias blocks (in-place, same var names) |
| `src/data/interfaces_compilation_test.go` | create — R-1 assertion |
| `src/serviceprovider/main.go` | modify — field retype + `legacyDatabase` field + `Legacy()` (≤ 15 lines) |
| `src/serviceprovider/data.go` | modify — internal assertion only, signature unchanged |

## File size guard

| File | Lines (wc -l 2026-11-03) | After | Verdict |
|---|---|---|---|
| `src/serviceprovider/main.go` | 530 | ~545 | ok |
| `src/serviceprovider/data.go` | 25 | ~30 | ok |
| `src/data/main.go` | 686 | ~686 (swap, −0) | ok |
| `src/data/users.go` … domain files | 422 / 933 / 714 … | unchanged ±2 | ok |
| new `interfaces/*.go` | new | 20–90 each (157 methods across 17 files) | ok |

No new file approaches 1,000 (ER-2/ER-4). No over-cap file is edited → no ER-3 trigger.

## Implementation spec

### 1. Sentinel relocation (D-3, mechanical, commit #1)

Pattern (example `data/users.go:14-27`):

```go
// interfaces/errors_users.go
package interfaces
var (
    ErrUserNotFound        = errors.NewWithCode("user not found", 404)
    ErrUserAlreadyExists   = errors.NewWithCode("user already exists", 400)
    // … all 12 verbatim from data/users.go:15-26
)

// data/users.go — replaces the var block in place:
var (
    ErrUserNotFound              = interfaces.ErrUserNotFound
    ErrUserAlreadyExists         = interfaces.ErrUserAlreadyExists
    // … every former name, message/code no longer duplicated
)
```

Relocate from the exact source blocks: `users.go:14-27`, `api_key.go:15-18`, `jobs.go:16-18`, `claims.go`
(`ErrClaim*` block), `roles.go` (`ErrRole*`), `orchestrator.go` (`ErrOrchestrator*`), `catalog.go`,
`reverse_proxy.go`, `user_config.go`, `packer_template.go`, `orchestrator_enrollment_token.go`, `config.go`,
`main.go:21-24` (`ErrDatabaseNotConnected`, `ErrNotAuthorized` → `interfaces/errors_main.go`). Grep
`^var (` in `src/data/*.go` for completeness; aliases must preserve *every* exported name exactly (they are the
public identity surface).

### 2. Interface files (commit #1, same PR — pure additive)

Method lists per file = the per-domain inventories in tasks 05–16 (each task file carries the authoritative
name+`file:line` list; 01 copies them into the interface — signature text must be copied from source, not
retyped). Header pattern:

```go
// interfaces/users.go
package interfaces

import (
    "github.com/Parallels/prl-devops-service/basecontext"
    "github.com/Parallels/prl-devops-service/data/models"
)

// UsersService is the data-access surface implemented by data.JsonDatabase (legacy)
// and database.GormDatabase (SQL). Signatures mirror data/users.go verbatim — see #421 D-2.
type UsersService interface {
    GetUsers(ctx basecontext.ApiContext, filter string) ([]models.User, error)
    GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, error)
    // … all 11
}
```

`interfaces/database_service.go`:

```go
type DatabaseService interface {
    LifecycleService; UsersService; RolesService; ClaimsService; ApiKeysService
    CatalogService; CatalogManagersService; OrchestratorService; SnapshotsService
    JobsService; ReverseProxyService; UserConfigsService; ConfigService; SchemaService
    PackerTemplatesService; EnrollmentTokensService
}
```

Oddities to copy faithfully: no-`ctx` methods (`GetCatalogManagers()` `catalog_managers.go:16`,
`GetCatalogManager(id string)` `:23`, `SetListVMSnapshotsByVMId` `vm_snapshots.go:7`,
`GetListVMSnapshotsByVMId` `:25`); map-return-no-error methods (`orchestrator.go:398,425,452,479,507`);
`[]*models.ReverseProxyHost` (`orchestrator.go:630`); `*models.OrchestratorHost` pointer receivers of args
(`UpdateOrchestratorHost(ctx, host *models.OrchestratorHost)` `:200`).

### 3. Assertion + provider (commit #2)

```go
// src/data/interfaces_compilation_test.go
package data_test
var _ interfaces.DatabaseService = (*data.JsonDatabase)(nil)   // R-1 CI guard
```

Provider S1 (`serviceprovider/main.go` struct line ~41; 4 init branches at :76,79,97,99,212,214,233,235 assign
concrete values — still assignable to the interface field):

```go
type ServiceProvider struct {
    JsonDatabase    interfaces.DatabaseService // S1: was *data.JsonDatabase
    legacyDatabase  *data.JsonDatabase         // concrete handle, set in each init branch next to the field assign
    SQLDatabase     *gorm.DB                   // from #418 task 05
    // …
}
func (s *ServiceProvider) Legacy() *data.JsonDatabase { return s.legacyDatabase }   // AD-1 escape hatch
```

`GetDatabaseService` (`data.go:8`) keeps `(*data.JsonDatabase, error)` and internally:
`db, ok := provider.JsonDatabase.(*data.JsonDatabase); if !ok { … }` until task 17 flips the signature. All
`SetDatabaseService(globalProvider.JsonDatabase)` injection sites (main.go:218,239 → parallelsdesktop) are changed
in this task to pass `globalProvider.legacyDatabase` (concrete stays concrete; D-7 exception).

## Out of scope

Any `GormDatabase` code (task 04); provider return-type change (17); backend selection (18); any behaviour
change — this PR must be provably inert (`git diff` shows types/locations only).

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./... -count=1` — legacy suite untouched and green (ER-5)
- [ ] `grep -rn "func (j \*JsonDatabase) [A-Z]" src/data/*.go | grep -v _test | wc -l` → 157 AND
      `grep -c "ctx basecontext.ApiContext\|)" src/data/interfaces/*.go`-style method-line count of the 16
      interface files sums to **157** (script in PR description; per-file sums vs. inventory table in plan D-2)
- [ ] `go vet ./data/... ./serviceprovider/...`
- [ ] Identity proof: a test asserting `data.ErrUserNotFound == interfaces.ErrUserNotFound` and
      `data.ErrRoleAlreadyContainsClaim == interfaces.ErrRoleAlreadyContainsClaim` (D-3)
- [ ] `grep -rn "Err[A-Za-z]* *= errors.NewWithCode" src/data/*.go | grep -v _test` → **zero** matches (all relocated)
- [ ] Manual: `go run . api` smoke still boots on legacy (log line identical)

## Rollback

Revert single commit-pair; pure relocation + type widening, no consumers changed. `Legacy()`/`legacyDatabase`
disappear with the revert.

## PBI acceptance mapping

Foundation for **all CRUD (AC-1)** — freezes the surface (R-1); moves nothing user-visible yet; prerequisite for
the "Legacy layer bypassed" audit list (AC-6) created by enumerating provider fields.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-01`.** Cut `slice/421-01` from `sql-migration` HEAD
(`.worktrees/slice-421-01`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
