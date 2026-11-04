# Task 02 — Query-language port: filter / order / index semantics

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 01 · **PR size:** M
**Plan refs:** [D-4](../plan.md#d-4--query-language-port-exact-semantics-go-evaluation-first-gated-sql-push-down-later) (risk R-3's core mitigation)

## Context

Every `Get*(ctx, filter string)` in the 157-method surface runs the Mongo-flavoured language of
`src/data/filter_layer.go` (166 lines), `order_layer.go` (56), `index_layer.go` (36) — consumed by
`users.go`, `roles.go`, `claims.go`, `api_key.go`, `catalog.go`, `orchestrator.go`, `user_config.go`,
`packer_template.go`, `reverse_proxy.go`. The filter **value is an unanchored Go regexp** (`filter_layer.go:81`),
property paths resolve through json tags and nested maps by reflection, and `struct`-kind properties never match.
SQL `WHERE` cannot reproduce this wholesale (no REGEXP in SQLite without a UDF; Postgres `~` ≠ RE2). The port is
therefore a **transcription** into `src/database/query/`, evaluated Go-side over GORM-loaded slices — parity by
construction; push-down is a later, per-shape, differential-tested opt-in.

## Objective

`src/database/query` provides `ParseFilter`, `FilterByProperty`, `OrderByProperty`, `GetRecordIndex` (plus the
`Filter`/`FilterOptions`/`Order`/`OrderDirection` types) with **identical observable behaviour** to the legacy
layer functions, pinned by golden tests that run both implementations over shared fixtures.

## Files

| File | Op |
|---|---|
| `src/database/query/doc.go` | create — package doc incl. the D-4 no-push-down-yesterday rule |
| `src/database/query/filter.go` | create — transcription of `data/filter_layer.go:10-166` |
| `src/database/query/order.go` | create — transcription of `data/order_layer.go:8-55` |
| `src/database/query/index.go` | create — transcription of `data/index_layer.go:8-35` |
| `src/database/query/filter_test.go`, `order_test.go`, `index_test.go` | create — differential goldens |
| `src/database/query/testdata/fixtures.json` | create — shared fixture structs + filter corpus |

## File size guard

New files ≈ 170/60/40/250/120/120/300 lines — all far below 1,000 (ER-2). Legacy layer files are **not modified**
(`data` keeps its copies until #426 deprecation; note: `src/serviceprovider/parallelsdesktop/main.go` also calls
`data.ParseFilter` — untouched, out of feature scope).

## Implementation spec

Transcription rules (mechanical, per function):

1. Copy bodies verbatim; rename package-level unexported helpers only where needed (`getProperty`,
   `propertyIsValid`, `iterateMaps` → same names, same package `query`); keep the same generic signatures:
   ```go
   func ParseFilter(filter string) (*Filter, error)
   func FilterByProperty[T any](objects []T, filter *Filter) ([]T, error)
   func OrderByProperty[T any](objects []T, order *Order) ([]T, error)
   func GetRecordIndex[T any](objects []T, propertyIndexer, propertyValue string) (int, error)
   ```
   (legacy spells the constraint `interface{}`; `any` is the same constraint — cosmetic, keep repo style if it
   predates `any` usage; check neighbouring files.)
2. **Pointer-slice hardening** (only deliberate delta, must be listed in PR): legacy `getProperty` panics on
   pointer element types (`reflect.Ptr` has no `NumField`) — but `GetOrchestratorReverseProxyHosts`
   (`orchestrator.go:630-654`) returns `[]*models.ReverseProxyHost` and legacy works because it filters after
   dereferencing? Verify during implementation: add `value = reflect.Indirect(value)` at `getProperty` entry —
   the *behaviour* for value types is unchanged (`Indirect` is a no-op on struct values); add golden test
   `TestGetProperty_PointerElement` proving parity for value types.
3. Every quirk below gets a named golden test — these tests are the AC-parity proof reviewers check:

| Golden test | Pins |
|---|---|
| `TestParseFilter_Quirks` | `""`→nil,nil; `"a=b"`; `"a=b,i"`; `"a=b,I"` (EqualFold :41); `"a=b,x"` (unknown option ignored); `"a=b,i,i"` → double `(?i)` prefix still compiles (`filter_layer.go:77` loop mutates `propertyValue`); `"a=b=c"` → error (3 parts, :30); `"a"` → error |
| `TestFilterByProperty_RegexpNotEquality` | value `"e"` matches `"email"` (unanchored); `"^ro"` prefix; `"[0-9]$"`; invalid regex `"("` → error surfaced verbatim (:82) |
| `TestFilterCaseInsensitiveOption` | `name=x,i` matches `X` — and asserts the `(?i)`-prefix semantics, not `LOWER()` |
| `TestFilterNestedJsonTag` | `authentication.username` resolves via field name **and** json tag (`filter_layer.go:136-140,151-155`), case-insensitively |
| `TestFilterMapPaths` | `map[string]string` direct index; `map[string]int` scans keys EqualFold (:120-126) |
| `TestFilterNonStringKinds` | int/bool matched via `%v` formatting (`FailedLoginAttempts` value `"3"` matches — regexp on `"3"`); `[]string` matches against `"[a b]"` rendering (:86); struct-kind property never matches (:165); nil `*string` matches `"<nil>"` |
| `TestOrderByProperty_TieOrder` | ties keep input order under identical input slices; only `String()`-able fields sort (`order_layer.go:37`); unknown property → all `""` → order preserved |
| `TestGetRecordIndex_DeadBranches` | string EqualFold first-match-wins; bool/numeric properties return **-1** even on textual match (`index_layer.go:22-28` — legacy dead code, preserve) |
| `TestParityVsLegacy_*` | table-driven: same fixture JSON + same filter/order/index corpus executed through `data.ParseFilter`/`data.FilterByProperty`… (test file imports `data` — **test-only import**, same exception as D-5) → `reflect.DeepEqual` outputs |

4. `fixtures.json` corpus must include at least one value of every persisted-model shape used later
   (user with `Roles []Role`, host with `Tags []string` + `Authentication *T`, job with `Steps`), so domain tasks
   reuse it in task 03's harness.

## Out of scope

Authorization layer + `GetOwnRecords` (task 03); the harness (task 03); any `database/<domain>.go` file;
any SQL push-down translation; changing `src/data` layer files.

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./... -count=1` (legacy suite untouched)
- [ ] All 9 golden tables green; PR description lists each quirk → test name mapping
- [ ] `grep -n "regexp\|reflect" src/database/query/*.go | grep -v _test` — implementation uses stdlib only, no new deps

## Rollback

Delete `src/database/query/` (nothing imports it until task 03+). Zero runtime surface.

## PBI acceptance mapping

Enables AC-1 "All CRUD operations" for every filtered `Get*`; the single hardest piece of R-3; pre-req for AC-6
(the bypass audit later proves domain code uses `query.*`, never `data.*` layer funcs).
