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
	"github.com/Parallels/prl-devops-service/errors"
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

// defaultLegacyStoreSource resolves <dbLocation>/data.json. The escape
// hatch (DB_TYPE=legacy) defers to the provider's LEGACY handle
// (Legacy()/legacyDatabase). The slot formerly used here —
// provider.JsonDatabase — is NOT the legacy store post-#421: in SQL mode
// it is the *GormDatabase engine, and its Filename() returns the SQL
// datasource (macOS matrix run b caught this: the sqlite path was stat'd
// instead of data.json ⇒ silent skip). SQL mode therefore takes the
// convention branch: database.defaultDatabasePath's documented AD-3
// duplication of the serviceprovider location convention
// (DATABASE_FOLDER → /etc/... → home).
func defaultLegacyStoreSource() (string, error) {
	if provider := serviceprovider.Get(); provider != nil && provider.Legacy() != nil {
		//nolint:staticcheck // #426: legacy escape hatch — remove with the legacy layer (removal issue TBD-04)
		return provider.Legacy().Filename(), nil
	}
	folder := config.Get().DatabaseFolder()
	if folder == "" {
		if os.Getuid() == 0 {
			folder = "/etc/parallels-devops"
		} else if home, err := os.UserHomeDir(); err == nil {
			folder = filepath.Join(home, ".parallels-devops")
		} else {
			return "", errors.NewWithCode("cannot resolve legacy store location: no DATABASE_FOLDER and no user home", 500)
		}
	}
	return filepath.Join(folder, "data.json"), nil
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
