package database

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newPostgresConnection configures a pooled postgres handle. Reality check
// against the spec assumption: gorm.io/driver/postgres v1.6 DIALS during
// gorm.Open (pgx initial connect), so network failures surface there; the
// Ping below stays as the engine-agnostic reachability gate (task 05 upgrades
// it to PingWithRetry; TODO at the call site). The canonical
// gorm.io/driver/postgres is deliberate: the pure-Go constraint (AD-2) is
// SQLite-only.
func newPostgresConnection(ctx basecontext.ApiContext, cfg DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), gormConfig(ctx))
	if err != nil {
		// Open dials (see above) — unreachable and misconfiguration both
		// arrive here at boot; the sentinel wrap keeps #421's errors.Is
		// branching working for either shape. The driver message goes
		// through scrubDriverError: pgx echoes unparseable inputs verbatim
		// (T16) and its own password masking is token-shape-bound.
		return nil, fmt.Errorf("%w (%s): %s", ErrDatabaseUnreachable,
			RedactDSN(cfg.DSN), scrubDriverError(err.Error(), cfg.DSN))
	}

	if err := applyPool(db, cfg); err != nil {
		Close(ctx, db)
		return nil, errors.NewFromError(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		Close(ctx, db)
		return nil, errors.NewFromError(err)
	}

	if err := PingWithRetry(ctx, sqlDB, DefaultPingAttempts, DefaultPingDelay); err != nil {
		Close(ctx, db)
		return nil, fmt.Errorf("%w (%s)", err, RedactDSN(cfg.DSN))
	}

	ctx.LogInfof("PostgreSQL connection established (pool: maxOpen=%d maxIdle=%d lifetime=%s)",
		cfg.Postgres.MaxOpen(), cfg.Postgres.MaxIdle(), cfg.Postgres.Lifetime()) // no DSN, ever
	return db, nil
}

// applyPool is the offline-verifiable pool application (task 02 getters supply
// zero-means-default values). Kept a pure function so unit tests assert the
// applied settings without a live server.
func applyPool(db *gorm.DB, cfg DatabaseConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	pool := cfg.Postgres
	sqlDB.SetMaxOpenConns(pool.MaxOpen())
	sqlDB.SetMaxIdleConns(pool.MaxIdle())
	sqlDB.SetConnMaxLifetime(pool.Lifetime())
	if d := pool.IdleTime(); d > 0 {
		sqlDB.SetConnMaxIdleTime(d)
	}

	return nil
}
