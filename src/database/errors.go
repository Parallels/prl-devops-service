package database

import "github.com/Parallels/prl-devops-service/errors"

var (
	ErrUnsupportedDatabaseType = errors.NewWithCode("unsupported database type, expected 'sqlite' or 'postgres'", 400)
	ErrPostgresDSNRequired     = errors.NewWithCode("database type 'postgres' requires a connection string (DSN)", 400)
	ErrSQLitePathRequired      = errors.NewWithCode("database type 'sqlite' requires a file path", 400)
	ErrDatabaseUnreachable     = errors.NewWithCode("database is unreachable", 503)
	// ErrPostgresNotWiredYet is an explicit, greppable interim error: the postgres
	// case of the factory switch exists since task 03 and must never silently
	// fall back to sqlite. Task 04 (slice 418-04) removes it along with its case.
	ErrPostgresNotWiredYet = errors.NewWithCode("postgres support arrives in PBI #418 task 04; not wired yet", 501)
)
