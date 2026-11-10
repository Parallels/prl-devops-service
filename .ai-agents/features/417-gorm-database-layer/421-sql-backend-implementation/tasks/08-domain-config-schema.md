# Task 08 — Domain port: configuration singleton + schema version

**PBI** [#421](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 04 · **PR size:** S
**Plan refs:** [D-1](../plan.md#d-1--hybrid-delegate-gormdatabase-adopts-the-recommended-architecture-refined), [D-7](../plan.md#d-7--lifecycle-semantics-in-sql-mode-escape-hatch-table) (`SchemaVersion`/`UpdateSchemaVersion` row), [D-9](../plan.md#d-9--transaction-inventory-ac-transactions), [D-11](../plan.md#d-11--error-mapping-ac-error-handling); `schema_metadata` table naming per [#419 mapping doc](../../419-gorm-model-tags/plan.md) OQ-3 (D-12)

## Context

The smallest domain in the chain and the only one with **singleton semantics**: `Configuration` is one row, and
`Id` is the installation's identity — seeded once on first boot (`SeedId`) and then stable forever. A bug here is
not a wrong list; it is a fresh instance identity or a re-run of schema migrators. It depends only on task 04
(the `GormDatabase` shell + `mapGormError`), so it is parallelisable with 05/06/10/16 (dependency note under the
plan's task index).

## Objective

`src/database/config.go` + `src/database/schema.go` implement the 4 + 2 methods on GORM with strict single-row
semantics; `delegate_config.go` / `delegate_schema.go` (created by task 04) are **deleted**; parity `DomainSuite`
"config_schema" green.

## Files

| File | Op |
|---|---|
| `src/database/config.go` | create (~110 lines) |
| `src/database/schema.go` | create (~60) |
| `src/database/delegate_config.go`, `delegate_schema.go` | delete (as created by task 04) |
| `src/database/config_test.go`, `parity/suite_config_schema_test.go` | create |

## File size guard

Actuals: legacy `src/data/config.go` = **82** lines, `src/data/schema.go` = **23** lines (**both untouched** —
#426 territory). Projections (new files, not yet written): `database/config.go` ~110, `database/schema.go` ~60,
`config_test.go` ~90, `parity/suite_config_schema_test.go` ~80. Far under the 500-line ER-2 cap; no splits.

## Implementation spec

**Method inventory (implement exactly these; source `src/data/`):**

| # | Method | Legacy | SQL shape / notes |
|---|---|---|---|
| 1 | `GetConfiguration(ctx)` | config.go:12 | Singleton read: `First`/`Take` of the one `configurations` row ordered by rowid; **no filter, no authz, no order layer** — legacy takes the stored struct. Missing row ⇒ legacy's lazy-create-or-error behaviour (**implementer to confirm at task start: :12-25**) mirrored exactly; if legacy creates on read, do the same (seed-on-read), do not "fix" it into an error |
| 2 | `GetId(ctx)` | :26 | Read `Id` off the singleton; empty-string vs error distinction matters to callers — mirror whichever legacy returns when unset |
| 3 | `SeedId(ctx)` | :40 | Idempotent: return the existing `Id` if present, else generate (`generateId`, `:75`) and persist. Generation must stay out of any tx (pure CPU) — generate, then a conditional single-row write |
| 4 | `SetId(ctx, id)` | :57 | Overwrite `Id` on the singleton; validation/normalisation from the body while porting (`:57-74`) |
| 5 | `GetSchemaVersion(ctx)` | schema.go:7 | `SELECT value FROM schema_metadata WHERE id = <the single row>`; empty table ⇒ legacy default (JSON `Data.Schema` zero value — the `""`/initial-version case, D-7 row) must be returned identically |
| 6 | `UpdateSchemaVersion(ctx, version)` | :15 | Upsert of the single `schema_metadata` row (see lifecycle below) |

**Single-row semantics (the core of this task).** No method here accepts a filter or an id. Read with an explicit
`Take`/`Limit(1)`; write by the existing row's primary key, never `Save` of a zero-id struct (that inserts a
second row). Read → if none, create with the id pinned to the #419-mapped constant; if ≥2 rows exist (corruption),
take the first by rowid and log once — legacy physically could not have two. Assert the invariant in
`config_test.go`: `SELECT COUNT(*) FROM configurations` → 1 after three `SeedId` calls.

**Id seeding.** Port `generateId` (`config.go:75`) *call-wise*: call the `helpers`/`security` primitive it uses,
or raise a same-PR **pure move** of the generator into `helpers` — do not copy-paste it (two copies of an id
generator is how identity collisions start). Exact strategy **implementer to confirm at task start: :75-82**.
Two boots racing is closed by the pinned primary key, not by a tx: on a unique conflict, re-read and return the
stored value.

**Schema-version row (D-7 last row, AD-4).** `SchemaVersion`/`UpdateSchemaVersion` map to the **single-row
`schema_metadata` table**, per D-7: "single-row `schema_metadata` table (mirrors #419's OQ-3 default naming)".
Keep it **distinct** from #424's data-migration completion marker: that is a *separate* table (AD-4) — do not
reuse, alias, or "co-locate" it, and do not write the marker from here. Legacy schema migrators keep running
**only in legacy mode** (task 18's gate); this task only stores/returns the string.

**Query approach.** None of these six methods touches `query.ParseFilter`, `order_layer`, `index_layer`,
`authorize_layer`, or `GetOwnRecords` — no list, no filter, no authz in legacy. State that in the file's package
comment so a reviewer does not go looking for the wiring.

**Atomicity (D-9).** No tx row in D-9 covers this domain ⇒ it needs **no `db.Transaction`**: every method is one
read or one targeted write on one row (legacy held `dataMutex` for the same reason). `SeedId`'s
read-then-write race is closed by the pinned primary key, not by a tx. Record "no tx sites" in the PR description
against D-9.

**Error mapping (D-11, via `mapGormError`):**

| GORM outcome | Sentinel returned |
|---|---|
| `gorm.ErrRecordNotFound` (missing singleton row) | the legacy behaviour for "no configuration yet" — **not** necessarily `Err*NotFound`; mirror the body (seed-on-read or a typed empty result). If legacy does error, map to the relocated `interfaces.Err*` for this domain |
| unique/PK violation on the pinned singleton/`schema_metadata` row | swallow + re-read (idempotent seeding), **not** `Err*AlreadyExists` |
| any other error | `errors.NewWithCode(fmt.Sprintf("…: %v", err), 500)` |
| `!IsConnected()` | `interfaces.ErrDatabaseNotConnected` first guard in every method (mirrors legacy `if !j.IsConnected()`) |

**Parity test sketch (`src/database/parity/suite_config_schema_test.go`):**

```go
func TestConfigSchemaParity(t *testing.T) {
	legacy := parity.NewLegacyEngine(t)
	sql := parity.NewSQLEngine(t, false) // :memory: via the real factory
	ctx := parity.TestCtx(t)             // helper from task 03

	wantCfg, wErr := legacy.GetConfiguration(ctx)
	gotCfg, gErr := sql.GetConfiguration(ctx)
	if (wErr == nil) != (gErr == nil) { // seed-on-read vs error on a cold engine
		t.Fatalf("empty-DB GetConfiguration: legacy=%v sql=%v", wErr, gErr)
	}
	if wantCfg != nil && gotCfg == nil { t.Fatalf("GetConfiguration: sql returned nil, legacy did not") }

	// Id: seeded → stable → readable → overwritable, identical call sequence on both engines
	for _, db := range []interfaces.DatabaseService{legacy, sql} {
		first, err := db.SeedId(ctx)
		if err != nil || first == "" { t.Fatalf("SeedId: %q %v", first, err) }
		if again, _ := db.SeedId(ctx); again != first { t.Fatalf("SeedId not idempotent: %q -> %q", first, again) }
		if got, _ := db.GetId(ctx); got != first { t.Fatalf("GetId=%q want %q", got, first) }
		if got, _ := db.SetId(ctx, "set-by-test"); got != "set-by-test" { t.Fatalf("SetId=%q", got) }
		for _, v := range []string{"0", "1", "2024-01-01"} { // schema-version round-trip
			if err := db.UpdateSchemaVersion(ctx, v); err != nil { t.Fatal(err) }
			if got, _ := db.GetSchemaVersion(ctx); got != v { t.Fatalf("version %q != %q", got, v) }
		}
	}
}
```

**Edge cases to pin (unit + suite):**
- **fresh engine** (empty `:memory:` DB): `GetId` / `GetSchemaVersion` on zero rows — this is the first-boot path
  #424 and the seeds rely on; assert legacy-vs-SQL agree on value **and** error-ness.
- idempotence: `SeedId` ×3 ⇒ one row, same value (the row-count assertion above).
- not-found: `GetSchemaVersion` with an empty `schema_metadata` table ⇒ the same default string as legacy's
  zero-valued `Data.Schema`, not `""`-vs-error drift.
- dup-row: hand-insert a second `configurations` row in `config_test.go` ⇒ `GetConfiguration` still returns the
  first by rowid and the app does not panic (defensive path only; no user-facing API reaches it).
- `UpdateSchemaVersion` is monotonic-agnostic: it stores whatever string it is given (legacy does) — assert no
  validation was invented.

## Out of scope

Provider flip / `DB_TYPE` selection and the legacy-mode migrator gate (17/18); the #424 completion-marker table
(AD-4, separate table, owned by #424); AutoMigrate + table/column design (#419/#420); seeds retyping (task 17);
list/filter domains and the `query` helpers (02/03); other domains and their delegates.

## Validation

- [ ] `cd src && go build ./... && go vet ./data/... ./database/... ./serviceprovider/...`
- [ ] `cd src && go test ./database/... -count=1 -race`
- [ ] `cd src && go test ./database/parity -run 'DomainSuite/config_schema' -count=1` (suite name per task 03 harness)
- [ ] `cd src && go test ./... -count=1`
- [ ] Parity harness opt-in: run the same suite with a **cold** SQL engine (fresh `:memory:`) as well as a seeded one
- [ ] `grep -n "schema_metadata" src/database/schema.go` → present; `grep -rn "migration_marker\|completion" src/database/` → **zero** (AD-4: #424's table is not ours)
- [ ] `grep -c "Transaction(" src/database/{config,schema}.go` → 0 (state it in the PR against D-9)
- [ ] `grep -n "g.legacy" src/database/config.go src/database/schema.go` → zero hits
- [ ] Row-count invariant test green after `SeedId` ×3 and after concurrent `SeedId` (race detector)

## Rollback

Revert the PR: both delegate files return as task 04 wrote them; domain, unit and parity files removed. No schema
change (`schema_metadata` and the singleton table are #419/#420-owned) ⇒ no data step; values simply keep living
in `data.json` under legacy mode. Nothing downstream depends on 08, so the revert can sit open until task 17.

## PBI acceptance mapping

| AC | How this task moves it |
|---|---|
| All CRUD via GORM verified against SQLite | 6 methods incl. first-boot + idempotence paths |
| Transactions | Explicit "no tx needed" finding recorded against D-9 (single-row reads/targeted writes) |
| Error handling (wrapped, consistent) | Mapping table on `mapGormError`; empty-vs-error parity asserted, not assumed |
| Legacy layer bypassed | Config/`Id`/schema-version reads served from SQL with zero `g.legacy`; AD-4 marker separation proven by grep |

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-08`.** Cut `slice/421-08` from `sql-migration` HEAD
(`.worktrees/slice-421-08`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
