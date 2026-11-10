package serviceprovider

import (
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
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
