// Package database provides a facade for managing database connections and stores.
package database

import (
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/service"
)

// Initialize bootstraps the Database service.
func Initialize(cfg common.Config) (*service.DatabaseService, error) {
	return service.Initialize(cfg)
}

// GetInstance returns the Database service singleton.
func GetInstance() *service.DatabaseService {
	return service.GetInstance()
}

// Reset clears the singleton (for tests).
func Reset() {
	service.Reset()
}
