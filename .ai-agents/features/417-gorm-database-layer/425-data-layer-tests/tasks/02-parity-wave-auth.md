# Task 02 — Engine-parity wave: **auth** (users, roles, claims, roles+claims, api keys) + filter/order/authorize differentials

**PBI** [#425](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 01 · **PR size:** M
**Plan refs:** [D-1](./plan.md#d-1--existing-tests-pass--three-tier-strategyresolves-the-ac-ambiguity), [D-5](./plan.md#d-5--parity-suite-mechanics-differentials-promoted-not-duplicated)

## Context

Wave 1 of the conversion. 36 legacy test functions across the authentication domains become
parity scenarios — each legacy `Test*` body is lifted (fixtures, error expectations, ordering
assertions preserved from the original file/line, e.g. `src/data/users_test.go:14`) into a
`testutil.Scenario` that `RunOnEngines` executes under both `legacy/` and `gorm/`. This wave also
lands the **cross-engine differential** over the Mongo-flavoured query layer
(`filter_layer.go`/`order_layer.go`/`authorize_layer.go` — 370 lines total with **zero legacy
tests**), which is feature risk R-3's suite-level net and part of AC "patterns not currently covered".

## Objective

`go test ./database/parity/...` reports one subtest pair (`legacy`+`gorm`) per converted scenario
for all ~35 auth scenarios; every pair passes; a cross-engine differential asserts identical ID
sequences for the filter/order table on a shared seed; the ledger in `parity/doc.go` ticks the auth
rows.

## Files

| File | Op |
|---|---|
| `src/database/parity/seed_auth_test.go` | create — auth seed helpers (root user, user w/ roles+claims, api keys incl. internal/external types) |
| `src/database/parity/users_parity_test.go` | create — 6 scenarios from `users_test.go` |
| `src/database/parity/roles_claims_parity_test.go` | create — 15 scenarios: `roles_test.go` 4 + `claims_test.go` 4 + `roles_claims_test.go` 7 |
| `src/database/parity/api_key_parity_test.go` | create — 13 scenarios from `api_key_test.go` (2 pure-function tests stay Tier 1 only — ledger row) |
| `src/database/parity/filter_differential_test.go` | create — cross-engine filter/order/authorize differential |
| `src/database/parity/doc.go` | modify — auth rows of the ledger flip draft→converted |

## File size guard (ER-4)

| File | now | After | Verdict |
|---|---|---|---|
| (all new files above) | new | ≤ 280 each | ok |
| `parity/doc.go` | ~80 (task 01) | ~110 | ok |
| Any `src/data/*` file | — | untouched | ok |

## Implementation spec

### Conversion recipe (applies to waves 02–04 — canonical here)

1. Read the legacy function (path:line listed in the ledger). Copy its arrangement → scenario
   `Run` body; assertions keep the **same** expected sentinels (`data.ErrUserNotFound`, …) and same
   expected list lengths — never weaken to `NoError` (coverage must match or exceed).
2. Register: `testutil.RunOnEngines(t, []testutil.Scenario{…}, testutil.LegacyEngine(), testutil.GormEngine())`.
3. Engine-legitimate divergence (ordering without an ORDER BY, update-timestamp freshness…) →
   pin the *observable contract*, not the implementation: e.g. assert `len(got)==want` + ID-set
   equality instead of slice order when the legacy code makes no ordering promise; where the public
   API **does** promise order (`Get*` with orderBy), assert exact sequence.
4. If the gorm leg fails: that's a #421 bug — **do not** paper it in test (no skip). Either fix #421
   in the same PR if ≤ ~20 lines with its own tests (ER-5: suite green at merge), or block this
   scenario as `t.Skip("blocked-by:#421 follow-up #NNN — <one-line cause>")` **and** file the issue;
   ledger row says `blocked` with the link. Skips without links are review-fail.
5. Names: scenario name = legacy function name minus `Test` (`TestCreateUser` →
   `users/create_user`) so the ledger and CI logs cross-reference literally.

### Scenario inventory (35 — names verified against `grep -n "^func Test"` 2026-11-03)

- **users (6):** Create·FailsMissingFields·Update·UpdateBlockStatus·AddAndRemoveRoleToUser·Delete.
  `AddAndRemoveRoleToUser` is the embedded-slice round-trip canary (R-2): assert role **and** claim
  slices survive re-read on both engines.
- **roles (4) / claims (4):** incl. `DeleteInternal*FailsWithoutRoot` authorization semantics —
  build the ctx with a non-root `ApiUser` (pattern from `authorize_layer.go` + legacy tests).
- **roles_claims (7):** success/duplicate/not-found matrices — pre-map for task 04's error matrix.
- **api_key (13):** expiration logic (`TestApiKeyExpirationLogic`, expiry field round-trip — string
  dates, #419 typing), user-ownership filtering, internal-key visibility rules (`TestInternalKeysNotInListButRetrievableByID`
  is an *authorization* scenario of the get-list path). Excluded pure: `TestNormalizeApiKeyType`,
  `TestFilterInternalKeys` (ledger notes "pure domain fn — engine-agnostic, stays Tier 1").

### `filter_differential_test.go` — the R-3 net

Shared seed (3 users + 2 api keys incl. internal) built once per expression set; harvest the
expression table from `src/data/filter_layer.go` operators (grep its operators: `=`, `!=`, `>`,
`<`, `contains`, comma/`and`-chained clauses, quoted values — **enumerate from source at task time,
min 12 expressions** incl. one with a quoted value containing a comma, one unknown-property
(empty-result semantics), one invalid-syntax (error semantics)). For each expression:
`legacyIDs == gormIDs` (exact sequence; ties broken by documented orderBy) for `GetUsers`,
`GetApiKeys`, and orderBy runs (`+name`, `-created`-style per `order_layer.go`). Authorization:
same expression via ctx with/without claims → both engines agree on empty-vs-visible
(`authorize_layer.go` semantics). New coverage — nothing like it exists today.

## Out of scope

Orchestrator/infra/jobs domains (tasks 03/04) · error-matrix table (task 04; wave scenarios carry
their own error assertions meanwhile) · fixing #421 beyond the ≤20-line allowance · touching
`filter_layer.go` itself · Postgres leg (task 05 runs the *auth scenarios* against PG via the
generic engine list — no extra work here).

## Validation

- [ ] `cd src && go test ./database/parity/... -count=1 -v 2>&1 | tee parity-wave-auth.log` — then
      `grep -c 'legacy' parity-wave-auth.log` ≥ 35 and same for `gorm`; every auth scenario appears in **both**
- [ ] `cd src && go test ./database/parity/... -count=5 -shuffle=on -race`
- [ ] `cd src && go test ./... -count=1` — Tier 1 untouched and green; **paste `time` output** (budget trail, D-9)
- [ ] Ledger: `grep -c '| converted' src/database/parity/doc.go` ≥ 35 (+2 "pure" +auth exclusion rows)
- [ ] `grep -rn "t.Skip" src/database/parity/*.go` → only skips with `blocked-by:` links (paste in PR)
- [ ] Zero assertions weaker than the source test (review checklist item; spot-check 3 named scenarios vs originals in PR)

## Rollback

Revert commit — five new parity files gone; `doc.go` ledger reverts; testutil untouched. Legacy
suite never referenced these files.

## PBI acceptance mapping

**AC 1** (auth slice of "updated and pass against the GORM implementation", with legacy proof
intact), **AC 2** (every gorm leg on the task-01 harness), **AC 6** (`-count=5 -shuffle -race`
evidence), **AC 5** partial (not-found/duplicate expectations inside scenarios). Constraint: wall
time delta pasted.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `425-02`.** Cut `slice/425-02` from `sql-migration` HEAD
(`.worktrees/slice-425-02`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
