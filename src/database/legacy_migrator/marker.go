package legacy_migrator

// Completion marker (AC-5 storage half, plan D-4): the new database owns
// this table — one row per completed legacy-source fingerprint. The table
// is ensured HERE (EnsureMarkerTable), deliberately NOT added to #419's
// AutoMigrate registry: this PBI never edits that list.
//
// AlreadyMigrated is the 4-step cheap→expensive gate: fast size+mtime probe
// (no data.json read at all on the happy path), sha256 fingerprint fallback
// (works for encrypted sources — raw bytes hash), and the frozen-legacy
// warn path (OQ-1 data for #426).

import (
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LegacyMigrationMarker struct {
	ID                string `gorm:"primaryKey;type:text" json:"id"` // helpers.GenerateId()
	SourcePath        string `gorm:"type:text;not null" json:"source_path"`
	SourceSizeBytes   int64  `json:"source_size_bytes"`
	SourceModTimeUnix int64  `json:"source_mod_time_unix"`
	ContentSHA256     string `gorm:"type:text;uniqueIndex" json:"content_sha256"`
	CollectionsTotal  int64  `json:"collections_total"`
	CompletedAt       string `gorm:"type:text;not null" json:"completed_at"` // helpers.GetUtcCurrentDateTime(); caller (task 03) stamps at commit time
}

func (LegacyMigrationMarker) TableName() string { return "legacy_migration_marker" }

// EnsureMarkerTable is idempotent AutoMigrate of the marker table only.
func EnsureMarkerTable(db *gorm.DB) error {
	return db.AutoMigrate(&LegacyMigrationMarker{})
}

// AlreadyMigrated implements the plan-D-4 4-step check (fast probe → hash →
// foreign-file → frozen-warn). rawContent is passed in when the caller
// already holds the bytes (the migration path does); nil ⇒ the hash step
// reads the file itself.
func AlreadyMigrated(ctx basecontext.ApiContext, db *gorm.DB, det Detection, path string, rawContent []byte) (bool, error) {
	if !det.Present {
		return true, nil // nothing to do; treat as "don't run" (AC-1 skip)
	}
	if err := EnsureMarkerTable(db); err != nil {
		return false, errors.NewFromError(err)
	}
	// Step 2: fast probe — exact same file, no content read.
	var hit int64
	if err := db.Model(&LegacyMigrationMarker{}).
		Where("source_path = ? AND source_size_bytes = ? AND source_mod_time_unix = ?",
			path, det.SizeBytes, det.ModTime.Unix()).Count(&hit).Error; err != nil {
		return false, errors.NewFromError(err)
	}
	if hit > 0 {
		// Visible skip (matrix round 3: a silent restart looked like a
		// missing hook). Every already-migrated path logs; absent/empty
		// stay silent per AC-1.
		ctx.LogInfof("legacy store %s already migrated (unchanged since migration)", path)
		return true, nil
	}
	// Step 3: full fingerprint over RAW bytes (encrypted files hash fine).
	if rawContent == nil {
		raw, err := readFileForHash(path)
		if err != nil {
			return false, err
		}
		rawContent = raw
	}
	sum := sha256.Sum256(rawContent)
	sha := hex.EncodeToString(sum[:])
	if err := db.Model(&LegacyMigrationMarker{}).Where("content_sha256 = ?", sha).Count(&hit).Error; err != nil {
		return false, errors.NewFromError(err)
	}
	if hit > 0 {
		ctx.LogInfof("legacy store %s already migrated (fingerprint %s…)", path, sha[:12])
		return true, nil
	}
	// Step 4: rows exist but none match — the frozen legacy store changed
	// after migration (OQ-1 posture: never re-migrate deltas).
	var any int64
	if err := db.Model(&LegacyMigrationMarker{}).Count(&any).Error; err != nil {
		return false, errors.NewFromError(err)
	}
	if any > 0 {
		ctx.LogWarnf("legacy store %s changed after migration (frozen legacy, OQ-1); not re-migrating", path)
		return true, nil
	}
	return false, nil
}

// MarkMigrated upserts on the content fingerprint (D-4): replaying a
// completed migration refreshes the row instead of duplicating it (AC-2
// leans on this).
func MarkMigrated(db *gorm.DB, m LegacyMigrationMarker) error {
	if m.ID == "" {
		m.ID = helpers.GenerateId() // callers that must report the ID (RunMigration) pre-set it
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "content_sha256"}},
		UpdateAll: true,
	}).Create(&m).Error
}

// readFileForHash is the hash-step fallback read (D-4 step 3). It is a plain
// read: banned helpers stay banned even here (D-3 covers every path through
// this package).
func readFileForHash(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewFromError(err)
	}
	return data, nil
}
