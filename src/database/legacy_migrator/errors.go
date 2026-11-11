// Package legacy_migrator — PBI #424 one-time migration of the legacy JSON
// store into the GORM schema. One concern per file (plan D-1); import graph
// stays acyclic: this package may import database (handle only), data (the
// read-only loader), data/models — and NEVER serviceprovider (task 06 wires
// it from the outside).
//
// Banned by review gate (D-3): JsonDatabase.IsDataFileEmpty CREATES the
// file when missing (src/data/main.go:645-655) — detection is os.Stat only;
// and no JsonDatabase construction anywhere in this package.
package legacy_migrator

import "github.com/Parallels/prl-devops-service/errors"

var (
	ErrNoLegacyFile       = errors.NewWithCode("legacy data file not found", 404)
	ErrLegacyFileEmpty    = errors.NewWithCode("legacy data file is empty", 404)
	ErrMarkerCheck        = errors.NewWithCode("legacy migration marker check failed", 500)
	ErrMarkerTableMissing = errors.NewWithCode("legacy_migration_marker table missing; run with DB_AUTO_MIGRATE=true", 500)
)
