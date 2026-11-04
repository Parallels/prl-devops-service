# PBI Plan #424 — One-time data migration (legacy → GORM)

> Reading order: [feature brief](../brief.md) → [feature plan](../plan.md) (locked **AD-3, AD-4, AD-5**, ER-1…ER-5, OQ-1) →
> [PBI brief](./brief.md) → this plan → your task spec.
> Repo anchors below re-verified against HEAD 2026-11-03 (`wc -l`: `src/data/main.go` 686, `src/startup/main.go` 345,
> `src/constants/main.go` 560, `src/serviceprovider/main.go` 530, `src/config/main.go` 1013, `src/security/main.go` 153).

## Scope

**In:** startup detection of the legacy `data.json` · read-only additive loader `data.LoadLegacyData` ·
completion-marker table `legacy_migration_marker` (in the **new** DB, AD-4) · transactional, idempotent
upsert migration of all 16 `Data` collections through #419's tagged models on #418's connection ·
per-collection progress logs · post-migration count validation (warn-only) · `--migrate-dry-run` flag +
`DB_MIGRATE_DRY_RUN` env · startup hook between the legacy-schema-migrations loop and serving, gated on
#420's `DB_AUTO_MIGRATE` · unit + integration tests and synthetic fixtures.

**Out:** touching `data.json` or any legacy write path (loader is additive-only, AD-5/OQ-1) · the
*frozen-legacy warn on fallback start* (that is #426) · `Data.Schema` → new DB (legacy JSON schema
versions and GORM AutoMigrate are different concepts — feature plan AD-4 explicitly says they must not
collide) · deleting/archiving `data.json` (fallback window) · `DB_TYPE`/`DB_PATH`/`DB_DSN` parsing
(#420) · GORM tags / join-table design (#419 — we **consume** its mapping table) · per-domain CRUD in
#421's query layer (we use raw batched upserts, see D-6) · MySQL stub `src/sql` (untouched).

**Depends-on (merged before task 01 starts):** #418 task 05 (`serviceprovider.GetSQLDatabase`,
failure-policy precedent), #419 (tagged models **+ its mapping-table document** — the field→column
truth), #420 (`DatabaseAutoMigrate()` accessor over `DB_AUTO_MIGRATE`/`database.auto_migrate`),
#421 (schema exists in production runs; models + connection are the compile-time substrate).
#419's mapping table is expected at
[`.ai-agents/features/417-gorm-database-layer/419-gorm-model-tags/plan.md`](../419-gorm-model-tags/plan.md)
(its brief: "a **mapping table document** … produced in `plan.md` — #424's spec derives from it
verbatim"). **That document does not exist yet at planning time.** Tasks 03–05 therefore consume it
through the `CollectionSpec` contract in D-7 and are marked `Depends-on #419 plan`; we invent no table,
column, or join names.

## Design decisions

### D-1 — New sub-package `src/database/legacy_migrator/`, one concern per file

Package `legacy_migrator` under `src/database/` per feature plan §7. Import graph stays acyclic:
`startup` → `legacy_migrator` → { `database` (parent, for the `*gorm.DB` handle only), `data` (loader +
`Data`), `data/models`, `serviceprovider`-free }. `data` and `database` never import `legacy_migrator`.

```
src/database/legacy_migrator/
├── errors.go        # sentinels (task 01)          ~25 lines
├── detector.go      # os.Stat-only detection (01)  ~70 lines
├── marker.go        # marker model + checks (01)   ~110 lines
├── loader.go        # thin call wrapper + Report types (02/03) ~60 lines
├── migrator.go      # orchestration loop + tx (03) ~180 lines
├── mapping.go       # CollectionSpec contract + plan (03/04) ~150 lines (split per domain if >400)
├── validation.go    # count checks, warn-only (05) ~110 lines
├── dryrun.go        # flag/env resolution + report (06) ~60 lines
└── testdata/        # synthetic fixtures (02/07) — see D-10
```

All ≤ ~200 lines → far under ER-2; no file this PBI edits is near the 1,000-line cap (guard tables per
task). `src/config/main.go` (1,013 — already over cap, #420 owns its ER-3 split) is **not** touched:
dry-run flag/env are read directly from `helper.GetFlagValue`/`os.LookupEnv` in `dryrun.go`.

### D-2 — Read-only loader: additive `data.LoadLegacyData`, never `JsonDatabase`

`NewJsonDatabase` (`src/data/main.go:75-125`) is unusable as migration source: it returns the
`memoryDatabase` **singleton**, immediately runs `Load` (which may *write*), starts `SaveAsync`
(`:102`, save-interval goroutine at `:226-247`), `RunBackup` (`:107`), and a ghost-job ticker
(`:110-122`). `loadFromFile` itself is also mutating: it calls `j.Backup(ctx)` (`:535`) before reading
and `RecoverOngoingJobs` (`:578`) after. AD-5 says *reuse the load path*, not *instantiate the engine* —
so we add one exported, **pure** function in package `data` (new file `src/data/legacy_load.go`, ~120
lines; existing `main.go` untouched — additive-only change approved by this PBI per the brief's
"legacy code is frozen except additive helpers approved in this PBI's tasks"):

```go
// LoadLegacyData reads and decodes a legacy data file WITHOUT any side effects:
// no file creation, no backup, no save, no singleton, no goroutines, no job recovery.
// Decoding mirrors loadFromFile (main.go:550-570): plain JSON first, then
// security.DecryptString(cfg.EncryptionPrivateKey(), raw) when unmarshalling fails.
// Residual recovery mirrors Load's order (*.save → *.save_bak → *.panic, main.go:135-157)
// but READ-ONLY: the newest matching glob file is decoded as fallback; nothing is saved,
// renamed, or removed (legacy store frozen — OQ-1 / D-4 of this plan).
// Returns ErrNoLegacyFile (404) when os.Stat says the file is absent, zero-value Data on error.
func LoadLegacyData(ctx basecontext.ApiContext, filename string) (Data, error)
```

Why additive rather than exporting a `JsonDatabase` constructor variant: a constructor can't avoid the
package-level `memoryDatabase` global being populated (`:76-78`), and the global feeds
`serviceprovider`/test harness resets (`src/data/testing_reset.go`) — instantiating it during migration
would let a half-loaded legacy DB leak into other components. `LoadLegacyData` duplicates only the
~20-line decode sequence of `loadFromFile` (documented cross-reference comment at the copy site), not
the decrypt key handling (still `security.DecryptString`, still `config.Get().EncryptionPrivateKey()`
— PEM content decoded from base64, `src/config/main.go:261-274` + `ENCRYPTION_SECURITY_KEY_ENV_VAR`).

### D-3 — Detection: `os.Stat` only; `IsDataFileEmpty` is banned

AC-1: present = file exists **and** size > 0. **Never** call `JsonDatabase.IsDataFileEmpty`
(`src/data/main.go:617-672`): when the file is missing it **creates it** (`:645-655`) — turning a fresh
install into a "legacy store exists" event. Absent or zero-byte → skip silently (one `ctx.LogDebugf`
line; AC says "skipped silently", so no Info/Warn). Source path resolution is not this PBI's job: the
startup hook passes `serviceprovider.Get().JsonDatabase.Filename()` (read-only getter,
`src/data/main.go:188-190`) — the same value the provider already resolved from `DATABASE_FOLDER` /
uid-home convention (`src/serviceprovider/main.go:69-104`, AD-3's folder).

### D-4 — Marker: table `legacy_migration_marker`, one row per completed source fingerprint

Feature plan AD-4 names the marker table loosely (`schema_migrations` or equivalent); **refinement
recorded here:** keep the concrete name `legacy_migration_marker` — `schema_migrations` would invite
confusion with the *legacy JSON* `Data.Schema` versioning and with golang-migrate conventions, and
AD-4's own rule is that the two systems never collide. The feature brief's OQ-1 ("frozen after
migration; warn on fallback start") stays with #426; we only supply the data it needs (fingerprint).

```go
// marker.go — model owned by legacy_migrator (NOT by #419's AutoMigrate list):
// the table is ensured by db.AutoMigrate(&LegacyMigrationMarker{}) inside RunMigration,
// so this PBI never edits #419's model registry. String dates match the models/ convention.
type LegacyMigrationMarker struct {
    ID               string `gorm:"primaryKey;type:text" json:"id"`          // helpers.GenerateId()
    SourcePath       string `gorm:"type:text;not null"   json:"source_path"`
    SourceSizeBytes  int64                               json:"source_size_bytes"`
    SourceModTimeUnix int64                              json:"source_mod_time_unix"`
    ContentSHA256    string `gorm:"type:text;uniqueIndex" json:"content_sha256"`
    CollectionsTotal int64                               json:"collections_total"`
    CompletedAt      string `gorm:"type:text;not null"   json:"completed_at"` // helpers.GetUtcCurrentDateTime()
}
func (LegacyMigrationMarker) TableName() string { return "legacy_migration_marker" }
```

Check order at startup (cheap → expensive):

1. `DetectLegacyStore` (D-3) — no file ⇒ skip, no marker queries at all.
2. **Fast probe:** `SELECT … WHERE source_path = ? AND source_size_bytes = ? AND source_mod_time_unix = ?`
   — hit ⇒ already migrated *this exact file*, skip (no full read of `data.json` on every boot).
3. Full fingerprint: read raw bytes (the loader will read them anyway when migrating), `sha256` →
   `SELECT … WHERE content_sha256 = ?` — hit ⇒ migrated file re-pointed elsewhere / moved by another
   install, skip with `ctx.LogInfof`.
4. Any marker rows exist but none match the current file ⇒ the frozen legacy store changed after
   migration ⇒ `ctx.LogWarnf("legacy store %s changed after migration (frozen legacy, OQ-1); not re-migrating")`
   and skip. Re-migrating deltas is explicitly **not** in scope — one-time migration only (PBI title).

Accepted-window risk (recorded in Risks R-3): step 2 trusts `size+mtime`; a same-size same-mtime edit
after migration is indistinguishable — impossible under OQ-1's freeze, and step 4's hash check covers
the realistic cases. `MarkMigrated` upserts with `clause.OnConflict{Columns: ContentSHA256, UpdateAll:true}`.

### D-5 — Atomicity: ONE `db.Transaction` around every collection + the marker

AC-3 literally says "runs inside **a** transaction; if any step fails, **all** changes are rolled back".
The PBI brief floated per-collection-group transactions to avoid long SQLite locks; rejected because
partial commits make "all rolled back" impossible without compensating deletes (= delete-then-insert,
which idempotency forbids). Argument (as required by the brief): the migration runs **after**
`startup.Start`'s legacy migration loop and **before** the API serves (`src/startup/main.go:136-140`;
AD-4's ordering), so the only budget it consumes is boot time — which is bounded by the very same
`data.json` the legacy service loads into RAM today (feature brief §5). With AD-2's driver
(`github.com/glebarez/sqlite` + WAL + `busy_timeout(10000)` from #418 D-4) and standard MVCC on
Postgres, one long write transaction is exactly what both backends are built for.

Consequences that are *features*: a crash mid-migration commits nothing (marker included — no partial
state to reason about); the next boot sees no marker and redoes the work; and even a hypothetically
partial state stays safe because every write is an ID-upsert (D-6), never delete-then-insert. Inside
the transaction, each collection is written with `tx.CreateInBatches(rows, chunkSize)` (default 500)
so per-statement memory and log volume stay bounded. GORM returns the rolled-back error from
`db.Transaction`; the caller applies the exit policy (D-9).

### D-6 — Write path: #419 tagged models + `clause.OnConflict` upserts on the #418 handle

`RunMigration(ctx basecontext.ApiContext, gdb *gorm.DB, opts Options)` receives the shared handle from
`serviceprovider.GetSQLDatabase(ctx)` (#418 task 05). Records are built as `*models.X` values carrying
#419's GORM tags, and written as:

```go
tx.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "id"}},   // PK column name per #419 mapping table
    UpdateAll: true,                            // re-migration overwrites with legacy truth
}).CreateInBatches(chunk, chunkSize)
```

Why not through #421's `GormDatabase` methods: they model single-record API CRUD (per-method
read-modify-write, domain error mapping, `LockRecord` semantics) — no batch/upsert semantics and each
method would manage its own transaction, breaking D-5's single-transaction guarantee. We *do* share
#421's substrate (same models, same handle, same schema), which is the dependency the feature plan §8
actually puts us behind. Join tables (`user_roles`, …) are written from the embedded slices exactly as
#419's mapping document prescribes; if #419 decides to serialise a slice into a JSON column instead of
a join table, that's transparent to us — we follow the doc (D-7).

Pointer collections (`Data.Configuration *models.Configuration`, `Data.ReverseProxy
*models.ReverseProxy`, `src/data/main.go:35,44`) map to single-row upserts; `nil` ⇒ skip + count 0.
String dates (`CreatedAt` etc. are `string` — #419 brief keeps them `string` v1) pass through verbatim:
byte-level fidelity, no parsing, no locale games.

### D-7 — `mapping.go` consumes #419's mapping document verbatim; nothing invented

The single source of truth for legacy-field → table/column/join/JSON-column/dropped is #419's plan.md
table. Our code encodes it through one contract:

```go
type CollectionSpec struct {
    Name        string                       // legacy field name: "Users", "ManifestsCatalog", …
    LegacyCount func(d *data.Data) int       // len() or 1/0 for pointer fields
    Migrate     func(ctx basecontext.ApiContext, tx *gorm.DB, d *data.Data, chunkSize int) (int64, error)
    Validate    func(tx *gorm.DB, d *data.Data) (expected int64, actual int64, err error) // AC-6 formula, straight from #419 doc
}
func collectionPlan() []CollectionSpec       // deterministic order: FK-safe (referenced before referrer),
                                             // then Data-field declaration order (main.go:34-50) as tiebreak
```

Scalar collections (`ApiKeys`, `EnrollmentTokens`, `UserConfigs`, `Jobs`, `PackerTemplates`, …: no
embedded entity slices) are a mechanical struct-conversion + upsert — task 03. Relational ones
(`Users`/`Roles`/`Claims` with `User.Roles`/`User.Claims`/`Role.Claims` embedded —
`src/data/models/user.go:11-12`, `role.go:8-9`; `CatalogManifest` nested contents —
`catalog_manifest.go:28-29`; `OrchestratorHost` nested VMs — `orchestrator_host.go:36,43`; snapshots
with `map[string][]VMSnapshot` — `orchestrator_vm_snapshots.go:6`) are task 04, each `Migrate` body
transcribed from the #419 row. `Data.Schema` is **not** migrated (AD-4).

### D-8 — Validation AC-6: legacy counts in memory vs `COUNT(*)` after commit, warn-only

Runs after the transaction commits (post-commit `COUNT` sees exactly the committed truth). Legacy side
= `CollectionSpec.LegacyCount(&legacy)`; new side = `CollectionSpec.Validate` (`tx` is the plain handle
now; join-table formulas, e.g. "`user_roles` rows = Σ len(u.Roles)", come from the #419 doc column of
the same row). Mismatch ⇒ `ctx.LogWarnf("validation: %s legacy=%d sql=%d — counts differ …")`,
per-connection, and **never** fails the migration (AC wording: "logs a warning"). Dry-run skips
SQL-side counts (nothing committed) and reports projected counts only.

### D-9 — Dry-run AC-7 + startup hook + failure policy

Flag/env idiom, verified: constants live in `src/constants/main.go` (flag block at `:160-175`, env block
`:82-98`) and are read `helper.GetFlagValue(constants.CONFIG_FILE_FLAG, "")`-style
(`src/config/main.go:122`; common-go parses `--flag=value` and `--flag value`). Two new constants
(`constants/main.go` 560 → ~563 lines): `MIGRATE_DRY_RUN_FLAG = "migrate-dry-run"` and
`MIGRATE_DRY_RUN_ENV_VAR = "DB_MIGRATE_DRY_RUN"`. Resolution order matches `Config.GetKey`'s
last-writer-wins (config→flag→env, `src/config/main.go:933-947`): env overrides flag. Because
`GetFlagValue` needs a value, the documented spellings are `--migrate-dry-run=true` or
`DB_MIGRATE_DRY_RUN=true`; a value-less `--migrate-dry-run` is ignored (same behaviour as every other
repo flag — noted for #427 docs).

Dry-run executes detection → marker checks → `LoadLegacyData` → **all** `Transform`/mapping code and
count math → prints the same per-collection lines with a `[dry-run]` prefix → writes **nothing**:
no transaction is opened, no marker row, `data.json` untouched. Exit code 0. During a normal API boot
with the flag set, the process logs the report and continues booting **without** migrating (an
explicit operator request must never be silently overridden; flagged as OQ-C for #427 wording).

Startup hook — inserted immediately after the legacy-schema-migrations loop
(`src/startup/main.go:136-140`, exactly AD-4's "legacy migrations → AutoMigrate → #424 → serve";
AutoMigrate has already run at provider-init time per #418/#420/#421):

```go
// after the `for _, migration := range schemaMigrations` loop:
if cfg.DatabaseAutoMigrate() { // #420: DB_AUTO_MIGRATE / database.auto_migrate, default true
    if err := runLegacyDataMigration(ctx); err != nil {
        ctx.LogErrorf("Legacy data migration failed (rolled back): %v", err)
        if cfg.GetRunningCommand() == constants.API_COMMAND {
            os.Exit(1) // AC-3 "clear error + exit"; precedent: #418 task 05 mustInitGormConnection,
                       // src/main.go:59-63 (fmt.Fprintf(os.Stderr, …); os.Exit(1))
        }              // non-API commands (CLI, tests): error already logged, boot continues —
}                      // same policy split as #418 D-7 so unit suites never die on exit(1)
}
```

`runLegacyDataMigration` is a ≤40-line local helper in `startup/main.go` (345 → ~385 lines, ok):
fetch `GetSQLDatabase` (error = DB down; #418's own policy already handled that path), build
`Options{SourcePath: provider.JsonDatabase.Filename(), DryRun: legacy_migrator.IsDryRunRequested()}`,
call `RunMigration`, `LogInfof` the report. Rollback is D-5's; nothing here compensates.

### D-10 — Fixtures: committed tiny JSON, everything generated at test time

`src/database/legacy_migrator/testdata/`: hand-crafted synthetic files only (never real user data):
`golden.json` (few users w/ roles+claims, 1 configuration, 1 reverse-proxy host, jobs, tokens — the
AC-proposal "checked-in golden data.json"), `partial.json` (sparse), `corrupt.json` (not JSON, not
valid ciphertext), `empty.json` (0 bytes; also produced via `os.WriteFile(…, nil, 0o600)` because the
detector test wants it under `t.TempDir()`). **Encrypted fixture is generated at test time, not
committed**: `EncryptString` (RSA-PKCS1v15-wrapped random AES key + CFB stream,
`src/security/main.go:69-104`) is randomised — committed golden bytes would be *decrypt-only* and
bigger than the test's value; instead `security.GenPrivateRsaKey(t.TempDir()+"/key.pem", 2048)` →
`EncryptString(pem, goldenJSON)` → `t.Setenv(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR,
base64.StdEncoding.EncodeToString(pem))` (`config.EncryptionPrivateKey()` reads env last,
`config/main.go:933-947`, so the seam is proven). Giant-store case is likewise generated: 5,000 synthetic
`Jobs` → asserts 10 × 500-batch upserts + progress line count — no multi-MB blobs in git.

## Target public API (contract sketch)

```go
// detector.go
type Detection struct{ Present bool; SizeBytes int64; Reason string }
func DetectLegacyStore(ctx basecontext.ApiContext, path string) (Detection, error)

// marker.go — model above, plus:
func EnsureMarkerTable(db *gorm.DB) error
func AlreadyMigrated(ctx basecontext.ApiContext, db *gorm.DB, det Detection, path string) (bool, error)
func MarkMigrated(db *gorm.DB, m LegacyMigrationMarker) error

// migrator.go
type Options struct {
    SourcePath string
    DryRun     bool
    ChunkSize  int                                  // <=0 ⇒ default 500
    FailAfter  func(name string, done int) error    // test seam for task 07; nil in prod
}
type CollectionReport struct{ Name string; LegacyCount int; WrittenCount int64 }
type Report struct { Skipped bool; SkipReason string; DryRun bool
                    Collections []CollectionReport; MarkerID string }
func RunMigration(ctx basecontext.ApiContext, gdb *gorm.DB, opts Options) (Report, error)

// validation.go
func ValidateCounts(ctx basecontext.ApiContext, db *gorm.DB, legacy *data.Data) []string

// dryrun.go
func IsDryRunRequested() bool
```

## Task index (each task = one PR, merge in order)

| Task | Goal | Depends on | Touches | Size |
|---|---|---|---|---|
| [01](./tasks/01-detector-and-marker.md) | `DetectLegacyStore` + marker model/table/checks + tests | #418-#421 merged | `errors.go`, `detector.go`, `marker.go`, tests | S |
| [02](./tasks/02-read-only-loader-loadlegacydata.md) | `data.LoadLegacyData` + plaintext/encrypted/corrupt/empty tests | 01 (fixtures layout) | `src/data/legacy_load.go`, tests, `testdata/` | M |
| [03](./tasks/03-core-migrator-scalar-collections.md) | `RunMigration` skeleton: single tx, upsert engine, chunking, progress logs; scalar collections | 02, **#419 doc** | `migrator.go`, `loader.go`, `mapping.go`, tests | M |
| [04](./tasks/04-relational-collections-mapping.md) | users/roles/claims joins, catalog, orchestrator, snapshots per #419 doc | 03 | `mapping*.go`, tests | M |
| [05](./tasks/05-validation-counters.md) | `ValidateCounts` + join-table formulas, warn-only | 04 | `validation.go`, tests | S |
| [06](./tasks/06-startup-wiring-dryrun-and-gate.md) | startup hook + `DB_AUTO_MIGRATE` gate + dry-run flag/env + exit policy | 03..05 | `dryrun.go`, `startup/main.go` (+~40), `constants/main.go` (+2) | M |
| [07](./tasks/07-integration-idempotency-rollback.md) | `//go:build integration` suite: golden e2e, re-run, injected mid-migration failure → full rollback + exit semantics, chunking | 06 | `*_integration_test.go`, `testdata/golden.json` | M |

Sequencing notes: tasks 01–05 ship dead-but-tested code (nothing wired until 06 → main never changes
behaviour); 03 and 04 both touch `mapping.go` → strictly sequential; only 06 edits files outside
`src/database/legacy_migrator/` besides 02 (`data/legacy_load.go`).

## AC-coverage (every checkbox of PBI #424)

| AC | Task(s) | Proof |
|---|---|---|
| Detection (silent skip) | 01 (+06 wiring) | `detector_test.go`: absent / 0-byte / present ×3; grep audit: no `IsDataFileEmpty`/`NewJsonDatabase` in package |
| Idempotent | 01, 03, 07 | marker short-circuit test; upsert re-run test (apply golden twice → identical counts, no dup rows); integration re-run |
| Atomic + exit | 03, 06, 07 | `FailAfter` seam test asserts zero rows incl. marker after error; integration asserts exit-1-under-API_COMMAND policy fn |
| Progress logging | 03 | captured-context test: one `migrated %d/%d …` per chunk, Info level |
| Completion marker | 01, 03 | marker row content (path/size/mtime/sha256/completed_at) + post-success boot = skip |
| Data integrity validation | 05 | per-collection count matrix incl. join formulas; warn-only proven (test passes with mismatch) |
| Dry-run | 03 (mode), 06 (flag) | dry-run test: DB byte-identical (`COUNT(*)==0`, no marker), `data.json` hash unchanged; flag/env matrix |
| DoD "first-run on golden fixture in CI" | 07 | `//go:build integration` golden e2e + rollback suite |

## Verification commands (all tasks)

```bash
cd src
go build ./...
go vet ./database/... ./data/... ./startup/...
go test ./database/legacy_migrator/... -count=1
go test ./data/... -count=1                       # loader addition must keep legacy suite green
go test ./... -count=1                            # full regression (ER-5)
go test -tags integration ./database/legacy_migrator/... -count=1   # task 07 onward
```

## Risks specific to #424

| # | Risk | Mitigation |
|---|---|---|
| R-1 | #419 mapping doc lands with decisions that differ from our sketches (JSON-column vs join table, PK column names) | Nothing in 01/02/06 depends on #419 specifics; 03/04 transcribe the doc at implementation time; `CollectionSpec` isolates the blast radius |
| R-2 | Huge `jobs` table ⇒ long boot (feature risk R-7) | chunked batches + per-chunk progress lines; OQ-2 (jobs retention) is #421's — if unbounded, document boot-time guidance in #427 |
| R-3 | Fast-probe skip trusts size+mtime (D-4 step 2) | accepted: frozen legacy (OQ-1) makes the collision scenario self-contradictory; hash check (step 3) covers path/move changes; warn (step 4) surfaces edits |
| R-4 | Long single-tx migration vs SQLite `busy_timeout(10000)` if another writer holds the new DB at boot | nobody serves traffic yet (hook is pre-serve); WAL allows the single writer; documented: don't run two instances against one new DB mid-upgrade |
| R-5 | Encrypted store + missing `ENCRYPTION_PRIVATE_KEY` at boot ⇒ load error ⇒ exit-1 bricks first post-upgrade boot | error message names the env var explicitly; integration test covers "encrypted + no key" = clear failure, not silent data loss; operator story documented for #427 |
| R-6 | Dry-run requested on normal boot silently skips real migration later surprises the operator | report line states "no data was written; migration did not run" (OQ-C for final wording) |
| R-7 | Marker table absent because `DB_AUTO_MIGRATE=false` | gate skips us too (D-9) with a single LogDebugf; documented interaction, matches AD-4 |

## Open questions (for the human review round)

| ID | Question | Leaning |
|---|---|---|
| OQ-A | #419 `plan.md` not yet written — mapping formulas will be transcribed at task 03/04 implementation | consume, don't invent (already enforced by D-7) |
| OQ-B | Marker fast-probe (size+mtime) acceptable given hash fallback? | yes — R-3 |
| OQ-C | Exact UX of `--migrate-dry-run` on a regular API boot (report and continue vs refuse to serve) | report + continue without migrating; #427 wording |
| OQ-D | Should `LoadLegacyData` *repair* via residual `.save` files like legacy `Load` does? | no — read-only fallback only; legacy stays frozen |
| OQ-E | `DB_AUTO_MIGRATE=false` on an unmigrated install ⇒ skip migration with debug log | yes; warn wording alternative for #427 |

## Handover to #425 / #426 / #427

- #425: fixture factory helpers from task 02 (encrypted-file builder, big-store generator) are exported
  from `legacy_migrator` under `// Exported for #425 migration tests.` comments; reuse, don't duplicate.
- #426: frozen-legacy warn should read `legacy_migration_marker` (D-4 gives path + sha256); no new state needed.
- #427: document `--migrate-dry-run=true` / `DB_MIGRATE_DRY_RUN=true`, boot-time expectations, and the
  "legacy file is kept, never modified" guarantee.
