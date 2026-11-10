// Package database implements the GORM-based persistence layer of the service
// (feature #417). This file is created by the dependency-foundation task; the
// connection factory, config contract, health checks and logger bridge arrive
// in the follow-up tasks of the same PBI.
//
// Driver note (feature decision AD-2): the SQLite dialect is the pure-Go
// github.com/glebarez/sqlite (modernc.org/sqlite engine) instead of the CGO
// gorm.io/driver/sqlite named in issue #418, keeping the project CGO-free for
// cross-compiled release binaries.
package database

import (
	_ "github.com/glebarez/sqlite"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/gorm"
)
