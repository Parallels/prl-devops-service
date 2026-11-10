package database

import "github.com/Parallels/prl-devops-service/errors"

var (
	ErrUnsupportedDatabaseType = errors.NewWithCode("unsupported database type, expected 'sqlite' or 'postgres'", 400)
	ErrPostgresDSNRequired     = errors.NewWithCode("database type 'postgres' requires a connection string (DSN)", 400)
	ErrSQLitePathRequired      = errors.NewWithCode("database type 'sqlite' requires a file path", 400)
	ErrDatabaseUnreachable     = errors.NewWithCode("database is unreachable", 503)
)
