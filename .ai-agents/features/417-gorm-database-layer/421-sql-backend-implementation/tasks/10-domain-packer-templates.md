# Task 10 — Domain port: packer templates

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 03 · **PR size:** S
**Plan refs:** [D-1](../plan.md#d-1--hybrid-delegate-gormdatabase-adopts-the-recommended-architecture-refined), [D-4](../plan.md#d-4--query-language-port-exact-semantics-go-evaluation-first-gated-sql-push-down-later) (regexp + authz rows), [D-9](../plan.md#d-9--transaction-inventory-ac-transactions), [D-11](../plan.md#d-11--error-mapping-ac-error-handling); tables per [#419 mapping doc](../../419-gorm-model-tags/plan.md) (D-12)

## Context

The cleanest domain to port and therefore the best early proof that the `database/query` helpers are faithful:
five methods, no join tables, no cross-domain reads, and the filter is *only* the Mongo-flavoured query language —
an unanchored Go regexp, not equality (D-4's first table row). It depends on task 03 only (filter + authz port +
the parity harness), so it is an independent leaf that can run in parallel with 05/06/08/16 (plan's dependency
note under the task index).

## Objective

`src/database/packer_template.go` implements the 5 packer-template methods on GORM using the `database/query`
helpers; `delegate_packer_template.go` (created by task 04) is **deleted**; parity `DomainSuite`
"packer_templates" green, including the golden filter strings from task 02.

## Files

| File | Op |
|---|---|
| `src/database/packer_template.go` | create (~120 lines) |
| `src/database/delegate_packer_template.go` | delete (as created by task 04) |
| `src/database/packer_template_test.go` | create |
| `src/database/parity/suite_packer_templates_test.go` | create |

## File size guard

Actuals: legacy `src/data/packer_template.go` = **130** lines (**untouched** — #426 territory). Projections (new
files, not yet written): `database/packer_template.go` ~120, `packer_template_test.go` ~100,
`parity/suite_packer_templates_test.go` ~100. Under the 500-line ER-2 cap; no split.

## Implementation spec

**Method inventory (implement exactly these; source `src/data/`):**

| # | Method | Legacy | SQL shape / notes |
|---|---|---|---|
| 1 | `GetPackerTemplates(ctx, filter)` | packer_template.go:19 | `Find` all (rowid/insertion order) → `query.ParseFilter` → `query.FilterByProperty` → `query.GetAuthorizedRecords` (task 05's note lists **packer** among the authz-wrapped domains — confirm the wrap at :19-38 and mirror; **no** `OrderByProperty`: order is applied at only 3 sites, none here, D-4 grep-verified row) |
| 2 | `GetPackerTemplate(ctx, nameOrId)` | :39 | One string addressed as name **or** id: `WHERE id = ? OR name = ?` candidate set + Go-side `EqualFold` fallback (D-4 ASCII fast path); empty `nameOrId` ⇒ legacy empty-input sentinel before any query (confirm exact sentinel at :39-57) |
| 3 | `AddPackerTemplate(ctx, template)` | :58 | Pre-check duplicate id / duplicate name from `:58-88`, timestamps via `helpers.GetUtcCurrentDateTime()`, single-row insert. Returns `*models.PackerTemplate` |
| 4 | `DeletePackerTemplate(ctx, nameOrId)` | :89 | Resolve by name-or-id (same predicate as #2) → not-found sentinel → single-row delete (no cascade: templates are leaf records — confirm no child rows exist in #419's mapping; **implementer to confirm at task start: :89-112**) |
| 5 | `UpdatePackerTemplate(ctx, template)` | :113 | Load → legacy field-assignment set from `:113-130` + `UpdatedAt` refresh; avoid GORM `Save(struct)` if legacy copies a subset (partial-update drift is the classic port bug here) |

**Query approach — this task is the D-4 showcase.** Every list read goes through the ported helpers in
`database/query` (tasks 02/03); **do not** translate a filter into SQL. Concretely, the semantics that must survive
unchanged (D-4 table, each pinned by task 02's golden tests — cite, don't re-derive):
- value is an **unanchored** compiled regexp matched with `MatchString` (`filter_layer.go:81-86`) ⇒ `name=ubu`
  matches `ubuntu-2204`; a `LIKE 'ubu'` push-down would not;
- the `i` option double-prefixes `(?i)(?i)` (`:73-78`) — legal, must not be "fixed";
- malformed filters error on exactly 2 `=`-split parts, unknown options after `,` silently ignored (`:27-56`);
- property names resolve case-insensitively by Go field name **or** json tag, dotted paths recurse
  (`:109-161`) — so `GetPackerTemplates` must accept whatever nested/json-tag spellings the model exposes;
- non-string scalars compare through `fmt.Sprintf("%v", …)` and struct-kind properties never match (`:86,163-166`).

Push-down is **opt-in and only after the suite is green**, restricted to `WHERE id = ?` / `WHERE name = ?`
lookups (#2/#4/#5 resolution) because generated ids are lowercase (`helpers.GenerateId`, D-4 push-down note).
Never push regexp semantics into SQLite/Postgres (D-4: no REGEXP UDF decision is taken here; Postgres `~` differs
from RE2). If #419 mapped any packer-template field to a serialised-JSON column, filters on that path are
Go-side only (D-12).

**Atomicity.** **No tx sites.** D-9's inventory lists no packer-template flow ⇒ each method is one read or one
targeted write on one row (legacy held one `dataMutex.Lock()` for the same span). Write "0 tx sites, checked
against D-9" in the PR description; add a tx only if #4 turns out to touch a second table.

**Error mapping (D-11, via `mapGormError` from task 04):**

| GORM outcome | Sentinel returned |
|---|---|
| `gorm.ErrRecordNotFound` | `interfaces.ErrPackerTemplateNotFound` (relocated var per task 01/D-3 — identity, not text) |
| unique violation on id / name | the legacy duplicate sentinel (`ErrPackerTemplateAlreadyExists`-family), only where #419 indexed the column; otherwise the `:58-88` pre-check is the whole story |
| empty `nameOrId` | the legacy empty-input sentinel, before any query |
| any other error | `errors.NewWithCode(fmt.Sprintf("…: %v", err), 500)` |
| `!IsConnected()` | `interfaces.ErrDatabaseNotConnected` first guard in every method |

**Parity test sketch (`src/database/parity/suite_packer_templates_test.go`):**

```go
func TestPackerTemplatesParity(t *testing.T) {
	legacy := parity.NewLegacyEngine(t)
	sql := parity.NewSQLEngine(t, false) // :memory: via the real factory
	ctx := parity.TestCtx(t)             // helper from task 03

	parity.SeedBoth(t, func(db interfaces.DatabaseService) {
		_, _ = db.AddPackerTemplate(ctx, seedTpl("ubuntu-2204"))
		_, _ = db.AddPackerTemplate(ctx, seedTpl("Windows-2022"))
		_, _ = db.AddPackerTemplate(ctx, seedTplDisabled("fedora-40"))
	})

	filters := []string{"", "name=ubu", "name=^win", "name=(?i)nia$", "id=nope", "bad~filter"}
	for _, f := range filters {
		want, wErr := legacy.GetPackerTemplates(ctx, f)
		got, gErr := sql.GetPackerTemplates(ctx, f)
		parity.AssertSameList(t, "GetPackerTemplates["+f+"]", want, got) // DeepEqual: unanchored + (?i)(?i) quirks included
		if (wErr == nil) != (gErr == nil) {
			t.Fatalf("filter %q error-ness differs: legacy=%v sql=%v", f, wErr, gErr)
		}
	}
	// mutation parity: identical call order on both engines, re-list after each
	for _, db := range []interfaces.DatabaseService{legacy, sql} {
		_, _ = db.UpdatePackerTemplate(ctx, edited("ubuntu-2204"))
		_ = db.DeletePackerTemplate(ctx, "fedora-40")
	}
	parity.AssertSameList(t, "after-mutations", must(legacy.GetPackerTemplates(ctx, "")),
		must(sql.GetPackerTemplates(ctx, "")))
}
```

`name=ubu` (unanchored hit), `name=^win` (anchored-by-the-caller miss on `Windows-2022` unless the `i` path
applies) and `name=(?i)nia$` are the three shapes that fail a naive `LIKE` port — keep all three.

**Edge cases to pin (unit + suite):**
- empty collection: zero templates ⇒ `[]models.PackerTemplate{}` vs `nil` must match legacy exactly (assert it —
  `reflect.DeepEqual` distinguishes the two and callers `len()`-check them).
- not-found: `GetPackerTemplate("nope")`, `DeletePackerTemplate("nope")`, `UpdatePackerTemplate` on a missing
  record, and empty `nameOrId` ⇒ four sentinel-identity assertions via `errors.Is`.
- duplicate id/name: `AddPackerTemplate` twice by id, then twice by name ⇒ the duplicate sentinel both times and
  `COUNT(*)` = 1.
- invalid filter syntax (`bad~filter`, one `=`-part, `name=a=b`) ⇒ same error, same code, as legacy — the
  "silently ignored options" branch too (`name=x,j`).
- no `OrderByProperty`: assert the suite's expected order is insertion/rowid, not sorted.

## Out of scope

Provider flip / `DB_TYPE` selection (17/18); the `query` layer itself (tasks 02/03 — consume, don't re-port or
"improve"); users (06), roles/claims (05), api-keys (07), config/schema (08), user-configs (09), catalog (11),
orchestrator/jobs/reverse-proxy (12–15) and their delegates; AutoMigrate/column design (#419/#420); `data.json`
migration (#424); the build/packer *service* layer that consumes templates (`src/pdfile`, controllers — untouched).

## Validation

- [ ] `cd src && go build ./... && go vet ./data/... ./database/... ./serviceprovider/...`
- [ ] `cd src && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./database/parity -run 'DomainSuite/packer_templates' -count=1` (suite name per task 03 harness)
- [ ] `cd src && go test ./... -count=1`
- [ ] Parity harness opt-in: replay task 02's golden filter corpus against `GetPackerTemplates` (the corpus, not a
      hand-picked subset) — this is the differential test D-4 requires before any push-down
- [ ] `grep -c "Transaction(" src/database/packer_template.go` → 0 (matches D-9)
- [ ] `grep -n "LIKE\|Regexp\|regexp" src/database/packer_template.go` → zero (no regexp pushed into SQL, D-4)
- [ ] `grep -n "g.legacy" src/database/packer_template.go` → zero hits

## Rollback

Revert the PR: `delegate_packer_template.go` returns as task 04 wrote it (stubs are deleted, never edited — D-1
risk row); domain, unit and parity files removed. No schema change ⇒ no data step. 10 is an independent leaf, so
the chain (05→06→…) proceeds unaffected while it is out.

## PBI acceptance mapping

| AC | How this task moves it |
|---|---|
| All CRUD via GORM verified against SQLite | 5 methods + full golden-filter corpus over SQL |
| Transactions | "No tx needed" verified against D-9 and recorded |
| Error handling (wrapped, consistent) | Mapping table on `mapGormError`; invalid-filter parity pinned |
| Legacy layer bypassed | List + lookups served from SQL through the ported query layer, zero `g.legacy` |

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-10`.** Cut `slice/421-10` from `sql-migration` HEAD
(`.worktrees/slice-421-10`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
