package legacy_migrator

// AC-7 input resolution + operator report. Lives in the migrator package so
// the startup hook stays a three-liner; the flag/env precedence mirrors
// Config.GetKey (src/config/main.go): env var beats CLI flag.

import (
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/cjlapao/common-go/helper"
)

// IsDryRunRequested resolves AC-7's input. Accepted truthy values:
// "true","1","yes" (case-insensitive). helper.GetFlagValue needs a VALUE
// in --flag=value form (the space-separated "--flag true" form is NOT
// parsed by common-go — pinned by test), so the operator spelling is
// --migrate-dry-run=true; a value-less flag is inert, exactly like every
// other repo flag (documented for #427).
func IsDryRunRequested() bool {
	flag := strings.TrimSpace(helper.GetFlagValue(constants.MIGRATE_DRY_RUN_FLAG, ""))
	if v, ok := os.LookupEnv(constants.MIGRATE_DRY_RUN_ENV_VAR); ok {
		flag = v
	}
	switch strings.ToLower(strings.TrimSpace(flag)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

// ReportDryRun prints the AC-7 report; called for report.DryRun.
func ReportDryRun(ctx basecontext.ApiContext, r Report) {
	ctx.LogInfof("[dry-run] legacy migration NOT executed; %d collection(s) would be written:", len(r.Collections))
	for _, c := range r.Collections {
		ctx.LogInfof("[dry-run] would write %d record(s) into %s", c.WrittenCount, c.Name)
	}
	ctx.LogInfof("[dry-run] no data was written (no marker; data.json untouched); migration did NOT run")
}
