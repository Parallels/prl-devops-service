# Task 09 — Domain port: user configs

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 06 (authz patterns) · **PR size:** S
**Plan refs:** [D-1](../plan.md#d-1--hybrid-delegate-gormdatabase-adopts-the-recommended-architecture-refined), [D-4](../plan.md#d-4--query-language-port-exact-semantics-go-evaluation-first-gated-sql-push-down-later) (`GetOwnRecords` / authz rows), [D-9](../plan.md#d-9--transaction-inventory-ac-transactions), [D-11](../plan.md#d-11--error-mapping-ac-error-handling); tables per [#419 mapping doc](../../419-gorm-model-tags/plan.md) (D-12)

## Context

User configs are per-user key/value documents addressed two ways — by id **and** by slug — and every method takes
the owner's `userID` as an explicit parameter, so scoping is a mixture of an owner predicate and the authz layer
task 03 ported. This task's whole risk surface is *leakage*: a GORM query that forgets the owner predicate
silently serves one user's configs to another, and a parity suite that only ever seeds one user cannot see it.
Hence the multi-user fixture below is mandatory, not optional.

## Objective

`src/database/user_config.go` implements the 6 user-config methods on GORM with owner-scoped queries;
`delegate_user_config.go` (created by task 04) is **deleted**; parity `DomainSuite` "user_configs" green with
≥2 seeded users and cross-user probes.

## Files

| File | Op |
|---|---|
| `src/database/user_config.go` | create (~140 lines) |
| `src/database/delegate_user_config.go` | delete (as created by task 04) |
| `src/database/user_config_test.go` | create |
| `src/database/parity/suite_user_configs_test.go` | create |

## File size guard

Actuals: legacy `src/data/user_config.go` = **145** lines (**untouched** — #426 territory). Projections (new
files, not yet written): `database/user_config.go` ~140, `user_config_test.go` ~120,
`parity/suite_user_configs_test.go` ~110. All well under the 500-line ER-2 cap; no split foreseen.

## Implementation spec

**Method inventory (implement exactly these; source `src/data/`):**

| # | Method | Legacy | SQL shape / notes |
|---|---|---|---|
| 1 | `GetUserConfigs(ctx, userID, filter)` | user_config.go:17 | Load the owner's rows (`WHERE user_id = ?`, rowid order) → `query.ParseFilter` + `query.FilterByProperty` (D-4: unanchored Go regexp, **not** equality) → authz per D-4 row (confirm at :17-41 whether legacy additionally calls `GetAuthorizedRecords`; the explicit `userID` predicate is the primary scope and must stay) |
| 2 | `GetUserConfig(ctx, userID, idOrSlug)` | :42 | Owner-scoped id-or-slug lookup: `WHERE user_id = ? AND (id = ? OR slug = ?)` + Go-side `EqualFold` fallback (D-4 ASCII fast path). Empty `idOrSlug` ⇒ the legacy empty-input sentinel **before** any query (mirror :42-56 — do not let `slug = ''` match a row) |
| 3 | `CreateUserConfig(ctx, cfg)` | :57 | Pre-checks from the body (`:57-78`): duplicate id and duplicate **slug within the owner** (global or per-owner uniqueness is #419's decision — read the mapping doc, then mirror legacy's check); timestamps via `helpers.GetUtcCurrentDateTime()`; single-row insert |
| 4 | `UpsertUserConfig(ctx, cfg)` | :79 | Find (owner + id/slug, per :79-92) → update if present else insert. Keep it **two statements, non-transactional** (D-9 lists no tx for this domain, and legacy is non-atomic here — "non-atomic-in-legacy, stays non-atomic (parity)"); document the race window in the PR instead of silently tightening it |
| 5 | `UpdateUserConfig(ctx, cfg)` | :93 | Load (owner-scoped) → field-level write with the legacy assignment set from `:93-123` + `UpdatedAt` refresh; do **not** use GORM `Save` of the incoming struct if legacy copies only some fields (partial-update drift is the classic GORM port bug) |
| 6 | `DeleteUserConfig(ctx, userID, idOrSlug)` | :124 | Owner-scoped id-or-slug resolution (same predicate as #2) → not-found sentinel → single-row delete |

**Owner scoping (task 03 patterns).** Two layers, both required: (a) the explicit `userID` parameter becomes a
`WHERE user_id = ?` predicate — legacy filtered the in-memory slice the same way; (b) the ported authz helpers
from `database/query` (`authz.go`, task 03) applied exactly where legacy applied them. `query.GetOwnRecords`
(`api_key.go:19-32` semantics, D-4) is the *context*-based analogue — check at task start whether
`user_config.go` calls it or relies purely on the `userID` argument, and mirror; **do not add it** where legacy
did not (a "safety improvement" here changes results for callers that pass another user's id, e.g. admins).
Root/microservice bypass lives in the authz layer (D-4 row) — do not re-implement it in the query.

**Query approach.** D-4 baseline for #1: load → Go-side filter; **no `OrderByProperty`** (user-configs is not one
of the three `UpdatedAt desc` call sites). Push-down opt-in after green parity, limited to owner+id / owner+slug
lookups (`WHERE user_id = ? AND id = ?`) and, only if #419 mapped `slug` with a per-owner unique index, the
duplicate pre-checks — regexp filters stay Go-side.

**Atomicity.** **No tx sites.** D-9's inventory has no user-configs row ⇒ reads and writes are single-row;
`UpsertUserConfig` deliberately stays two statements (legacy parity). State "0 tx sites, checked against D-9" in
the PR description.

**Error mapping (D-11, via `mapGormError`):**

| GORM outcome | Sentinel returned |
|---|---|
| `gorm.ErrRecordNotFound` | `interfaces.ErrUserConfigNotFound` (exact relocated name — the `Err*` block moved to `interfaces/errors_*` by task 01; identity per D-3) |
| unique violation on the id / `(user_id, slug)` index | the legacy duplicate sentinel (`ErrUserConfigAlreadyExists`-family — exact name from the relocated block), only where #419 actually indexed the column; otherwise the pre-check is the whole story |
| empty `idOrSlug` | the legacy empty-input sentinel, returned before any query |
| any other error | `errors.NewWithCode(fmt.Sprintf("…: %v", err), 500)` |
| `!IsConnected()` | `interfaces.ErrDatabaseNotConnected` first guard in every method |

**Parity test sketch (`src/database/parity/suite_user_configs_test.go`):**

```go
func TestUserConfigsParity(t *testing.T) {
	legacy := parity.NewLegacyEngine(t)
	sql := parity.NewSQLEngine(t, false)
	ctx := parity.TestCtx(t)                 // helper from task 03 (user u1)

	parity.SeedBoth(t, func(db interfaces.DatabaseService) {
		_, _ = db.CreateUser(ctx, seedUser("u1"))
		_, _ = db.CreateUser(ctx, seedUser("u2"))
		for _, pair := range [][2]string{{"u1", "theme"}, {"u1", "layout"}, {"u2", "theme"}} {
			_, _ = db.CreateUserConfig(ctx, seedCfg(pair[0], pair[1])) // same slug, two owners ⇒ per-owner dup probe
		}
	})

	for _, uid := range []string{"u1", "u2", "nobody"} {
		for _, f := range []string{"", "slug=the.*", "id=nope"} {
			want, wErr := legacy.GetUserConfigs(ctx, uid, f)
			got, gErr := sql.GetUserConfigs(ctx, uid, f)
			parity.AssertSameList(t, "GetUserConfigs["+uid+"/"+f+"]", want, got)
			if (wErr == nil) != (gErr == nil) { t.Fatalf("%s/%s: %v vs %v", uid, f, wErr, gErr) }
		}
	}
	// cross-user probe: u1 asks for u2's config by slug — identical verdict on both engines
	for _, db := range []interfaces.DatabaseService{legacy, sql} {
		_, err := db.GetUserConfig(ctx, "u1", "theme")   // must NOT see u2's "theme"
		t.Logf("cross-user slug lookup err=%v", err)
	}
}
```

The `{"u1","theme"}` / `{"u2","theme"}` pair is the point: a missing owner predicate turns the cross-user probe
green on the SQL side while legacy says not-found — `AssertSameList` catches it only if both users own the same
slug.

**Edge cases to pin (unit + suite):**
- empty collection: owner with zero configs ⇒ same nil-vs-empty shape as legacy (assert explicitly, `reflect.DeepEqual`
  distinguishes them).
- not-found: unknown `userID`, unknown slug, unknown id, empty `idOrSlug` ⇒ four assertions on the same sentinel
  identity via `errors.Is`.
- dup slug: same owner, same slug ⇒ duplicate sentinel + `COUNT(*)` = 1; **different** owner, same slug ⇒ allowed
  (or rejected) exactly as legacy — pin whichever it is, and make sure #419's index agrees with it.
- upsert both branches: existing row updated (fields **not** passed in stay untouched), missing row inserted.
- `userID` mismatch: `cfg.UserID` in the payload vs the caller's context — mirror legacy (which one wins) rather
  than choosing; confirm at task start in `:57-78`.

## Out of scope

Provider flip / `DB_TYPE` selection (17/18); users + `user_roles`/`user_claims` (06), api-keys (07),
roles/claims (05), catalog (11), the remaining domains (12–15) and their delegates; the authz/filter layer itself
(tasks 02/03 — consume it); AutoMigrate and column/index design (#419/#420); `data.json` migration (#424).

## Validation

- [ ] `cd src && go build ./... && go vet ./data/... ./database/... ./serviceprovider/...`
- [ ] `cd src && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./database/parity -run 'DomainSuite/user_configs' -count=1` (suite name per task 03 harness)
- [ ] `cd src && go test ./... -count=1`
- [ ] Parity harness opt-in: run the matrix with a root context **and** a plain user context (authz bypass vs scoped)
- [ ] `grep -c "Transaction(" src/database/user_config.go` → 0 (matches D-9 — no tx row for this domain)
- [ ] `grep -n "user_id\|UserID" src/database/user_config.go` → every read/write path owner-scoped
- [ ] `grep -n "g.legacy" src/database/user_config.go` → zero hits
- [ ] Cross-user leakage test green (u1 cannot read u2's config by slug or by id)

## Rollback

Revert the PR: `delegate_user_config.go` returns as task 04 wrote it; domain, unit and parity files removed. No
schema change (the table belongs to #419/#420) ⇒ no data step. 09 is a leaf in the dependency graph — nothing
downstream blocks on it before task 17, so the revert can sit for as long as needed.

## PBI acceptance mapping

| AC | How this task moves it |
|---|---|
| All CRUD via GORM verified against SQLite | 6 methods incl. id/slug dual addressing + upsert both branches |
| Transactions | "No tx needed" verified against D-9 and recorded (legacy non-atomic parity kept) |
| Error handling (wrapped, consistent) | Mapping table on `mapGormError`, incl. empty-input sentinel before query |
| Legacy layer bypassed | Owner-scoped SQL reads, zero `g.legacy`; cross-user leakage asserted |

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-09`.** Cut `slice/421-09` from `sql-migration` HEAD
(`.worktrees/slice-421-09`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
