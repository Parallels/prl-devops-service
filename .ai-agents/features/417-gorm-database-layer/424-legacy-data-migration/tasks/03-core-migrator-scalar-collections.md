# Task 03 — Core migrator: single transaction, upsert engine, scalar collections

**PBI** [#424](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 02, **#419 merged**
(tagged models + its mapping-table document at
`../419-gorm-model-tags/plan.md` — the PK column names and per-field rows below are transcribed from it
at implementation time; nothing is invented here) · **PR size:** M
**Plan refs:** [D-5](../plan.md#d-5--atomicity-one-dbtransaction-around-every-collection--the-marker), [D-6](../plan.md#d-6--write-path-419-tagged-models--clauseonconflict-upserts-on-the-418-handle), [D-7](../plan.md#d-7--mappinggo_consumes_419s_mapping_document_verbatim_nothing_invented) *(anchors: D-5/D-6/D-7 in plan.md)*

## Context

This is the engine: one `db.Transaction` (AC-3), deterministic collection order, chunked
upsert-by-ID writes (`clause.OnConflict` — never delete-then-insert), per-chunk `info` progress
(AC-4), dry-run short-circuit (AC-7 core, flag wiring is task 06), and the marker write inside the same
tx (AC-5). Task 03 implements only the **scalar** collections (those whose legacy struct maps 1:1 onto a
#419-tagged model with no embedded entity slices); the relational ones are task 04 on top of the same
skeleton.

## Objective

`RunMigration(ctx, gdb, opts)` moves every scalar collection (`ApiKeys`, `PackerTemplates`,
`OrchestratorHosts` scalars-only pass *(nested slices handled in 04 — here the host row itself)*,
`ReverseProxyHosts`, `CatalogManagers`, `Jobs`, `EnrollmentTokens`, `UserConfigs`, pointer
`Configuration`) idempotently and atomically; a second run is a no-op; an injected mid-run failure
leaves the DB byte-identical to pre-run (incl. no marker).

## Files

| File | Op |
|---|---|
| `src/database/legacy_migrator/loader.go` | create — `Options`, `Report`, `runLoad` wrapper |
| `src/database/legacy_migrator/migrator.go` | create |
| `src/database/legacy_migrator/mapping.go` | create — `CollectionSpec`, `collectionPlan()`, scalar specs |
| `src/database/legacy_migrator/migrator_test.go` | create |
| `src/database/legacy_migrator/mapping_test.go` | create |

## File size guard

All paths new; planned ≈ 60 / 180 / 150 / 240 / 60 lines (ER-2 cap 1,000, D-1 target ≤ 200/impl file —
`mapping.go` carries the split plan: if #419's rows push it past ~400 lines, split
`mapping_scalar.go` + `mapping_relational.go`, same package, in task 04). Verdict: **ok**.

## Implementation spec

### `loader.go`

```go
type Options struct {
        SourcePath string
        DryRun     bool
        ChunkSize  int                                // <=0 ⇒ defaultChunkSize = 500
        FailAfter  func(name string, done int) error  // nil in production; task-07 seam: return err
                                                       // right after collection `name` wrote `done` rows
}
type CollectionReport struct { Name string; LegacyCount int; WrittenCount int64 }
type Report struct { Skipped bool; SkipReason string; DryRun bool; Collections []CollectionReport; MarkerID string }

const defaultChunkSize = 500 // R-2: bounded statement size & log volume for the jobs table
```

### `migrator.go` — the loop (full sketch)

```go
func RunMigration(ctx basecontext.ApiContext, gdb *gorm.DB, opts Options) (Report, error) {
        report := Report{DryRun: opts.DryRun}
        if opts.SourcePath == "" { return report, errors.NewWithCode("legacy migration: SourcePath is required", 500) }
        chunk := opts.ChunkSize; if chunk <= 0 { chunk = defaultChunkSize }

        det, err := DetectLegacyStore(ctx, opts.SourcePath)
        if err != nil { return report, err }
        if !det.Present {
                report.Skipped = true; report.SkipReason = det.Reason
                ctx.LogDebugf("legacy migration skipped: %s", det.Reason) // AC-1 "silently"
                return report, nil
        }
        raw, err := os.ReadFile(opts.SourcePath) // bytes needed for the fingerprint (D-4); encrypted-safe
        if err != nil { return report, errors.NewFromError(err) }
        done, err := AlreadyMigrated(ctx, gdb, det, opts.SourcePath, raw)
        if err != nil { return report, err }
        if done { report.Skipped = true; report.SkipReason = "already-migrated"; return report, nil }

        legacy, err := data.LoadLegacyData(ctx, opts.SourcePath) // AD-5 — the only read path (task 02)
        if err != nil { return report, err }

        plan := collectionPlan()                                  // deterministic, FK-safe (D-7)
        if opts.DryRun { return runDryRun(ctx, plan, &legacy, report) } // computes mappings, WRITES NOTHING (AC-7)

        marker := LegacyMigrationMarker{
                SourcePath: opts.SourcePath, SourceSizeBytes: det.SizeBytes,
                SourceModTimeUnix: det.ModTime.Unix(), ContentSHA256: fmt.Sprintf("%x", sha256.Sum256(raw)),
                CompletedAt: helpers.GetUtcCurrentDateTime(),
        }
        total := 0
        for _, s := range plan { total += s.LegacyCount(&legacy) }

        err = gdb.Transaction(func(tx *gorm.DB) error {           // AC-3: ONE tx, all-or-nothing (D-5)
                var written int64
                for _, s := range plan {
                        before := time.Now()
                        n, mErr := s.Migrate(ctx, tx, &legacy, chunk)
                        if mErr != nil {
                                return errors.NewWithCodef(500, "legacy migration failed in %s: %v", s.Name, mErr)
                        }
                        written += n
                        report.Collections = append(report.Collections, CollectionReport{s.Name, s.LegacyCount(&legacy), n})
                        ctx.LogInfof("migrated %d/%d %s records (total so far %d/%d, %s)", // AC-4
                                s.LegacyCount(&legacy), total, s.Name, written, total, time.Since(before).Round(time.Millisecond))
                        if opts.FailAfter != nil {
                                if fErr := opts.FailAfter(s.Name, int(written)); fErr != nil { return fErr } // rollback provokes (task 07)
                        }
                }
                marker.CollectionsTotal = written
                if uErr := MarkMigrated(gdbInTx(tx), marker); uErr != nil { return uErr } // marker INSIDE the tx: crash ⇒ nothing
                report.MarkerID = marker.ID
                return nil
        })
        if err != nil {
                ctx.LogErrorf("legacy migration rolled back: %v", err)
                return Report{DryRun: opts.DryRun}, err
        }
        return report, nil
}
```

`gdbInTx(tx)` note for reviewers: `MarkMigrated` takes `*gorm.DB`; pass `tx` (the callback param) so the
marker shares the transaction — signature `MarkMigrated(tx *gorm.DB, …)` from task 01 already accepts
any handle; the *concrete* tx is what must be threaded.

### `mapping.go` — spec contract + scalar rows

```go
type CollectionSpec struct {
        Name        string
        LegacyCount func(d *data.Data) int
        Migrate     func(ctx basecontext.ApiContext, tx *gorm.DB, d *data.Data, chunkSize int) (int64, error)
        Validate    func(db *gorm.DB, d *data.Data) (expected, actual int64, err error) // used by task 05; scalar rows: simple Count
}

// collectionPlan — order = FK-safe (referenced before referrer), Data-declaration order
// (src/data/main.go:34-50) as tiebreak. Final list MUST match #419 mapping doc rows;
// until that document is read at implementation time, ONLY the mechanical scalar entries below are
// added, and task 04 appends the relational ones. Table/PK/column names come from #419 — verbatim.
func collectionPlan() []CollectionSpec { return append(scalarPlan(), relationalPlan()...) } // 04 fills its half
```

Scalar `Migrate` is the upsert engine, once, generically:

```go
func upsertAll[T any](ctx basecontext.ApiContext, tx *gorm.DB, rows []T, pk string, chunk int) (int64, error) {
        if len(rows) == 0 { return 0, nil }
        res := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: pk}}, UpdateAll: true}).
                CreateInBatches(&rows, chunk)      // AC-2: re-run = same PK ⇒ update, never duplicate
        return res.RowsAffected, res.Error          // caveat (in code comment): RowsAffected ≠ len(rows) on
}                                                   // MySQL-family update-in-place; fine for SQLite/PG semantics used here
```

Example rows (the `pk` strings and target-model names are **#419-doc placeholders** — confirm at
implementation): `{Name:"ApiKeys", LegacyCount: lenOf(&Data.ApiKeys), Migrate: modelCopy[models.ApiKey]}`
where `modelCopy` converts the legacy struct to the #419 tagged model **by field assignment, not JSON
round-trip** (compile-time typo-catching; string dates copied verbatim, D-6). Pointer
`Configuration`: nil ⇒ `return 0,nil` (count 0); else one-row `upsertAll` slice of 1. Same shape for
`ReverseProxy`. Collections whose nested slices still exist (`Jobs.Steps`, host VM lists) get their
row written here and slices completed in task 04 — task 03's tests use fixtures without nested slices so
no half-mapping ships.

### Tests (`migrator_test.go`)

`:memory:` handle; `LoadLegacyData` fed from `testdata/` fixtures; log capture via a ctx wrapper
(recording `LogInfof` calls).

| Test | Assert |
|---|---|
| golden → scalar tables | rows exist; `Report.Collections` matches fixture counts |
| idempotent re-run (AC-2) | run twice ⇒ identical `COUNT(*)` per table, marker row count == 1, no error |
| crash rollback (AC-3) | `FailAfter` on 2nd collection ⇒ every table `COUNT(*)==0` **and** `legacy_migration_marker` empty |
| progress logs (AC-4) | one `migrated %d/%d` Info line per collection; contains counts |
| marker content (AC-5) | path/size/mtime/sha256 match the fixture file; `completed_at` non-empty |
| dry-run core (AC-7) | `DryRun:true` ⇒ all tables empty, no marker, `Report` populated with projected counts, fixture file sha256 unchanged |
| chunking | 5,000 synthetic jobs (generator helper, D-10) + `ChunkSize:500` ⇒ completes; ≥10 batch statements (assert via GORM `DryRun` counter or log line count) — helper `BuildLargeJobsFixture` exported for #425 |
| `Options.SourcePath==""` | error, no panic |

## Out of scope

Relational/join collections (04), validation counting (05), flags/startup (06), any edit to `src/data`
or `src/startup`.

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go vet ./database/...`
- [ ] `cd src && go test ./database/legacy_migrator/... -count=1` green (incl. tasks 01–02 suites)
- [ ] `cd src && go test ./... -count=1` regression green
- [ ] Cross-check PR description: every table/PK string used maps to a row in #419's mapping doc table (link + paste the 8-10 rows) — reviewer gate against invented columns

## Rollback

Revert PR: package files only; still zero call sites ⇒ no runtime surface.

## PBI acceptance mapping

Moves **AC-3 Atomic** (single tx + rollback test), **AC-2 Idempotent** (upsert engine), **AC-4 Progress
logging**, **AC-5 marker write**, half of **AC-7 Dry-run** (mode semantics; flag in 06).
