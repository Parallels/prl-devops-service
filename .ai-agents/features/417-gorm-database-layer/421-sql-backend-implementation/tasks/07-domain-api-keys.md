# Task 07 — Domain port: API keys

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 06 · **PR size:** S
**Plan refs:** [D-1](../plan.md#d-1--hybrid-delegate-gormdatabase-adopts-the-recommended-architecture-refined), [D-4](../plan.md#d-4--query-language-port-exact-semantics-go-evaluation-first-gated-sql-push-down-later) (`GetOwnRecords` row), [D-9](../plan.md#d-9--transaction-inventory-ac-transactions), [D-11](../plan.md#d-11--error-mapping-ac-error-handling); tables per [#419 mapping doc](../../419-gorm-model-tags/plan.md) (D-12)

## Context

API keys are the auth path: `GetUserByApiKey` runs on every API-key request, so this domain is read far more than
written and any behavioural drift is a production auth incident rather than a test failure. Legacy keeps two
listings deliberately apart — the user-facing `GetApiKeys` (own-record filtered) and the admin-only
`GetAllApiKeysIncludingInternal` — and the whole difference lives in `GetOwnRecords` (`api_key.go:19-32`,
D-4's last table row). Task 03 already ported that helper, so this task is mostly correct wiring, not new
semantics.

## Objective

`src/database/api_key.go` implements the 8 api-key methods on GORM; `delegate_api_key.go` (created by task 04) is
**deleted**; parity `DomainSuite` "api_keys" green; key hashing/collision behaviour unchanged (delegated to the
same helpers legacy calls).

## Files

| File | Op |
|---|---|
| `src/database/api_key.go` | create (~180 lines) |
| `src/database/delegate_api_key.go` | delete (as created by task 04) |
| `src/database/api_key_test.go` | create |
| `src/database/parity/suite_api_keys_test.go` | create |

## File size guard

Actuals: legacy `src/data/api_key.go` = **269** lines (**untouched** — #426 territory). Projections (new files,
not yet written): `database/api_key.go` ~180, `api_key_test.go` ~120, `parity/suite_api_keys_test.go` ~110 — all
under the 500-line ER-2 cap. No split expected; if the internal/external split pushes the file past ~220, keep it
as-is rather than break the 1:1 domain↔delegate mapping.

## Implementation spec

**Method inventory (implement exactly these; source `src/data/`):**

| # | Method | Legacy | SQL shape / notes |
|---|---|---|---|
| 1 | `GetApiKeys(ctx, filter)` | api_key.go:56 | `Find` all (rowid order) → `query.ParseFilter`/`FilterByProperty` → **`query.GetOwnRecords`** (D-4 row: keeps records whose `GetUserID()` equals `authContext.User.ID`, and skips filtering entirely when there is no auth context — both branches must be ported) → return |
| 2 | `GetAllApiKeysIncludingInternal(ctx)` | :90 | Admin listing: **no** `GetOwnRecords`, internal keys included; D-4 lists api-keys among the authz-wrapped domains — whether this one additionally calls `GetAuthorizedRecords` is **implementer to confirm at task start: :90-109** and mirror |
| 3 | `GetApiKey(ctx, idOrName)` | :110 | `WHERE id = ? OR name = ?` + Go-side `EqualFold` fallback (D-4 fast path); empty `idOrName` ⇒ the legacy empty-input sentinel before any query (confirm the exact sentinel at :110-136) |
| 4 | `GetUserByApiKey(ctx, apiKey)` | :137 | Hot path. Lookup key by value → on hit resolve the owning user. The key **value** is not stored raw (or is stored hashed) and the internal/external split (internal keys resolve differently) is decided by the body: **implementer to confirm at task start: :137-156**. Read path only ⇒ no write. Resolve the user through the **GORM** users path (`database/users.go`, task 06), never the delegate — that is why 06 gates this task |
| 5 | `CreateApiKey(ctx, apiKey)` | :157 | **tx (D-9 row 9: "`CreateApiKey` (+ internal-key flows)", evidence `api_key.go:157-188`)**: dup/name pre-check, key generation + hashing (see below), timestamps, insert — one `db.Transaction` if the flow writes more than the key row (internal-key flow ⇒ confirm at :157-188 while porting) |
| 6 | `DeleteApiKey(ctx, id)` | :189 | Load → guard (internal keys must not be deletable by non-root — port the guard from :189-215 exactly, error by identity) → single-row delete |
| 7 | `UpdateKey(ctx, key)` | :216 | Field-level update + `UpdatedAt` refresh; field set from the body while porting |
| 8 | `RevokeKey(ctx, id)` | :244 | Targeted `UPDATE` of the revoked/blocked flag + `UpdatedAt`; single statement |

**Key hashing — untouched.** Do **not** reimplement key generation, hashing or collision-retry. Call the same
helper legacy calls (verify the call target by signature while porting `:157-188`; if it lives unexported inside
`src/data`, that is a blocker ⇒ PR note: promote to `helpers`/`security` as a pure move, no logic change). The
column that stores the key/hash follows #419's mapping doc (D-12) — this task writes what the model says.

**Query approach.** D-4 baseline everywhere: `Find` in rowid/insertion order → ported Go-side filter → Go-side
own-record filter → **no `OrderByProperty`** (api-keys is not one of the three ordering call sites). Authz wrap
`query.GetAuthorizedRecords` only where legacy wraps (rows 1/2 above, confirm per-body). Push-down opt-in after
green parity, limited to `WHERE id = ?` / `WHERE name = ?` / the key-value lookup in #4 (generated ids are
lowercase — D-4 push-down note); regexp filters (`name=.*`) stay Go-side.

**Atomicity.** One tx site, `CreateApiKey` (D-9). Everything else is a single statement or a pure read. If the
port shows `DeleteApiKey` touching a second table, add the tx and record it against D-9 in the PR.

**Error mapping (D-11, via `mapGormError`):**

| GORM outcome | Sentinel returned |
|---|---|
| `gorm.ErrRecordNotFound` | `interfaces.ErrApiKeyNotFound` (exact name confirm at :110-136 — the `Err*` block moved to `interfaces/errors_*` in task 01; use the relocated var, identity per D-3) |
| unique violation on `api_keys.name` | `interfaces.ErrApiKeyAlreadyExists` (pre-check stays primary, constraint is the race backstop) |
| internal-key guard tripped | the legacy guard sentinel by identity (`ErrRemoveInternalApiKey`-family — exact name confirm at :189-215) |
| invalid/unknown key in `GetUserByApiKey` | the same sentinel legacy returns for a bad key — **not** `ErrUserNotFound` (auth semantics; mirror the body) |
| any other error | `errors.NewWithCode(fmt.Sprintf("…: %v", err), 500)` |
| `!IsConnected()` | `interfaces.ErrDatabaseNotConnected` first guard in every method |

**Parity test sketch (`src/database/parity/suite_api_keys_test.go`):**

```go
func TestApiKeysParity(t *testing.T) {
	legacy := parity.NewLegacyEngine(t)
	sql := parity.NewSQLEngine(t, false)
	ctx := parity.TestCtx(t)          // helper from task 03
	other := parity.TestCtxAsUser(t, "u2")

	parity.SeedBoth(t, func(db interfaces.DatabaseService) {
		_, _ = db.CreateUser(ctx, seedUser("u1"))
		_, _ = db.CreateUser(ctx, seedUser("u2"))
		_, _ = db.CreateApiKey(ctx, seedKey("k1", "u1"))
		_, _ = db.CreateApiKey(ctx, seedKey("k2", "u2"))
		_, _ = db.CreateApiKey(ctx, seedInternalKey("k-int"))   // internal ⇒ visible only to #2
	})

	// own-records filter is per-caller: same seed, two auth contexts, both engines
	for _, c := range []basecontext.ApiContext{ctx, other} {
		want, _ := legacy.GetApiKeys(c, "")
		got, _ := sql.GetApiKeys(c, "")
		parity.AssertSameList(t, "GetApiKeys[user="+c.GetUserID()+"]", want, got)
	}
	for _, f := range []string{"", "name=k.*", "id=nope"} {
		parity.AssertSameList(t, "GetApiKeys["+f+"]", must(legacy.GetApiKeys(ctx, f)))
	}
	parity.AssertSameList(t, "GetAllApiKeysIncludingInternal",
		must(legacy.GetAllApiKeysIncludingInternal(ctx)), must(sql.GetAllApiKeysIncludingInternal(ctx)))
}
```

Plus a `GetUserByApiKey` loop: for each created key, both engines must resolve the **same** user (or the same
error), and a tampered value `"nope"` must produce the identical error.

**Edge cases to pin (unit + suite):**
- empty collection: user with zero keys ⇒ `GetApiKeys` returns the same nil/empty shape on both engines
  (`reflect.DeepEqual` distinguishes them — pin it, don't normalise).
- no auth context: `GetApiKeys` with a context that carries no user ⇒ legacy `GetOwnRecords` skips filtering
  (D-4 row) ⇒ GORM side must return **all** rows, not zero. This is the highest-value single assertion here.
- not-found: `GetApiKey("nope")`, `DeleteApiKey("nope")`, `RevokeKey("nope")` ⇒ sentinel identity via `errors.Is`.
- duplicate name: `CreateApiKey` twice with the same name ⇒ `ErrApiKeyAlreadyExists`, exactly one row (`COUNT(*)`).
- double revoke: `RevokeKey` on an already-revoked key ⇒ same error/success verdict on both engines.

## Out of scope

Provider flip / `DB_TYPE` selection (17/18); users (06, already ported), roles/claims (05), user-configs (09),
catalog (11), orchestrator/jobs/reverse-proxy (12–15) and their delegates; key-hashing algorithm changes;
AutoMigrate/schema (#419/#420); `data.json` migration (#424); authz-layer semantics themselves (task 03 — consume,
don't re-port).

## Validation

- [ ] `cd src && go build ./... && go vet ./data/... ./database/... ./serviceprovider/...`
- [ ] `cd src && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./database/parity -run 'DomainSuite/api_keys' -count=1` (suite name per task 03 harness)
- [ ] `cd src && go test ./... -count=1`
- [ ] Parity harness opt-in: internal/external visibility matrix (root ctx, owner ctx, other-user ctx, no-auth ctx)
- [ ] `grep -c "Transaction(" src/database/api_key.go` → 1 (D-9 `CreateApiKey`)
- [ ] `grep -n "g.legacy" src/database/api_key.go` → zero hits
- [ ] `grep -n "GetOwnRecords" src/database/api_key.go` → present on `GetApiKeys` (and nowhere else)

## Rollback

Revert the PR: `delegate_api_key.go` returns as task 04 wrote it, domain/tests/parity files removed. No schema
change (tables belong to #419/#420) ⇒ nothing to migrate back. Until re-landed, key lookups ride the legacy
engine through the delegate — the hybrid design makes this a no-risk revert (D-1).

## PBI acceptance mapping

| AC | How this task moves it |
|---|---|
| All CRUD via GORM verified against SQLite | 8 methods + per-caller parity matrices |
| Transactions | D-9 `CreateApiKey` (+ internal-key flow) tx site |
| Error handling (wrapped, consistent) | Mapping table on `mapGormError` (task 04), sentinel identity preserved |
| Legacy layer bypassed | Auth hot path `GetUserByApiKey` served from SQL with zero `g.legacy` |
