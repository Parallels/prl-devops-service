package legacy_migrator

// Migration surface (plan D-5/D-6): the options the startup hook (task 06)
// passes in, the report the callers log/return, and the shared constants.
// Kept separate from migrator.go (loop) and mapping.go (per-collection
// knowledge) per plan D-1.

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
)

type Options struct {
	SourcePath string
	DryRun     bool
	ChunkSize  int                               // <=0 ⇒ defaultChunkSize
	FailAfter  func(name string, done int) error // nil in production; task-07 seam: return err
	// right after collection `name` wrote `done` rows
}

type CollectionReport struct {
	Name         string
	LegacyCount  int
	WrittenCount int64
}

type Report struct {
	Skipped     bool
	SkipReason  string // "absent", "empty", "already-migrated"
	DryRun      bool
	Collections []CollectionReport
	MarkerID    string
	Validation  []ValidationIssue // AC-6: empty on the happy path; warn-only by construction (task 05)
}

const defaultChunkSize = 500 // R-2: bounded statement size & log volume for the jobs table

// BuildLargeJobsFixture writes a legacy-format file with `count` synthetic
// jobs (D-10: generated, never committed) into dir and returns its path.
// Exported for #425 (chunking/perf reuse); test-support only — production
// code never calls it.
func BuildLargeJobsFixture(tb testing.TB, dir string, count int) string {
	tb.Helper()
	d := data.Data{}
	for i := 0; i < count; i++ {
		d.Jobs = append(d.Jobs, models.Job{
			ID:           fmt.Sprintf("j-%05d", i),
			Owner:        "u-load",
			OwnerName:    "Load Tester",
			State:        "finished",
			JobType:      "vm",
			JobOperation: "bulk",
			Progress:     100,
			CreatedAt:    "2024-06-01 00:00:00",
			UpdatedAt:    "2024-06-01 00:00:00",
		})
	}
	raw, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		tb.Fatalf("marshal large fixture: %v", err)
	}
	path := fmt.Sprintf("%s/large_jobs_%d.json", dir, count)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		tb.Fatalf("write large fixture: %v", err)
	}
	return path
}
