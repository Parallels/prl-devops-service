# Task 06 — Domain port: users

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 05 · **PR size:** M
**Plan refs:** [D-1](../plan.md#d-1--hybrid-delegate-gormdatabase-adopts-the-recommended-architecture-refined), [D-4](../plan.md#d-4--query-language-port-exact-semantics-go-evaluation-first-gated-sql-push-down-later), [D-9](../plan.md#d-9--transaction-inventory-ac-transactions), [D-11](../plan.md#d-11--error-mapping-ac-error-handling), [D-12](../plan.md#d-12--schema-ownership-reference-dont-re-decide); join tables `user_roles` / `user_claims` per [#419 mapping doc](../../419-gorm-model-tags/plan.md) (D-12)

## Context

Users is the hub domain: after this task, api-keys (07), user-configs (09), catalog (11) and jobs (15) stop
reaching through the delegate for `GetUser*`. Two things make it non-trivial. First, `CreateUser` materialises
default roles/claims and writes three tables — D-9 row 1 makes it the chain's biggest transaction. Second,
legacy **re-hydrates** each user's `Roles[].Claims` from the live roles table on every read (`users.go:44-56`),
so the GORM read shape must preload or authz decisions downstream diverge silently.

## Objective

`src/database/users.go` implements the 11 user methods on GORM (models + `user_roles`/`user_claims` join rows);
`delegate_users.go` (created by task 04) is **deleted**; parity `DomainSuite` "users" green; zero
`g.legacy` references in the users path.

## Files

| File | Op |
|---|---|
| `src/database/users.go` | create (~280 lines) |
| `src/database/delegate_users.go` | delete (as created by task 04) |
| `src/database/users_test.go` | create |
| `src/database/parity/suite_users_test.go` | create |

## File size guard

Actuals: legacy `src/data/users.go` = **422** lines (**untouched** — #426 territory). Projections (new files, not
yet written): `database/users.go` ~280, `users_test.go` ~150, `parity/suite_users_test.go` ~130 — all under the
500-line ER-2 cap. If `users.go` grows past ~300 in review, split `database/users_assign.go` (role/claim
assignment) rather than exceed; keep 1:1 with the delegate deletion list in the PR description.

## Implementation spec

**Method inventory (implement exactly these; source `src/data/`):**

| # | Method | Legacy | SQL shape / notes |
|---|---|---|---|
| 1 | `GetUsers(ctx, filter)` | users.go:29 | `Find` all in rowid order with `Preload("Claims").Preload("Roles.Claims")` → `query.FilterByProperty` → return. **No authz wrap** (legacy :29-58 returns the filtered slice; only re-hydration follows) — mirror that |
| 2 | `GetUser(ctx, idOrEmail)` | :61 | Legacy resolves one string across ID / Email / Username (proven by `:99-109`, which calls `GetUser` with each of the three as an existence probe). SQL: `WHERE id = ? OR email = ? OR username = ?` + Go-side `EqualFold` fallback (D-4 ASCII fast path). Exact field set + precedence order **implementer to confirm at task start: :61-79** |
| 3 | `CreateUser(ctx, user)` | :80 | **tx (D-9 row 1)**: dup pre-checks ×3 ID/Email/Username (`:99-109`) → roles: empty ⇒ `constants.DefaultRoles` each via `GetRole` (identity lookup, push-down allowed), else validate each `role.ID` exists else `errors.NewWithCodef(400, "role %s does not exist", role.Name)` (`:111-126`) → same for `constants.DefaultClaims` (`:128-143`) → `password.Get().Hash(user.Password, user.ID)` (`:145-150`) → `CreatedAt`/`UpdatedAt` = `helpers.GetUtcCurrentDateTime()` (`:152-153`) → insert user + `user_roles` + `user_claims` rows. **Hash the password before the tx body opens** (it is CPU work; keep it out of the lock) |
| 4 | `UpdateUser(ctx, key)` | :159 | Load → field-level write per legacy field set (port the assignment list while implementing) + `UpdatedAt` refresh; join rows rewritten only if legacy touches `Roles`/`Claims` here — **implementer to confirm at task start: :159-201** |
| 5 | `UpdateUserBlockStatus(ctx, key)` | :202 | Targeted `UPDATE` of the blocked field + `UpdatedAt`; no join writes |
| 6 | `UpdateRootPassword(ctx, newPassword)` | :222 | Targets the root user by `constants` root id/username (**implementer to confirm at task start: :222-243**); re-hash with the same `password.Get().Hash(pw, userID)` argument order as `:146` |
| 7 | `DeleteUser(ctx, id)` | :244 | Single-entity removal; **D-9 lists no tx row** ⇒ default is a non-transactional delete. If the body also purges roles/claims off the user, that is ≥2 tables ⇒ one `db.Transaction` and record the added tx in the PR (D-9 table gets a line) — **implementer to confirm at task start: :244-267** |
| 8 | `AddRoleToUser(ctx, userId, roleId)` | :268 | **tx (D-9 row 2)**: load user + role, dedupe check (`ErrUserAlreadyContainsRole`), insert `user_roles` row |
| 9 | `RemoveRoleFromUser(ctx, userId, roleId)` | :308 | **tx**: load user, locate matching role (ID and/or Name — port the matching predicate from :308-345 exactly), delete the join row; legacy "not found"-flavour error preserved |
| 10 | `AddClaimToUser(ctx, userId, claimId)` | :346 | **tx**: mirror #8 on `user_claims` (`ErrUserAlreadyContainsClaim`, `users.go:26`) |
| 11 | `RemoveClaimFromUser(ctx, userId, claimId)` | :386 | **tx**: mirror #9 on `user_claims` |

**Read shape (mandatory).** Legacy overwrites `user.Roles[r].Claims` with the *live* `dbRole.Claims` for every
role, matched by `EqualFold` on ID (`users.go:44-56`). On GORM this is `Preload("Claims").Preload("Roles.Claims")`
— #419 D-7/D-3 makes the preload mandatory, not an optimisation: `authorize_layer` / `ComputeEffectiveClaims`
read `Roles[].Claims`, so a stale or empty preload silently changes authz verdicts in tasks 07/09/11 parity runs.
The preload runs on `GetUsers`, `GetUser` and on the value returned by `CreateUser`.

**Query approach.** Per D-4 baseline: `Find` (rowid order) → ported `query.ParseFilter` + `query.FilterByProperty`
in Go. No `OrderByProperty` (order applied at only 3 sites, none in users — D-4 grep-verified row), no
`GetAuthorizedRecords`. Push-down is opt-in **after** parity is green and limited to the identity lookups
(`WHERE id = ?` replacing `GetRecordIndex` scans); regexp filters on `username`/`email` stay Go-side.

**Atomicity.** Exactly the D-9 rows: 1 tx in `CreateUser`, 1 tx each in `Add`/`Remove`Role/Claim-to-User
(sites 8–11). `db.Transaction(func(tx *gorm.DB) error)`; inside the tx use `tx` for every read and write so the
dedupe-check + join-write pair is one snapshot. `UpdateUser`/`UpdateUserBlockStatus`/`UpdateRootPassword` stay
single-statement — legacy held one `dataMutex.Lock()` for them, which a single UPDATE already matches.

**Error mapping (D-11, via `mapGormError` from task 04):**

| GORM outcome | Sentinel returned |
|---|---|
| `gorm.ErrRecordNotFound` | `interfaces.ErrUserNotFound` (identity, D-3) |
| unique violation on `users.id` / `users.email` / `users.username` | `interfaces.ErrUserAlreadyExists` — pre-checks at `:99-109` remain the primary path; the constraint is only the race backstop |
| unique violation on `user_roles` / `user_claims` | `interfaces.ErrUserAlreadyContainsRole` / `ErrUserAlreadyContainsClaim` (`users.go:25-26`, relocated by task 01) |
| role/claim missing during `CreateUser` | `errors.NewWithCodef(400, "role %s does not exist", …)` / `"claim %s does not exist"` — constructed, not a sentinel; match message **and** code |
| any other GORM/driver error | `errors.NewWithCode(fmt.Sprintf("…: %v", err), 500)` |
| `!IsConnected()` | `interfaces.ErrDatabaseNotConnected` as first guard in every method |

**Parity test sketch (`src/database/parity/suite_users_test.go`):**

```go
func TestUsersParity(t *testing.T) {
	legacy := parity.NewLegacyEngine(t)
	sql := parity.NewSQLEngine(t, false) // :memory: via the real factory, no legacy fallback
	ctx := parity.TestCtx(t)             // helper from task 03; exact name per that package

	parity.SeedBoth(t, func(db interfaces.DatabaseService) {
		_, _ = db.CreateRole(ctx, seedRole("admin"))   // roles/claims first: task 05 already ported
		_, _ = db.CreateClaim(ctx, seedClaim("read-all"))
		_, _ = db.CreateUser(ctx, seedUser("u1"))      // empty Roles/Claims ⇒ defaults materialise
		_, _ = db.CreateUser(ctx, seedUserWith("u2", "admin", "read-all"))
	})

	for _, f := range []string{"", "username=u.*", "email=.*@x\\.io", "blocked=true", "id=nope"} {
		want, wErr := legacy.GetUsers(ctx, f)
		got, gErr := sql.GetUsers(ctx, f)
		parity.AssertSameList(t, "GetUsers["+f+"]", want, got) // reflect.DeepEqual — catches a missing Preload
		if (wErr == nil) != (gErr == nil) { t.Fatalf("filter %q: %v vs %v", f, wErr, gErr) }
	}
}
```

Then the same file drives the mutation matrix: `AddRoleToUser`/`RemoveRoleFromUser`/`Add`/`RemoveClaimToUser`
applied **to both engines in the same order**, re-asserting `GetUsers` after each step (so a join-row bug shows
up as a deep-equal diff, not a green suite).

**Edge cases to pin (unit + suite):**
- empty collection: filter matches nothing ⇒ compare nil-vs-empty explicitly, `reflect.DeepEqual` distinguishes
  them; legacy returns what `FilterByProperty` yields — pin whichever it is in a golden assertion.
- not-found: `GetUser("missing")`, `RemoveRoleFromUser` on an absent role, `AddRoleToUser` with a garbage
  `roleId` ⇒ sentinel identity (`errors.Is`) asserted, not just a non-nil error.
- duplicate ID: `CreateUser` with an existing Email *or* Username (not only ID) ⇒ `ErrUserAlreadyExists` ×3 paths.
- double-assign: `AddRoleToUser` twice ⇒ `ErrUserAlreadyContainsRole`, and exactly one `user_roles` row (`COUNT(*)`
  asserted in `users_test.go`).
- re-hydration: add a claim to a role *after* the user exists ⇒ `GetUsers` on both engines shows the new claim
  inside `Roles[].Claims`.

## Out of scope

Provider flip / `DB_TYPE` selection (tasks 17/18); roles/claims CRUD (task 05, already ported — consume it, do not
touch `database/roles.go`); AutoMigrate + schema design (#419/#420 — consume the mapping doc, raise an amendment
PR if `user_roles`/`user_claims` columns are unspecified); `data.json → SQL` migration (#424); seeds
(task 17); other domains and their delegates.

## Validation

- [ ] `cd src && go build ./... && go vet ./data/... ./database/... ./serviceprovider/...`
- [ ] `cd src && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./database/parity -run 'DomainSuite/users' -count=1` (suite name per task 03 harness)
- [ ] `cd src && go test ./... -count=1` (ER-5 green at every merge boundary)
- [ ] Parity harness opt-in for the authz cross-check: run the users matrix with a non-root `authContext` to prove
      the preloaded claims feed `query.GetAuthorizedRecords` identically in the 07/11 suites
- [ ] `grep -c "Transaction(" src/database/users.go` → 5 (matches D-9 rows 1–2)
- [ ] `grep -n "g.legacy" src/database/users.go` → zero hits (delegate file deleted, domain fully ported)
- [ ] `grep -n "Preload" src/database/users.go` → both `Claims` and `Roles.Claims` present on every read path

## Rollback

Revert the PR: `delegate_users.go` returns untouched (task 04's one-liners are deleted, never edited — D-1 risk
row), domain + tests + parity suite removed, no schema change (tables belong to #419/#420), so no migration to
undo. Downstream tasks stay on the delegate until users lands again.

## PBI acceptance mapping

| AC | How this task moves it |
|---|---|
| All CRUD via GORM verified against SQLite | 11 methods + users parity matrix |
| Transactions | D-9 rows 1–2: `CreateUser` + 4 assignment methods, 5 tx sites |
| Error handling (wrapped, consistent) | Mapping table above on `mapGormError` (task 04) |
| Legacy layer bypassed | `users.go` has zero `g.legacy`; `GetUser*` callers stop hitting `data` once 17 retypes them |
