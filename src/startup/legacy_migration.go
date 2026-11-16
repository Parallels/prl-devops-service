package startup

// PBI #424 boot hook: legacy JSON store → GORM schema, exactly once,
// idempotent (marker-gated inside the migrator). AD-4 order: runs after the
// legacy schema-migration loop and before serving; AutoMigrate already ran
// at provider-init (#419/#420). The DB_AUTO_MIGRATE gate lives at the CALL
// SITE; this helper stays policy-free and testable.

import (
	"os"
	"path/filepath"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/serviceprovider"

	"github.com/Parallels/prl-devops-service/database/legacy_migrator"
	"gorm.io/gorm"
)

// Test seams (overridden in _test.go, restored via t.Cleanup): the real
// implementations never get constructed by tests. The legacy store's
// filename is reached via Legacy().Filename() (escape hatch) or the
// folder convention (SQL mode) — see defaultLegacyStoreSource.
type migratorFn func(basecontext.ApiContext, *gorm.DB, legacy_migrator.Options) (legacy_migrator.Report, error)

var (
	getSQLDatabase               = serviceprovider.GetSQLDatabase
	legacyStoreSource            = defaultLegacyStoreSource
	runMigration      migratorFn = legacy_migrator.RunMigration
)

// defaultLegacyStoreSource resolves the legacy store path. The escape
// hatch (DB_TYPE=legacy) defers to the provider's LEGACY handle
// (Legacy()/legacyDatabase). The slot formerly used here —
// provider.JsonDatabase — is NOT the legacy store post-#421: in SQL mode
// it is the *GormDatabase engine, and its Filename() returns the SQL
// datasource (macOS matrix run b caught this: the sqlite path was stat'd
// instead of data.json ⇒ silent skip). SQL mode therefore takes the
// convention branch below.
//
// Unification posture: the SQL destination is always the unified folder
// (database.DefaultDataFolder), but the SOURCE is DISCOVERED — an existing
// install may hold data.json in any historical folder (the pre-unification
// root answer, the non-root home, or an old explicit pin), and the point of
// the migration is to import it wherever it was left behind.
func defaultLegacyStoreSource() (string, error) {
	if provider := serviceprovider.Get(); provider != nil && provider.Legacy() != nil {
		//nolint:staticcheck // #426: legacy escape hatch — remove with the legacy layer (removal tracked in #426)
		return provider.Legacy().Filename(), nil
	}
	// An explicit DATABASE_FOLDER is a pin: look exactly there, no fallback
	// (task-06 pinned posture).
	if folder := config.Get().DatabaseFolder(); folder != "" {
		return filepath.Join(folder, "data.json"), nil
	}
	for _, p := range legacyStoreCandidates() {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	// Nothing on disk: report the first candidate — the detector classifies
	// it absent and the boot skips (unchanged posture).
	candidates := legacyStoreCandidates()
	return candidates[0], nil
}

// legacyStoreCandidates is the discovery-order seam (overridden in tests).
var legacyStoreCandidates = defaultLegacyStoreCandidates

// defaultLegacyStoreCandidates lists where a data.json may live, priority
// order: unified folder → LegacyDataFolder (pre-unification root answer,
// discovery-only) → non-root home folder. De-duplicated; first existing
// file wins, so once the import lands in the unified folder the answer is
// stable across reboots.
func defaultLegacyStoreCandidates() []string {
	folders := []string{database.DefaultDataFolder(), database.LegacyDataFolder}
	if home, err := os.UserHomeDir(); err == nil {
		folders = append(folders, filepath.Join(home, ".parallels-devops"))
	}
	seen := make(map[string]bool, len(folders))
	out := make([]string, 0, len(folders))
	for _, f := range folders {
		p := filepath.Join(f, "data.json")
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

func runLegacyDataMigration(ctx basecontext.ApiContext) error {
	gdb, err := getSQLDatabase(ctx) // #418 accessor; nil/err ⇒ #418 policy already ruled
	if err != nil {
		return err
	}
	sourcePath, err := legacyStoreSource()
	if err != nil {
		return err
	}
	report, err := runMigration(ctx, gdb, legacy_migrator.Options{
		SourcePath: sourcePath,
		DryRun:     legacy_migrator.IsDryRunRequested(),
	})
	if err != nil {
		return err
	}
	if report.DryRun {
		legacy_migrator.ReportDryRun(ctx, report)
	}
	return nil
}
