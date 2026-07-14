// Package database provides a facade for managing database connections and stores.
package database

import (
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/connection"
)

// Initialize bootstraps the Database connection service.
func Initialize(cfg common.Config) (*connection.DatabaseService, error) {
	return connection.Initialize(cfg)
}

// GetInstance returns the Database connection service singleton.
func GetInstance() *connection.DatabaseService {
	return connection.GetInstance()
}

// Reset clears the singleton (for tests).
func Reset() {
	connection.Reset()
}
