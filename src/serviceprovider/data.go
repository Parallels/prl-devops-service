package serviceprovider

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/database/legacy_migrator"
	"gorm.io/gorm"
)

// #426 D-3: the two runtime deprecation strings, verbatim from the record
// (versions per D-6 as amended by the 2026-11-12 owner decision: effective
// v1.2.0, removal v1.3.0). Single source of truth — tests assert these
// constants, call sites must carry no literals.
const (
	legacyDeprecationWarnMsg = "[Database] DEPRECATED: legacy JSON data layer selected (DB_TYPE=%s). " +
		"The legacy layer is deprecated since v1.2.0 and will be removed in v1.3.0 (issues #426/#417). " +
		"Plan migration to DB_TYPE=sqlite — see CHANGELOG."
	legacyFrozenWarnMsg = "[Database] DEPRECATED: legacy data file %s was already migrated and is now " +
		"FROZEN/stale — this instance serves stale data; writes here never reach the SQL database. " +
		"Switch to DB_TYPE=sqlite (#424, #426)."
	legacyProbeFailWarnMsg = "could not verify migration status: %v"
)

// warnLegacySelection fires the selection warn on every startup that ends up
// on the legacy engine, then probes #424's completion marker for the OQ-1
// frozen-legacy hazard.
//
// Boundary (pinned by task 02): the probe PRESUPPOSES a reachable SQL
// connection. Legacy mode never opens one (the #421 rollback lever), so a
// pure-legacy deployment — SQLDatabase nil, store never migrated — skips the
// frozen warn silently; that is the correct posture, not a missed warning.
// When a handle IS present (e.g. SQL init succeeded but the operator pinned
// DB_TYPE=legacy, or a test injects one), the probe reads the marker; ANY
// probe error degrades to one warn + continue — legacy must stay functional
// (PBI constraint), never fatal.
func (s *ServiceProvider) warnLegacySelection(ctx basecontext.ApiContext, filename string) {
	// Raw config value, exactly as the operator set it (both branches of
	// selectDatabaseBackend reach here: explicit DB_TYPE=legacy, and the
	// non-API SQL-failure fallback, where the shown value honestly exposes
	// the config-vs-reality gap).
	ctx.LogWarnf(legacyDeprecationWarnMsg, config.Get().DatabaseType())
	if s.SQLDatabase == nil {
		return
	}
	migrated, err := legacyStoreAlreadyMigrated(s.SQLDatabase)
	if err != nil {
		ctx.LogWarnf(legacyProbeFailWarnMsg, err)
		return
	}
	if migrated {
		ctx.LogWarnf(legacyFrozenWarnMsg, filename)
	}
}

// legacyStoreAlreadyMigrated reads #424's marker through its EXPORTED parts
// only: EnsureMarkerTable (the same idempotent ensure the migration gate
// itself performs) plus a raw row-exists count. #424's AlreadyMigrated is a
// migration gate (Detection + size/mtime + sha256 fingerprint semantics),
// not the Migrated(ctx)(bool,error)-shaped reader task 02 hoped for — per
// its sanctioned fallback, row EXISTENCE is the startup-hazard predicate;
// fingerprint logic is deliberately not copied.
func legacyStoreAlreadyMigrated(db *gorm.DB) (bool, error) {
	if err := legacy_migrator.EnsureMarkerTable(db); err != nil {
		return false, err
	}
	var n int64
	if err := db.Model(&legacy_migrator.LegacyMigrationMarker{}).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

// #421 D-6 S2 (task 17): the consumer-facing surface is the interface.
// Until S3 (task 18) the returned value is still the legacy instance; the
// type system, not a convention, now prevents access to legacy-only
// machinery (D-7 escape hatch: Legacy()).
func GetDatabaseService(ctx basecontext.ApiContext) (datainterfaces.DatabaseService, error) {
	provider := Get()
	if provider == nil {
		return nil, data.ErrDatabaseNotConnected
	}

	dbService := provider.JsonDatabase
	if dbService == nil {
		return nil, data.ErrDatabaseNotConnected
	}

	err := dbService.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return dbService, nil
}
