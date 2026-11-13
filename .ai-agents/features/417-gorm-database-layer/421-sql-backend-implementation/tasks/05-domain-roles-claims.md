# Task 05 — Domain port: roles + claims

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 04 · **PR size:** M
**Plan refs:** [D-1](../plan.md#d-1--hybrid-delegate-gormdatabase-adopts-the-recommended-architecture-refined), [D-4](../plan.md#d-4--query-language-port-exact-semantics-go-evaluation-first-gated-sql-push-down-later), [D-9](../plan.md#d-9--transaction-inventory-ac-transactions), [D-11](../plan.md#d-11--error-mapping-ac-error-handling); schema per [#419 mapping doc](../../419-gorm-model-tags/plan.md) (D-12)

## Context

Auth primitives first: users (06), api-keys (07), catalog (11) all read roles/claims, so porting them early
removes delegate hops from later parity suites. Both domains carry the **internal-record guards**
(`role.Internal && !IsRootUser(ctx)` — `roles.go:138,166,188`; `claims.go:123`) and the delete cascades that
touch `user_roles`/`role_claims` — the first real GORM transactions of the chain.

## Objective

`src/database/roles.go` + `src/database/claims.go` implement the 15 methods on GORM; `delegate_roles.go` /
`delegate_claims.go` (created by task 04) are **deleted**; parity `DomainSuite` "roles_claims" green.

## Files

| File | Op |
|---|---|
| `src/database/roles.go` | create (~230 lines) |
| `src/database/claims.go` | create (~180) |
| `src/database/delegate_roles.go`, `delegate_claims.go` | delete |
| `src/database/roles_test.go`, `claims_test.go`, `src/database/parity/suite_roles_claims_test.go` | create |

## File size guard

New files ≤ 230 lines (ER-2 ok). Legacy `src/data/roles.go` (266) / `claims.go` (225): **untouched** (#426
territory). No over-cap files touched.

## Implementation spec

**Method inventory (implement exactly these; source `src/data/`):**

| # | Method | Legacy | SQL shape |
|---|---|---|---|
| 1 | `GetRoles(ctx, filter)` | roles.go:23 | `Find` all → `query.FilterByProperty` → `query.GetAuthorizedRecords` — legacy applies authz? **verify while porting** (roles/claims Get* do NOT authz-wrap in legacy; api_keys/catalog/orchestrator/packer do) → mirror exactly what the file does |
| 2 | `GetRole(ctx, idOrName)` | :59 | EqualFold ID-or-Name scan legacy-side; SQL: `WHERE id = ? OR name = ?` + Go-side EqualFold fallback row set (ASCII fast path, D-4) |
| 3 | `CreateRole` | :92 | dup pre-check ×2 (ID, Name — mirror :92-125 checks incl. `ErrRoleAlreadyExists`-family) + `Create`; timestamps `helpers.GetUtcCurrentDateTime()` |
| 4 | `AddClaimToRole` | :126 | **tx**: load role (guard `ErrUpdateInternalRole` :138), dedupe check `ErrRoleAlreadyContainsClaim`, insert `role_claims` row |
| 5 | `RemoveClaimFromRole` | :154 | tx; note legacy matches existing claim by ID **or Name** (`:181` EqualFold both) |
| 6 | `DeleteRole` | :177 | guard internal (`ErrRemoveInternalRole`), **usage scan across user_roles** (`ErrRoleHasUsers` :196-199), delete role + role_claims rows in one tx |
| 7 | `UpdateRoleInternalFlag` | :210 | unguarded targeted `UPDATE` (seeds rely on it — identity of `ErrRoleNotFound` from D-3) |
| 8 | `UpdateRoleDescription` | :228 | targeted update |
| 9 | `UpdateRole` | :244 | field-level update semantics: copy non-zero behaviour from legacy body while porting |
| 10 | `GetClaims(ctx, filter)` | claims.go:21 | load → filter |
| 11 | `GetClaim(ctx, idOrName)` | :57 | like #2 |
| 12 | `CreateClaim` | :91 | dup checks, plain create (non-atomic in legacy — keep) |
| 13 | `DeleteClaim` | :112 | **tx**: internal guard (`ErrRemoveInternalClaim`), user-usage scan (`ErrClaimHasUsers` :130-135), delete claim + cascade `DELETE FROM user_claims/role_claims WHERE claim_id = ?` (legacy cascade claims.go:139-155) |
| 14 | `UpdateClaimMetadata` | :165 | targeted update, no internal guard (:162-164 comment) |
| 15 | `UpdateClaim` | :184 | full-ish update per legacy field set |

**Cross-domain reads:** the usage scans in #6/#13 read `user_roles`/`user_claims` join tables **directly via
GORM models**, not via delegate — avoids legacy/SQL divergence during the chain (legacy read its one store
directly; same principle).

**Filters/order:** `GetRoles`/`GetClaims` = `query.ParseFilter` + `query.FilterByProperty`; no `OrderByProperty`,
no authz wrapper (verify claim-side at :21-55 while porting).

**Error mapping (D-11):** `ErrRecordNotFound`→`ErrRoleNotFound`/`ErrClaimNotFound`; unique violations on
`roles.name`/`claims.name` (only if #419 indexed them; else pre-check is the whole story) → `Err*AlreadyExists`;
guards return the interface sentinels **by identity** (`interfaces.ErrRemoveInternalRole` etc.).

**Quirks to pin (parity suite + unit):** empty `idOrName` → `ErrRoleEmptyNameOrId`/`ErrClaimEmptyNameOrId`
before any query (`roles.go:182`, `claims.go:117`); internal-guard precedes usage-scan; `AddClaimToRole` loads
the claim first → `ErrClaimNotFound` even when the role id is garbage (order of checks at :126-133 — port
order-for-order).

## Out of scope

Users join-table ownership (06 owns `user_roles`/`user_claims` writes; here read-only for guards + cascade
delete of claim rows), AutoMigrate (already #419/#420 wiring), seeds.

## Validation

- [ ] `cd src && go build ./... && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./... -count=1`
- [ ] Parity suite: seeds (default-like roles+claims from `constants.DefaultRoles`/`DefaultClaims`) then matrix
      `GetRoles`/`GetClaims` × filters `{ "", "name=.*", "internal=true", "id=nonexistent" }` → `AssertSameList`
      (order = insertion/rowid; **no ORDER BY** — legacy is unsorted, D-4)
- [ ] Tx-rollback test: fake constraint failure mid-cascade (drop table post-check trick or unique spike) →
      rows unchanged
- [ ] `grep -n "Transaction(" src/database/{roles,claims}.go` — 4 sites matching D-9 rows
- [ ] `grep -n "g.legacy" src/database/roles.go src/database/claims.go` → zero (domain fully ported)

## Rollback

Revert: delegate files return; parity suite skipped via revert. No schema change.

## PBI acceptance mapping

AC-1 (15 methods), AC-3 (4 transaction sites), AC-5 (sentinel identity mapping proven), AC-6 (first domain with
zero legacy reach-through).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-05`.** Cut `slice/421-05` from `sql-migration` HEAD
(`.worktrees/slice-421-05`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
