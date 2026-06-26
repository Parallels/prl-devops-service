package common

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	// ErrRecordNotFound is returned when a record is not found in the database.
	ErrRecordNotFound = errors.New("record not found")

	// ErrDuplicateKey is returned when a unique constraint is violated.
	ErrDuplicateKey = errors.New("duplicate key violation")

	// ErrForeignKeyViolation is returned when a foreign key constraint is violated.
	ErrForeignKeyViolation = errors.New("foreign key violation")

	// ErrDatabaseConnection is returned when the database connection fails.
	ErrDatabaseConnection = errors.New("database connection error")
)

// MapError translates database-specific errors into domain errors.
func MapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateKey
	}

	// PostgreSQL specific error handling
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrDuplicateKey
		case "23503": // foreign_key_violation
			return ErrForeignKeyViolation
		}
	}

	// Generic fallback checks if driver-specific checks fail
	if strings.Contains(err.Error(), "duplicate key value") {
		return ErrDuplicateKey
	}

	return err
}

// IsRecordNotFound returns true if the error is a record not found error.
func IsRecordNotFound(err error) bool {
	return errors.Is(err, ErrRecordNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}
