# Task 01 — Legacy-store detector + completion marker

**PBI** [#424](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** #418 PR5 merged
(`serviceprovider.GetSQLDatabase`, `database.NewDatabaseConnection` with `:memory:`), #419/#420/#421 merged
(merge order per feature plan §8) · **PR size:** S
**Plan refs:** [D-1](../plan.md#d-1--new-sub-package-srcdatabaselegacy_migrator-one-concern-per-file), [D-3](../plan.md#d-3--detection-osstat-only-isdatafileempty-is-banned), [D-4](../plan.md#d-4--marker-table-legacy_migration_marker-one-row-per-completed-source-fingerprint)

## Context

Everything after this task assumes two primitives: a detection answer that never mutates the filesystem
(the legacy `IsDataFileEmpty` helper *creates* `data.json` when missing — `src/data/main.go:645-655` —
so it is banned here), and a marker the new database owns (AD-4: marker lives in the **new** DB, one
row per completed legacy source fingerprint). Both are dead code until task 06 wires them — this PR can
not change runtime behaviour.

## Objective

`legacy_migrator.DetectLegacyStore` answers "present? how big?" from `os.Stat` alone, and the marker
table `legacy_migration_marker` can be created, probed (fast size+mtime path, sha256 fallback path),
and upserted — all proven by unit tests on a `:memory:` SQLite handle from #418's real factory (no
GORM mocks, feature plan §12).

## Files

| File | Op |
|---|---|
| `src/database/legacy_migrator/errors.go` | create |
| `src/database/legacy_migrator/detector.go` | create |
| `src/database/legacy_migrator/marker.go` | create |
| `src/database/legacy_migrator/detector_test.go` | create |
| `src/database/legacy_migrator/marker_test.go` | create |

## File size guard

`wc -l` at plan time: nothing existing is edited. New files ≈ 25 / 70 / 110 / 90 / 130 lines.
Verdict: **ok** (ER-2 cap 1,000; target ≤ ~200/file per plan D-1).

## Implementation spec

### `errors.go`

```go
package legacy_migrator

import "github.com/Parallels/prl-devops-service/errors"

var (
        ErrNoLegacyFile       = errors.NewWithCode("legacy data file not found", 404)
        ErrLegacyFileEmpty    = errors.NewWithCode("legacy data file is empty", 404)
        ErrMarkerCheck        = errors.NewWithCode("legacy migration marker check failed", 500)
        ErrMarkerTableMissing = errors.NewWithCode("legacy_migration_marker table missing; run with DB_AUTO_MIGRATE=true", 500)
)
```

Repo idiom: `errors.NewWithCode` (see `src/data/main.go:21-24`), HTTP-ish codes per convention.

### `detector.go`

```go
// Detection is the AC-1 answer. Reason is human-readable ("absent", "empty", "present").
type Detection struct {
        Present   bool
        SizeBytes int64
        ModTime   time.Time
        Reason    string
}

// DetectLegacyStore stats the legacy file. Pure: os.Stat only.
// BANNED here (review gate): JsonDatabase.IsDataFileEmpty — it CREATES the file
// (src/data/main.go:645-655) — and anything that constructs JsonDatabase.
func DetectLegacyStore(ctx basecontext.ApiContext, path string) (Detection, error) {
        info, err := os.Stat(path)
        if err != nil {
                if os.IsNotExist(err) {
                        return Detection{Reason: "absent"}, nil // AC-1: skip silently, no error
                }
                return Detection{}, errors.NewFromError(err)
        }
        if info.Size() == 0 {
                return Detection{Reason: "empty"}, nil
        }
        return Detection{Present: true, SizeBytes: info.Size(), ModTime: info.ModTime(), Reason: "present"}, nil
}
```

Callers log at **debug** level for absent/empty ("skipped silently" per AC-1). `path` comes from the
startup hook (task 06) via `JsonDatabase.Filename()` (`src/data/main.go:188-190`) — never re-derived
here (single resolution point; AD-3/#420).

### `marker.go`

Model + 3 funcs exactly per [plan D-4](../plan.md#d-4--marker-table-legacy_migration_marker-one-row-per-completed-source-fingerprint):

```go
type LegacyMigrationMarker struct { /* plan D-4 fields, TableName() = "legacy_migration_marker" */ }

func EnsureMarkerTable(db *gorm.DB) error { return db.AutoMigrate(&LegacyMigrationMarker{}) }

// AlreadyMigrated implements the plan-D-4 4-step check (fast probe → hash → foreign-file → frozen-warn).
// rawContent is passed in when the caller already holds the bytes (the migration path does);
// nil ⇒ the hash step reads the file itself.
func AlreadyMigrated(ctx basecontext.ApiContext, db *gorm.DB, det Detection, path string, rawContent []byte) (bool, error) {
        if !det.Present { return true, nil }                       // nothing to do; treat as "don't run"
        if err := EnsureMarkerTable(db); err != nil { return false, err }
        var hit int64
        if err := db.Model(&LegacyMigrationMarker{}).
                Where("source_path = ? AND source_size_bytes = ? AND source_mod_time_unix = ?",
                        path, det.SizeBytes, det.ModTime.Unix()).Count(&hit).Error; err != nil {
                return false, errors.NewFromError(err)
        }
        if hit > 0 { return true, nil }                            // fast probe: exact same file
        sum := sha256.Sum256(rawContent)                           // raw bytes — works for encrypted files too
        sha := hex.EncodeToString(sum[:])
        if err := db.Model(&LegacyMigrationMarker{}).Where("content_sha256 = ?", sha).Count(&hit).Error; err != nil {
                return false, errors.NewFromError(err)
        }
        if hit > 0 { ctx.LogInfof("legacy store %s already migrated (fingerprint %s…)", path, sha[:12]); return true, nil }
        var any int64
        db.Model(&LegacyMigrationMarker{}).Count(&any)
        if any > 0 {
                ctx.LogWarnf("legacy store %s changed after migration (frozen legacy, OQ-1); not re-migrating", path)
                return true, nil
        }
        return false, nil
}

func MarkMigrated(db *gorm.DB, m LegacyMigrationMarker) error {
        m.ID = helpers.GenerateId()
        return db.Clauses(clause.OnConflict{
                Columns:   []clause.Column{{Name: "content_sha256"}},
                UpdateAll: true,
        }).Create(&m).Error
}
```

`helpers.GenerateId()`, `helpers.GetUtcCurrentDateTime()` (string-date convention, cf.
`src/data/main.go:412`); `completed_at` set by the caller (task 03) at commit time.

### Tests (`detector_test.go`, `marker_test.go`)

testify (`require`/`assert`), `t.TempDir()`, `database.NewDatabaseConnection(ctx,
database.DatabaseConfig{Type: sqlite, Path: ":memory:"})` per #418 D-3; ctx =
`basecontext.NewRootBaseContext()`.

| Test | Assert |
|---|---|
| detector: path absent | `Present==false, Reason=="absent"`, err nil |
| detector: zero-byte file | `Present==false, Reason=="empty"` |
| detector: file w/ bytes | `Present==true, SizeBytes>0` |
| detector side-effect audit | after each case, `os.Stat` dir listing identical (no file created!) |
| marker: EnsureMarkerTable idempotent | call twice, no error; `Migrator().HasTable` true |
| marker: fast probe hit | row seeded w/ matching path/size/mtime ⇒ `AlreadyMigrated==true` **and** no content read (pass nil raw) |
| marker: hash hit after mtime touch | seed row w/ sha256 of bytes; det built from touched file ⇒ true |
| marker: changed file after migration | one foreign marker row ⇒ `true` + warn logged (capture ctx) |
| marker: clean slate ⇒ false | no rows ⇒ false |
| `MarkMigrated` upsert | same sha twice ⇒ exactly 1 row, updated `completed_at` |

## Out of scope

Migration execution, loader, dry-run, startup wiring (tasks 02–06). Any change under `src/data/`,
`src/startup/`, `src/serviceprovider/` — none belongs to this PR.

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go vet ./database/...`
- [ ] `cd src && go test ./database/legacy_migrator/... -count=1` green, < 5 s (no disk fixtures beyond `t.TempDir`)
- [ ] Grep audit (paste in PR): `grep -rn "IsDataFileEmpty\|NewJsonDatabase\|loadFromFile" src/database/legacy_migrator/` ⇒ **no matches**
- [ ] `cd src && go test ./... -count=1` regression green (new package is inert)

## Rollback

Revert PR: deletes 5 brand-new files; nothing else referenced them ⇒ zero blast radius.

## PBI acceptance mapping

Moves **AC-1 Detection** (logic complete; wiring lands in 06) and **AC-5 Completion marker** (storage
half; the write half lands in 03). Provides the idempotency gate AC-2 leans on.
