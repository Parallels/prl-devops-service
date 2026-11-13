package legacy_migrator

// AC-6 (plan D-8): post-commit count comparison — legacy side in memory
// (len(Data.<Collection>) via the plan's LegacyCount accessors), SQL side
// via each spec's Validate formula (the #419 validation-column expressions,
// join-table-aware: user_roles ⇒ Σ len(u.Roles) etc).
//
// WARN-ONLY is literal AC wording: a mismatch (or even a broken Count) must
// NEVER fail the migration — the data already committed, and bricking boot
// over a counter would be worse than the bug the counter reports. Hence: no
// error return, issues recorded and returned for the caller's report.

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"gorm.io/gorm"
)

type ValidationIssue struct {
	Collection string
	Expected   int64
	Actual     int64 // -1 ⇒ the count query itself failed (also warn-only)
}

// ValidateCounts compares in-memory legacy counts against SQL counts via
// the plan's formulas. Runs AFTER db.Transaction success — COUNT sees the
// committed truth exactly.
func ValidateCounts(ctx basecontext.ApiContext, db *gorm.DB, legacy *data.Data, plan []CollectionSpec) []ValidationIssue {
	var issues []ValidationIssue
	for _, s := range plan {
		expected, actual, err := s.Validate(db, legacy)
		if err != nil {
			ctx.LogWarnf("legacy migration validation: %s count query failed: %v", s.Name, err)
			issues = append(issues, ValidationIssue{s.Name, expected, -1})
			continue
		}
		want := int64(s.LegacyCount(legacy))
		if want != expected { // defensive: formula vs accessor drift is itself a finding
			ctx.LogWarnf("legacy migration validation: %s accessor=%d formula=%d — mapping doc drift?", s.Name, want, expected)
		}
		if want != actual {
			ctx.LogWarnf("legacy migration validation MISMATCH %s: legacy=%d new=%d — investigate (migration NOT failed by validation)", s.Name, want, actual)
			issues = append(issues, ValidationIssue{s.Name, want, actual})
			continue
		}
		ctx.LogInfof("legacy migration validation ok: %s %d/%d", s.Name, actual, want)
	}
	if len(issues) > 0 {
		ctx.LogWarnf("legacy migration validation finished with %d mismatched collection(s)", len(issues))
	} else {
		ctx.LogInfof("legacy migration validation finished: all collections matched")
	}
	return issues
}
