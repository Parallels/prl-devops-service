package database

import (
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrUnsupportedDatabaseType = errors.NewWithCode("unsupported database type, expected 'sqlite' or 'postgres'", 400)
	ErrPostgresDSNRequired     = errors.NewWithCode("database type 'postgres' requires a connection string (DSN)", 400)
	ErrSQLitePathRequired      = errors.NewWithCode("database type 'sqlite' requires a file path", 400)
	ErrDatabaseUnreachable     = errors.NewWithCode("database is unreachable", 503)
)

// mapGormError maps GORM/driver failures onto legacy sentinel identity (D-3,
// plan D-11) so controllers keep seeing the exact sentinels they compare on
// today. notFound/dup may be nil when the legacy op has no sentinel for that
// failure mode — a generic 500 wrap is returned instead.
func mapGormError(err error, notFound, dup *errors.SystemError) error {
	if err == nil {
		return nil
	}
	if stderrors.Is(err, gorm.ErrRecordNotFound) && notFound != nil {
		return notFound
	}
	if isUniqueViolation(err) && dup != nil {
		return dup
	}
	return errors.NewWithCode(fmt.Sprintf("database operation failed: %v", err), 500)
}

// isUniqueViolation recognises constraint errors across drivers: sqlite
// surfaces "UNIQUE constraint failed: …", the gorm translate driver
// pre-maps ErrDuplicatedKey, and PostgreSQL reports SQLSTATE 23505 on
// *pgconn.PgError (the typed check task-18/T2 replaced the interim
// "duplicate key value" string match with — kept as a belt for drivers
// that surface the message without the typed error).
func isUniqueViolation(err error) bool {
	if stderrors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	if stderrors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "duplicate key value")
}
