package serviceprovider

import (
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/database"
	"gorm.io/gorm"
)

// initGormConnection opens the SQL backend via the #418 factory and stores it
// on the provider. Failure leaves SQLDatabase nil and returns the error — the
// policy (exit vs warn) belongs to the call site, see mustInitGormConnection.
//
// PBI #420: operator settings (DB_TYPE / DB_PATH / DB_DSN / DB_AUTO_MIGRATE —
// flat config-file keys db_type/… per #420 plan D-1) now drive the connection;
// unset, FromApplicationConfig returns DefaultDatabaseConfig byte-identically.
func (s *ServiceProvider) initGormConnection(ctx basecontext.ApiContext) error {
	cfg, err := database.FromApplicationConfig()
	if err != nil {
		s.SQLDatabase = nil
		return err // config errors carry no DSN (#420 D-8; task 03 matrix)
	}

	db, err := database.NewDatabaseConnection(ctx, cfg)
	if err != nil {
		s.SQLDatabase = nil
		return err
	}

	s.SQLDatabase = db
	ctx.LogInfof("SQL database connection ready (type: %s)", cfg.Type)
	return nil
}

// mustInitGormConnection is the D-7 failure policy: the API command cannot
// serve data without a database, so it logs the (already redacted — the
// factory redacts) error and exits 1; every other command continues without
// SQL. Exit lives here, not in initGormConnection, so the init itself stays
// testable without process death.
func (s *ServiceProvider) mustInitGormConnection(ctx basecontext.ApiContext) {
	if err := s.initGormConnection(ctx); err != nil {
		ctx.LogErrorf("SQL database unavailable: %v", err)
		if config.Get().GetRunningCommand() == constants.API_COMMAND {
			os.Exit(1) // D-7: sole os.Exit in this package, grep-audited
		}
		ctx.LogWarnf("continuing without SQL connection (non-API command)")
	}
}

// GetSQLDatabase is the accessor #421 will route all data access through.
// A missing provider or unopened connection is ErrDatabaseUnreachable —
// callers never see a nil-without-error handle.
// TODO(#421): graceful shutdown hook — Close on the shutdown path (none exists
// today; the OS reaps handles meanwhile).
func GetSQLDatabase(ctx basecontext.ApiContext) (*gorm.DB, error) {
	if globalProvider == nil || globalProvider.SQLDatabase == nil {
		return nil, database.ErrDatabaseUnreachable
	}
	return globalProvider.SQLDatabase, nil
}

// GetSQLDatabaseType reports the engine actually driving the live SQL
// connection ("sqlite" / "postgres"), or "" when no SQL backend is running
// (legacy JSON mode, or a failed open on a non-API command). Derived from the
// connection's dialector, so it answers "what is running", not "what was
// configured" — which is exactly what the sql_database module report needs.
func GetSQLDatabaseType() string {
	if globalProvider == nil || globalProvider.SQLDatabase == nil || globalProvider.SQLDatabase.Dialector == nil {
		return ""
	}
	return globalProvider.SQLDatabase.Dialector.Name()
}

// EnabledModulesForReport returns the operator-facing module list for
// reporting surfaces (hardware endpoint, host records). It mirrors
// GetEnabledModules and appends the runtime-derived sql_database module when
// a SQL connection is live, and placement when capacity admission is
// enforcing. Deliberately NOT persisted through
// EnableModule/DisableModule: both reflect what the process is actually doing,
// not something operators toggle.
func EnabledModulesForReport() []string {
	cfg := config.Get()
	if cfg == nil {
		return []string{}
	}
	modules := cfg.GetEnabledModules()
	if GetSQLDatabaseType() != "" {
		modules = append(modules, constants.SQL_DATABASE_MODE)
		// Placement is reported when it is actually guarding dispatch, not when
		// its flag is merely true: without the SQL data layer the reservation
		// ledger cannot exist and admission is skipped, so an operator reading
		// this list to find out whether the over-commit bug is closed would be
		// told "yes" by a switch with no effect. Same runtime-derived reasoning
		// as sql_database above, and the same reason it is not in VALID_MODULES.
		if cfg.IsOrchestrator() && cfg.PlacementReservationEnabled() {
			modules = append(modules, constants.PLACEMENT_MODE)
		}
	}
	return modules
}

// selectDatabaseBackend is the #421 D-6 S3 flip (task 18). The raw
// DB_TYPE/db_type value selects the engine BEFORE FromApplicationConfig
// runs — by design, that parser rejects the legacy value (its allowed set
// is sqlite|postgres), so reading it here is the only way the documented
// `DB_TYPE=legacy` rollback lever can work at all.
//
// legacy mode: byte-identical to pre-flip — the provider builds the JSON
// engine, Connect starts its goroutines (save queue, stale-job ticker,
// backup) inside, and no SQL connection is attempted (mustInitGormConnection
// would exit(1) on the rejected legacy type on API boots).
//
// SQL mode (default: DB_TYPE unset ⇒ sqlite via #420 defaults): GormDatabase
// over the #418 connection becomes the data surface with a NIL legacy
// handle — NewJsonDatabase never runs, so zero legacy goroutines exist
// (D-7/D-8; the SQL stale-job watcher is started by startup, not here).
// A non-API command with a broken SQL config falls back to legacy,
// mirroring mustInitGormConnection's non-API posture.
func (s *ServiceProvider) selectDatabaseBackend(ctx basecontext.ApiContext, newLegacy func() *data.JsonDatabase) {
	if isLegacyBackendSelected() {
		// #426 D-3: deprecation warn (+ OQ-1 frozen probe) landed here —
		// warn-only additive, selection behaviour unchanged.
		legacyDb := newLegacy()
		s.setJsonDatabaseLegacy(legacyDb)
		_ = legacyDb.Connect(ctx)
		s.warnLegacySelection(ctx, legacyDb.Filename())
		return
	}
	s.mustInitGormConnection(ctx) // API command exits(1) on broken SQL config
	if s.SQLDatabase != nil {
		// DB_AUTO_MIGRATE gates the production schema application
		// (database.AllGormModels is the registry; default ON per AD-3).
		// Failure reuses the connection-failure D-7 posture verbatim:
		// API exits, other commands fall through to the legacy fallback.
		if am, amErr := config.Get().DatabaseAutoMigrate(); amErr == nil && am {
			if err := database.RunAutoMigrate(ctx, s.SQLDatabase); err != nil {
				ctx.LogErrorf("SQL AutoMigrate failed: %v", err)
				if config.Get().GetRunningCommand() == constants.API_COMMAND {
					os.Exit(1)
				}
				s.SQLDatabase = nil
			}
		}
		if s.SQLDatabase != nil {
			s.JsonDatabase = database.NewGormDatabase(ctx, s.SQLDatabase)
			s.legacyDatabase = nil
			return
		}
	}
	legacyDb := newLegacy()
	s.setJsonDatabaseLegacy(legacyDb)
	_ = legacyDb.Connect(ctx)
	s.warnLegacySelection(ctx, legacyDb.Filename())
}

// isLegacyBackendSelected reads the raw selection exactly as
// FromApplicationConfig reads it (lower+trim; env > flag > config key).
func isLegacyBackendSelected() bool {
	return strings.EqualFold(strings.TrimSpace(config.Get().DatabaseType()), database.DBTypeLegacy)
}
