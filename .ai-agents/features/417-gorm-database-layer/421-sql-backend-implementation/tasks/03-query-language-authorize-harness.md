# Task 03 — Query-language port: authorization layer + differential harness

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 02 · **PR size:** M
**Plan refs:** [D-4](../plan.md#d-4--query-language-port-exact-semantics-go-evaluation-first-gated-sql-push-down-later), [D-5](../plan.md#d-5--differential-parity-harness-srcdatabaseparity) — R-3's second half and the gate every domain task reuses

## Context

`Get*` results are post-wrapped by record-level authorization (`authorize_layer.go:13-100`) and
owner-filtering (`GetOwnRecords`, `api_key.go:19-32`) inside legacy domain methods; the rules are
auth-context-driven (root/api-key/microservice bypass **suppressed when injected claims/roles exist** — the
:48-54 comment block is an explicit past security fix and must not be "simplified"), so they belong to the query
language, not to any domain. This task also builds `src/database/parity`, the two-engine fixture harness
(R-3 mitigation) that every domain task 05–16 plugs a `DomainSuite` into.

## Objective

`query.Authorized[T]`/`query.GetAuthorizedRecords`/`query.IsRootUser`/`query.OwnRecords` behave identically to
`data.IsAuthorized`/`data.GetAuthorizedRecords`/`data.IsRootUser`/`data.GetOwnRecords` across a matrix of
authorization contexts; `parity.NewLegacyEngine/NewSQLEngine/SeedBoth/AssertSameList` run both engines over one
seed function in unit tests.

## Files

| File | Op |
|---|---|
| `src/database/query/authz.go` | create — transcription of `data/authorize_layer.go:8-100` + `data/api_key.go:19-32` |
| `src/database/query/authz_test.go` | create — context-matrix goldens |
| `src/database/parity/doc.go` | create — documents the test-only `data` import exception (AD-1) |
| `src/database/parity/harness.go` | create (~150 lines) |
| `src/database/parity/harness_test.go` | create — self-proof (identity engine pair over a trivial domain seed) |
| `src/database/parity/testdata/` | create — shared seed corpora (reuses task 02 fixtures) |

## File size guard

New files ≈ 110/260/25/160/120 lines. `data/authorize_layer.go` and `data/api_key.go` untouched. Verdict: **ok**.

## Implementation spec

### `authz.go` (transcription, zero semantic drift)

```go
func IsRootUser(ctx basecontext.ApiContext) bool                    // verbatim :24-37
func IsAuthorized[T AuthorizedRecord](ctx basecontext.ApiContext, t T) bool   // verbatim :39-100
func GetAuthorizedRecords[T AuthorizedRecord](ctx basecontext.ApiContext, t ...T) []T // :13-22
func OwnRecords[T interface{ GetUserID() string }](ctx basecontext.ApiContext, records ...T) []T // GetOwnRecords verbatim
type AuthorizedRecord interface { GetRequiredClaims() []string; GetRequiredRoles() []string } // = data.AuthorizedRecord
```

Golden context matrix (fake `ApiContext` — implement a recording fake or mutate
`basecontext`'s `AuthorizationContext` via the setters the middleware uses; a fake context type ~60 lines is
acceptable and also serves domain tests):

| # | AuthorizationContext state | Expect (record w/ 1 required role + 1 required claim, neither held) |
|---|---|---|
| 1 | nil ctx | false |
| 2 | `AuthorizedBy:"RootAuthorization"` | true (:51) |
| 3 | `AuthorizedBy:"ApiKeyAuthorization"`, `IsMicroService` | true |
| 4 | same as 2/3 **plus** `InjectedClaims=["x"]` | filtering applies — false (:48-54) |
| 5 | `IsSuperUser` | true (:66) |
| 6 | `HasEffectiveRole(SUPER_USER_ROLE)`, no injected | true |
| 7 | `IsAuthorized:false` | false |
| 8 | user w/ role but not claim | false (AND-of-groups :95) |
| 9 | record w/ empty required roles + claims | true (public :75-76,85-86) |
| 10 | `constants.SUPER_USER_ROLE` constant used verbatim (no string literal) | — |

Plus `OwnRecords`: nil auth ctx / nil user ⇒ passthrough **same slice** (`api_key.go:21-24` returns `records`,
not a copy — pin it); matching `GetUserID() == authContext.User.ID` is case-**sensitive** `==` (:28) — pin.

### `parity/harness.go` (D-5 contract)

```go
type DomainSuite struct {
    Name string
    Seed func(t testing.TB, db interfaces.DatabaseService)      // exported-API-only seeding
    Checks []func(t testing.TB, legacy interfaces.DatabaseService, sql interfaces.DatabaseService)
}
func NewLegacyEngine(t testing.TB) *data.JsonDatabase
// temp dir via t.TempDir(); data.ResetForTesting() (data/testing_reset.go:7) BEFORE
// data.NewJsonDatabase(basecontext ctx, filepath.Join(dir,"data.json")); t.Cleanup(ResetForTesting).
// NOTE (implementer must verify): NewJsonDatabase is a global-singleton factory (data/main.go:76-78) —
// legacy engines cannot run in parallel within one process; harness serialises them (mutex) and
// tests must not t.Parallel().
func NewSQLEngine(t testing.TB, legacyFallback *data.JsonDatabase) *database.GormDatabase
// database.NewDatabaseConnection(ctx, DatabaseConfig{Type:"sqlite", Path:":memory:"}) — real factory (#418),
// AutoMigrate over the #419 model list helper (database/migrate.go — if #419 ships one; else local list).
func SeedBoth(t testing.TB, seed func(db interfaces.DatabaseService), legacy *data.JsonDatabase, sql *database.GormDatabase)
func AssertSameList[E any](t testing.TB, what string, want, got []E)  // DeepEqual + length/content diff print
func RunMatrix(t testing.TB, suites ...DomainSuite)
```

Self-proof test this PR: `TestHarness_SelfProof` — seeds 3 users through the *legacy* engine and, for the SQL
side, `legacyFallback` delegate (task 04 not yet merged, so this task defines `database.GormDatabase` **only if
needed for compile**: if constructor not yet available, the harness takes the SQL side as
`interfaces.DatabaseService` and self-proof runs legacy-vs-legacy with a mutation proving DeepEqual actually
catches drift). **Preferred:** land task 04 first if sequencing allows; harness signature stays identical either
way (takes interfaces, not concrete types).

## Out of scope

Any domain port; SQL push-down; changing legacy authorize code; performance work.

## Validation

- [ ] `cd src && go build ./... && go vet ./database/...`
- [ ] `cd src && go test ./database/... -count=1 -race` (harness mutex under `-race`)
- [ ] `cd src && go test ./... -count=1`
- [ ] Context matrix table above green (10 rows, both engines' funcs side-by-side in one test via `data.IsAuthorized` vs `query.IsAuthorized`)
- [ ] `grep -rn "SUPER_USER_ROLE" src/database/query/*.go` — constant, no literal `"super_user"` string (verify constant's actual value in `src/constants`)

## Rollback

Delete `query/authz.go`, `parity/` — no production importers.

## PBI acceptance mapping

Foundation slice of AC-1 for every authorized `Get*`; builds the R-3 differential harness that every later task's
parity proof (and AC-6's per-domain evidence) depends on.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-03`.** Cut `slice/421-03` from `sql-migration` HEAD
(`.worktrees/slice-421-03`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
