# Task 06 — AutoMigrate verification + schema snapshot + mapping-table finalisation

## Context
Final PR of #419. AC3 lives or dies here: `db.AutoMigrate(...)` against a **fresh SQLite file from
the #418 factory** (temp dir — no mocks), the schema is pinned as a golden snapshot, and the
mapping table in `plan.md` gets a closing gap-pass (02–05 may have forced tag adjustments; the table
is #424's verbatim migration spec, so it must match the shipped code exactly).

## Objective
`TestAutoMigrate_Fresh` creates the DB at `t.TempDir()` via
`database.NewDatabaseConnection(ctx, database.DefaultDatabaseConfig(ctx))` (ctx-taking form, #420 D-4),
AutoMigrates the 18-model list, asserts every table/column/index from the plan mapping via the GORM
`Migrator()` + raw `sqlite_master` reads, writes/compares a golden DDL snapshot, and closes the
mapping table (any drift → fix code or table, then re-run all #419 tests).

## Files
- Create: `src/data/models/automigrate_test.go` (`package models_test`),
  `testdata/schema_snapshot.sql` (golden DDL)
- Modify: `plan.md` mapping table (finalisation pass rows only), wave files **only** for gaps the
  pass uncovers (e.g. missing index tag) — each fix re-runs 02–05 tests
- Create: shared `allAutoMigrateModels()` helper (or promote it out of task-05 test into a small
  `gormtest` helper file so 05/06 share one source of truth)

## File size guard (ER-4)
| File | Now | After (proj.) | Verdict |
|---|---|---|---|
| `automigrate_test.go` | new | ~220 (proj.) | ok |
| `testdata/schema_snapshot.sql` | new | ~120 lines generated (proj.) | ok (testdata, ER-2 n/a) |
| `plan.md` | ~350 | +10 (AC5 evidence note) | ok |

## Implementation spec
```go
func TestAutoMigrate_Fresh(t *testing.T) {
	cfg := database.DefaultDatabaseConfig(ctx)              // ctx-taking form (post-#420)
	cfg.Path = filepath.Join(t.TempDir(), "prl-devops.db")  // DATABASE_FOLDER/home convention honoured by factory defaults
	db, err := database.NewDatabaseConnection(ctx, cfg)     // glebarez/sqlite dialect (#418, AD-2)
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(allAutoMigrateModels()...)) // 18 roots, plan §Mapping "AutoMigrate list"

	m := db.Migrator()
	for _, tbl := range []string{"users","roles","claims","role_claims","user_roles","user_claims",
		"api_keys","catalog_managers","catalog_manifests","jobs","packer_templates","user_configs",
		"orchestrator_enrollment_tokens","orchestrator_hosts","reverse_proxies","reverse_proxy_hosts",
		"vm_snapshots","virtual_machines","configurations","database_schemas"} {          // 20 tables
		require.Truef(t, m.HasTable(tbl), "missing table %s", tbl)
	}
	require.True(t, m.HasIndex(&models.User{}, "Username"))
	require.True(t, m.HasColumn(&models.VMSnapshot{}, "HostID"))          // normalisation carriers
	require.False(t, m.HasTable("vm_snapshots_records"))                  // wrappers stay tables-free (D-2)
	var ddl []string  // snapshot: ordered dump for byte-diff vs testdata/schema_snapshot.sql
	db.Raw("SELECT sql || ';' FROM sqlite_master WHERE type IN ('table','index') ORDER BY name").Scan(&ddl)
	require.Equal(t, strings.Join(ddl, "\n"), golden(t, "testdata/schema_snapshot.sql")) // regen mode via -update flag
}
```
- Assert serializer columns are `TEXT` (`PRAGMA table_info(catalog_manifests)`), composite PK on
  `role_claims`, unique indexes (`users.username`, `api_keys.key`,
  `user_configs(user_id,slug)`, `orchestrator_enrollment_tokens.token`).
- **Mapping finalisation pass**: diff every row of the plan mapping vs `PRAGMA table_info` output —
  serialised fields are TEXT cols; dropped fields absent (`DbRecord` ⇒ no `is_locked` column
  anywhere); update any drifted row + re-run waves' tests.
- Record D-5 evidence at top of `automigrate_test.go` comment block (closes AC5): grep found **no**
  delete-flag semantics ⇒ no `gorm.DeletedAt`; `Revoked*` kept as status columns.

## Out of scope
Data migration itself (#424 consumes this schema + mapping; does not run here), #421 queries,
`src/models` package.

## Validation
1. `cd src && go build ./... && go test ./... -count=1` green (whole repo, final boundary).
2. `go test ./data/models/... -run AutoMigrate -count=1` green; re-run twice (idempotent AutoMigrate).
3. `go test ./data/models/... -count=1` — 02–05 suites all green against final tags.
4. Mapping spot-audit: 5 random rows vs `sqlite_master` DDL by eye (PR checklist box).

## Rollback
Revert PR — test + doc only (unless gap-fixes shipped, which are additive tag/index changes).

## PBI acceptance mapping
AC3 (**primary**), AC1 (gap pass closes stragglers), AC4 (schema covers every mapped column —
finalisation pass is its proof), AC5 (evidence recorded — deliberately skipped), D-2 table list
confirmed = 20.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `419-06`.** Cut `slice/419-06` from `sql-migration` HEAD
(`.worktrees/slice-419-06`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
