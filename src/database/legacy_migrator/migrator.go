package legacy_migrator

// The engine (plan D-5/D-6): detection → marker gate → read-only load →
// ONE transaction over the deterministic collection plan → marker inside the
// same tx. AC-2 rides on the upsert engine (mapping.go); AC-3 on
// gdb.Transaction; AC-4 on the per-collection info line; AC-5 on the marker;
// AC-7's dry-run semantics live here (the flag/env wiring is task 06).
//
// Until task 06 the exported entry point below has ZERO production callers
// — this whole package stays inert code behind the human gate.

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
)

func RunMigration(ctx basecontext.ApiContext, gdb *gorm.DB, opts Options) (Report, error) {
	report := Report{DryRun: opts.DryRun}
	if opts.SourcePath == "" {
		return report, errors.NewWithCode("legacy migration: SourcePath is required", 500)
	}
	chunk := opts.ChunkSize
	if chunk <= 0 {
		chunk = defaultChunkSize
	}

	det, err := DetectLegacyStore(ctx, opts.SourcePath)
	if err != nil {
		return report, err
	}
	if !det.Present {
		report.Skipped = true
		report.SkipReason = det.Reason
		ctx.LogDebugf("legacy migration skipped: %s", det.Reason) // AC-1 "silently"
		return report, nil
	}
	raw, err := os.ReadFile(opts.SourcePath) // bytes needed for the fingerprint (D-4); encrypted-safe
	if err != nil {
		return report, errors.NewFromError(err)
	}
	done, err := AlreadyMigrated(ctx, gdb, det, opts.SourcePath, raw)
	if err != nil {
		return report, err
	}
	if done {
		report.Skipped = true
		report.SkipReason = "already-migrated"
		return report, nil
	}

	legacy, err := data.LoadLegacyData(ctx, opts.SourcePath) // AD-5 — the only read path (task 02)
	if err != nil {
		return report, err
	}

	plan := collectionPlan() // deterministic, FK-safe (D-7)
	if opts.DryRun {
		return runDryRun(ctx, plan, &legacy, report) // computes mappings, WRITES NOTHING (AC-7)
	}

	marker := LegacyMigrationMarker{
		ID:                helpers.GenerateId(), // pre-set so the Report can name it (MarkMigrated keeps it)
		SourcePath:        opts.SourcePath,
		SourceSizeBytes:   det.SizeBytes,
		SourceModTimeUnix: det.ModTime.Unix(),
		ContentSHA256:     fmt.Sprintf("%x", sha256.Sum256(raw)),
		CompletedAt:       helpers.GetUtcCurrentDateTime(),
	}
	total := 0
	for _, s := range plan {
		total += s.LegacyCount(&legacy)
	}

	err = gdb.Transaction(func(tx *gorm.DB) error { // AC-3: ONE tx, all-or-nothing (D-5)
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
				if fErr := opts.FailAfter(s.Name, int(written)); fErr != nil {
					return fErr // rollback provokes (task 07)
				}
			}
		}
		marker.CollectionsTotal = written
		if uErr := MarkMigrated(tx, marker); uErr != nil {
			return uErr // marker INSIDE the tx: crash ⇒ nothing (AC-5)
		}
		report.MarkerID = marker.ID
		return nil
	})
	if err != nil {
		ctx.LogErrorf("legacy migration rolled back: %v", err)
		return Report{DryRun: opts.DryRun}, err
	}
	return report, nil
}

// runDryRun projects the migration — every mapping computed, zero writes,
// zero marker (AC-7). Counts match what a real run would attempt (relational
// specs join this projection automatically when task 04 fills the plan).
func runDryRun(ctx basecontext.ApiContext, plan []CollectionSpec, d *data.Data, report Report) (Report, error) {
	total := 0
	for _, s := range plan {
		n := s.LegacyCount(d)
		total += n
		report.Collections = append(report.Collections, CollectionReport{s.Name, n, int64(n)})
		ctx.LogInfof("legacy migration (dry-run): would migrate %d %s records", n, s.Name)
	}
	ctx.LogInfof("legacy migration (dry-run): %d records across %d collections — nothing written", total, len(plan))
	return report, nil
}
