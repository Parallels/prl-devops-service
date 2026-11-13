# Task 11 — Domain port: catalog manifests + catalog managers

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 06 · **PR size:** L
**Plan refs:** [D-1](../plan.md#d-1--hybrid-delegate-gormdatabase-adopts-the-recommended-architecture-refined), [D-4](../plan.md#d-4--query-language-port-exact-semantics-go-evaluation-first-gated-sql-push-down-later) (order + authz rows), [D-7](../plan.md#d-7--lifecycle-semantics-in-sql-mode-escape-hatch-table) (`SaveNow` row), [D-9](../plan.md#d-9--transaction-inventory-ac-transactions) (`catalog.go:216-320` row), [D-11](../plan.md#d-11--error-mapping-ac-error-handling), [D-12](../plan.md#d-12--schema-ownership-reference-dont-re-decide); `Tags`/`RequiredRoles`/`RequiredClaims` array columns per [#419 mapping doc](../../419-gorm-model-tags/plan.md) (`serializer:json`)

## Context

The biggest domain in the chain — **30 methods** — and the only one that exercises all four query layers at once:
`GetCatalogManifests` is the ordering call site D-4 pins at `catalog.go:35` (`UpdatedAt desc`), it is
authz-wrapped (`GetAuthorizedRecords` at `catalog.go:33`), and it filters over **string-array** fields (`Tags`,
`RequiredRoles`, `RequiredClaims`) that #419 maps to serialised-JSON columns, so those filters are Go-side only
(D-12). Catalog also carries the taint/revoke security flows: a wrong port here leaks a revoked image version.

## Objective

`src/database/catalog.go` (+ optional split) and `src/database/catalog_managers.go` implement all 30 methods on
GORM; `delegate_catalog.go` / `delegate_catalog_managers.go` (task 04) are **deleted**; parity `DomainSuite`
"catalog" green including the ordering and array-filter matrices; 3 tx sites matching D-9.

## Files

| File | Op |
|---|---|
| `src/database/catalog.go` | create (~320 lines: reads, lifecycle, taint/revoke) |
| `src/database/catalog_arrays.go` | create (~150: the 8 tags/claims/roles mutators) — split only if `catalog.go` would exceed ~350 (ER-2; same convention task 12 uses) |
| `src/database/catalog_managers.go` | create (~90) |
| `src/database/delegate_catalog.go`, `delegate_catalog_managers.go` | delete (as created by task 04) |
| `src/database/catalog_test.go`, `catalog_managers_test.go`, `parity/suite_catalog_test.go` | create |

## File size guard

Actuals: legacy `src/data/catalog.go` = **714** lines (already over the 500-line ER-2 cap — **untouched**; its
split is #426/ER-3 work, not this PR), `src/data/catalog_managers.go` = **99** lines (untouched). Projections
(new files, not yet written): `database/catalog.go` ~320, `catalog_arrays.go` ~150, `catalog_managers.go` ~90,
`catalog_test.go` ~200, `parity/suite_catalog_test.go` ~180. If the manifest reads alone exceed 500, split reads
into `catalog_reads.go` rather than grow one file (D-1's 1:1 domain↔delegate rule survives: both delegates are
deleted by this one PR).

## Implementation spec

**Method inventory — manifests, `src/data/catalog.go` (25):**

| # | Method | Legacy | SQL shape / notes |
|---|---|---|---|
| 1 | `GetCatalogManifests(ctx, filter)` | catalog.go:18 | `Find` all (rowid order) → `query.FilterByProperty` → `GetAuthorizedRecords` (`:33`) → **order `UpdatedAt desc`** (`:35`) via the ported `order_layer` — see *Ordering* below |
| 2 | `GetCatalogManifestByName(ctx, idOrName)` | :43 | `WHERE id = ? OR name = ?` candidates + Go-side `EqualFold` (D-4 fast path); empty input ⇒ legacy empty-input sentinel first |
| 3 | `GetCatalogManifestByTag(ctx, catalogId, tag)` | :62 | Resolve catalog by id, then **tag inside the serialised `Tags` JSON column** ⇒ Go-side match after load (D-12: no SQL dependency on the #419 choice; no `LIKE '%tag%'` — substring false-positives) |
| 4 | `GetCatalogManifestsByCatalogId(ctx, catalogId)` | :85 | `WHERE catalog_id = ?` (push-down-safe equality) + authz wrap if the legacy body applies one — confirm at :85-106 |
| 5 | `GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)` | :107 | `WHERE catalog_id = ? AND version = ?` + Go-side `EqualFold` fallback for case semantics of the body |
| 6 | `GetCatalogManifestsByCatalogIdVersionAndArch(…)` | :134 | As #5 + `architecture` — returns a single manifest; not-found sentinel |
| 7 | `CreateCatalogManifest(ctx, manifest)` | :157 | Dup pre-checks + timestamps + insert; **D-9 has no row here ⇒ default single-row insert**; if #419 mapped nested contents (`CatalogManifest` contents, D-12) to child rows this becomes multi-table ⇒ tx + record against D-9. **implementer to confirm at task start: :157-215** |
| 8 | `DeleteCatalogManifest(ctx, catalogIdOrId)` | :216 | **tx (D-9 `catalog.go:216-320` cascade)**: resolve, guard, multi-row delete |
| 9 | `DeleteCatalogManifestVersion(ctx, catalogIdOrId, version)` | :261 | **tx**: delete all rows of that (catalog,version) |
| 10 | `DeleteCatalogManifestVersionArch(…, architecture)` | :291 | **tx**: narrowest cascade |
| 11 | `UpdateCatalogManifest(ctx, record)` | :321 | Load → legacy assignment set from `:321-358` + `UpdatedAt` (this feeds the ordering in #1 — a missed `UpdatedAt` changes list order silently) |
| 12 | `UpdateCatalogManifestTags(ctx, record)` | :359 | Replace the whole `Tags` array — serialised-JSON column write |
| 13 | `UpdateCatalogManifestRequiredRoles(ctx, recordId, roles...)` | :375 | Replace array; validation/dedupe from `:375-402` |
| 14 | `AddCatalogManifestRequiredClaims(ctx, recordId, claims...)` | :403 | Read-modify-write of the array (append + dedupe per `:403-430`) |
| 15 | `RemoveCatalogManifestRequiredClaims(…)` | :431 | Read-modify-write (`:431-457`) |
| 16 | `AddCatalogManifestRequiredRoles(…)` | :458 | Read-modify-write (`:458-485`) |
| 17 | `RemoveCatalogManifestRequiredRoles(…)` | :486 | Read-modify-write (`:486-513`) |
| 18 | `AddCatalogManifestTags(ctx, recordId, tags...)` | :514 | Read-modify-write (`:514-541`) |
| 19 | `RemoveCatalogManifestTags(…)` | :542 | Read-modify-write (`:542-569`) |
| 20 | `UpdateCatalogManifestDownloadCount(ctx, catalogId, version)` | :570 | Addressed by catalogId+version (not id); targeted `UPDATE` of the counter — keep it non-transactional and non-atomic **as legacy is** (a lost update on a counter is the legacy behaviour; do not invent `UPDATE … = x+1` SQL, port the read-modify-write) |
| 21 | `TaintCatalogManifestVersion(ctx, catalogId, version)` | :594 | **Taint flow**: targeted flag write + return the updated manifest; flag name/semantics from `:594-617` |
| 22 | `UnTaintCatalogManifestVersion(ctx, catalogId, version)` | :618 | Inverse flag write (`:618-642`) |
| 23 | `RevokeCatalogManifestVersion(ctx, catalogId, version)` | :643 | **Revoke flow** (`:643-666`) — assert the returned manifest and the stored flag agree |
| 24 | `UpdateCatalogManifestMetadata(ctx, catalogId, version, architecture, description *string, tags, claims, roles []string)` | :667 | Multi-field targeted update; the `description *string` is a **nil-means-don't-touch** pointer — honour it (GORM `Select` of only the non-nil fields) |
| 25 | `UpdateCatalogManifestProvider(ctx, recordId, provider)` | :701 | Provider write; if #419 mapped provider as nested/JSON ⇒ one column write, else a child-row write ⇒ tx. **implementer to confirm at task start: :701-714 + #419 mapping** |

**Method inventory — managers, `src/data/catalog_managers.go` (5):**

| # | Method | Legacy | SQL shape / notes |
|---|---|---|---|
| 26 | `GetCatalogManagers()` | catalog_managers.go:16 | **No `ctx` parameter** ⇒ no not-connected guard and no authz: mirror exactly (still guard on connection state the way the body does — **implementer to confirm at task start: :16-22**). `Find` all, rowid order, no filter/order |
| 27 | `GetCatalogManager(id)` | :23 | No `ctx`; `WHERE id = ?` + `EqualFold` fallback; not-found sentinel |
| 28 | `AddCatalogManager(ctx, mgr)` | :36 | Dup pre-check (`:36-50`) + insert |
| 29 | `UpdateCatalogManager(ctx, mgr)` | :51 | Load → legacy assignment set (`:51-73`) |
| 30 | `DeleteCatalogManager(ctx, id)` | :74 | Guard (the managers list has a "last one"/internal notion? confirm at `:74-99`) + single-row delete |

**Ordering (D-4, `catalog.go:35`).** Baseline: after filter + authz, sort **Go-side** with the ported
`query` order layer — `sort.Slice` on `.FieldByName(prop).String()`, non-string fields sort as `""`, ties keep
input order, unstable sort (`order_layer.go:32-55`). Load order is therefore **rowid**, never `ORDER BY` in SQL,
until push-down. Push-down (opt-in, after green parity): `ORDER BY updated_at DESC, rowid` — D-4 explicitly
allows the rowid secondary sort "to mimic legacy tie order"; ship it only with the differential test in the same
PR. This is the one domain where the *sort port itself* is load-bearing for parity.

**String-array columns.** `Tags` / `RequiredRoles` / `RequiredClaims` are `serializer:json` columns per #419
(D-12: "Where #419 chose serialised-JSON columns, filters/orderings on such paths are **evaluated Go-side
only**"). Consequences: (a) all seven array mutators (#12–#19) are load → mutate slice → `Select` that column and
`Update` (never `Save(struct)`, which would clobber concurrent `UpdatedAt`/counter writes); (b) filters on those
paths run through the ported Go evaluator, which already knows how `[]string` renders under
`fmt.Sprintf("%v", …)` (`filter_layer.go:86,163-166` — D-4 pins `[]string` matching against `"[a b]"`); (c) no
SQL `LIKE`/`json_each` shortcuts.

**Atomicity (D-9).** Exactly three `db.Transaction(func(tx *gorm.DB) error)` sites: #8, #9, #10 (the
`DeleteCatalogManifest*` cascades, D-9 evidence `catalog.go:216-320`, "multi-row deletes"). Everything else is a
single read or a single targeted write ⇒ **no tx** (D-9 rule: one tx per public method legacy performed under one
`dataMutex.Lock()`). #7/#25 upgrade to tx only if the #419 mapping puts their nested data in child tables — then
add the row to D-9 in the PR description, do not decide locally (D-12).

**`SaveNow` on the import path (D-7).** `src/catalog/import.go:204` calls `db.SaveNow(s.ctx)` after the manifest
is obtained/created. In SQL mode `SaveNow` is a **no-op returning nil** (D-7's lifecycle table — "all existing
`SaveNow` callers (… `catalog/import.go:204` …) keep compiling and become harmless") because GORM already
committed the row per statement. So: **no change to `import.go` in this PR** (its retyping to
`interfaces.DatabaseService` is task 17), and **no `SaveNow` call inside `database/catalog.go`** — there is nothing
to flush. Note this explicitly in the PR: a reviewer will look for it.

**Error mapping (D-11, via `mapGormError`):**

| GORM outcome | Sentinel returned |
|---|---|
| `gorm.ErrRecordNotFound` | `interfaces.ErrCatalogManifestNotFound` / `interfaces.ErrCatalogManagerNotFound` (names relocated by task 01; identity per D-3 — pick by the domain the method belongs to) |
| unique violation on the manifest key (#419's `(catalog_id, version, architecture)` or `id`) | the legacy duplicate sentinel (`ErrCatalogManifestAlreadyExists`-family) — pre-checks in `:157-215` stay primary |
| internal/protected-manifest guard tripped (#8–#10, #30) | the legacy guard sentinel **by identity**, returned before the cascade |
| any other error | `errors.NewWithCode(fmt.Sprintf("…: %v", err), 500)` |
| `!IsConnected()` | `interfaces.ErrDatabaseNotConnected` first guard in every method **that receives a ctx** (#26/#27 have none — mirror their bodies) |

**Parity test sketch (`src/database/parity/suite_catalog_test.go`):**

```go
func TestCatalogParity(t *testing.T) {
	legacy := parity.NewLegacyEngine(t)
	sql := parity.NewSQLEngine(t, false)
	ctx := parity.TestCtx(t) // helper from task 03

	parity.SeedBoth(t, func(db interfaces.DatabaseService) {
		for _, m := range []*models.CatalogManifest{
			seedManifest("win", "1.0", "amd64", taint(false), tags("stable", "prod")),
			seedManifest("win", "1.1", "amd64", taint(true), tags("beta")),  // same catalogId, 2 versions
			seedManifest("win", "1.0", "arm64", taint(false), tags()),       // same catalogId+version, 2 arches
			seedManifest("lin", "2.0", "amd64", taint(false), tags("prod")),
		} {
			_, _ = db.CreateCatalogManifest(ctx, *m)
		}
		_ = db.AddCatalogManager(ctx, seedManager("importer-a"))
	})

	for _, f := range []string{"", "name=win.*", "tags=.*prod.*", "requiredroles=.*admin.*", "tainted=true", "id=nope"} {
		want, wErr := legacy.GetCatalogManifests(ctx, f)
		got, gErr := sql.GetCatalogManifests(ctx, f)
		parity.AssertSameList(t, "GetCatalogManifests["+f+"]", want, got) // ORDER-SIGNIFICANT: UpdatedAt desc, ties = rowid
		if (wErr == nil) != (gErr == nil) { t.Fatalf("filter %q: %v vs %v", f, wErr, gErr) }
	}
	// taint/revoke flows: identical mutation order on both engines, then re-list
	for _, db := range []interfaces.DatabaseService{legacy, sql} {
		_, _ = db.TaintCatalogManifestVersion(ctx, "lin", "2.0")
		_, _ = db.UnTaintCatalogManifestVersion(ctx, "win", "1.1")
		_, _ = db.RevokeCatalogManifestVersion(ctx, "win", "1.0")
		_ = db.AddCatalogManifestTags(ctx, idOf("lin", "2.0", "amd64"), "revoked-manually")
		_ = db.RemoveCatalogManifestRequiredClaims(ctx, idOf("win", "1.1", "amd64"), "write-vm")
	}
	parity.AssertSameList(t, "after-mutations",
		must(legacy.GetCatalogManifests(ctx, "")), must(sql.GetCatalogManifests(ctx, "")))
	parity.AssertSameList(t, "managers", must(legacy.GetCatalogManagers()), must(sql.GetCatalogManagers()))
}
```

**Edge cases to pin (unit + suite):**
- empty collection: no manifests ⇒ nil-vs-empty asserted; `GetCatalogManifestsByCatalogId("nope")` ⇒ **empty
  slice**, not an error (callers `len()`-check it).
- not-found: by name, by tag, by (id,version), by (id,version,arch), managers by id, empty `idOrName`, and
  taint/revoke on an unknown version ⇒ sentinel identity via `errors.Is` in all seven cases.
- duplicate key: same `(catalogId, version, architecture)` twice ⇒ duplicate sentinel + `COUNT(*)` = 1; a
  **different** arch with the same version still succeeds (seed row 3 proves it).
- cascade: `DeleteCatalogManifestVersion` removes **both** arch rows and nothing else (row counts asserted); a
  forced mid-cascade failure leaves counts unchanged (task 05's trick).
- ordering: two manifests with an identical `UpdatedAt` ⇒ tie order = rowid on both engines (catches a stray SQL
  `ORDER BY`).
- arrays: `AddCatalogManifestTags` dedupes; removing an absent tag returns legacy's verdict; `nil` description in
  #24 leaves the stored description untouched (reload and compare).

## Out of scope

Provider flip / `DB_TYPE` selection (17/18) and any change to `src/catalog/*.go` incl. `import.go:204` (retyping is
task 17); splitting legacy `data/catalog.go` (714 lines — #426/ER-3); users (06), roles/claims (05), api-keys
(07), orchestrator hosts/VMs (13/14 — they *consume* manifests), reverse-proxy (12), jobs (15) and their
delegates; the query layer itself (02/03); AutoMigrate, the serializer choice, and any nested-contents table
design (#419/#420 — raise a mapping amendment instead); `data.json` → SQL import (#424).

## Validation

- [ ] `cd src && go build ./... && go vet ./data/... ./database/... ./serviceprovider/...`
- [ ] `cd src && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./database/parity -run 'DomainSuite/catalog' -count=1` (suite name per task 03 harness)
- [ ] `cd src && go test ./... -count=1`
- [ ] Parity harness opt-in: the ordering corpus (`UpdatedAt` ties, one null, same-timestamp batch) run against
      **both** engines before any `ORDER BY` push-down is attempted
- [ ] `grep -c "Transaction(" src/database/catalog*.go` → 3 (matches D-9 `catalog.go:216-320`)
- [ ] `grep -n "SaveNow" src/database/catalog*.go` → zero (D-7: no-op caller lives in `catalog/import.go`, not here)
- [ ] `grep -n "LIKE\|json_each" src/database/catalog*.go` → zero (array filters are Go-side, D-12)
- [ ] `grep -n "g.legacy" src/database/catalog.go src/database/catalog_managers.go src/database/catalog_arrays.go` → zero hits
- [ ] Count assertion: 30 methods implemented = the two delegates deleted (list them in the PR body)

## Rollback

Revert the PR: both delegate files return as task 04 wrote them (deleted, never edited — D-1 risk row); domain,
split, unit and parity files removed. No schema change (manifest/manager tables and the JSON array columns belong
to #419/#420) ⇒ no data step; `import.go:204` was never touched, so the import path is unaffected. Catalog
consumers (13/14) keep riding the legacy delegate through `GormDatabase` until this lands again. If only the
push-down experiment misbehaves, revert **that commit** — the Go-side ordering baseline stays valid on its own.

## PBI acceptance mapping

| AC | How this task moves it |
|---|---|
| All CRUD via GORM verified against SQLite | 30 methods (25 manifests + 5 managers) incl. the ordering + array-filter matrices |
| Transactions | D-9 `DeleteCatalogManifest*` cascades = 3 tx sites; #7/#25 outcomes recorded |
| Error handling (wrapped, consistent) | Mapping table on `mapGormError`, guard sentinels by identity before cascades |
| Legacy layer bypassed | Manifest reads/authoring + managers served from SQL, zero `g.legacy`; `SaveNow` no-op reliance documented (D-7) |

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-11`.** Cut `slice/421-11` from `sql-migration` HEAD
(`.worktrees/slice-421-11`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
