package connection

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Parallels/prl-devops-service/database/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

// initializeSQLite initializes SQLite database connection
func initializeSQLite(config *common.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	storagePath := config.SQLite.StoragePath

	// Convert to absolute path if relative
	absPath := filepath.Join(storagePath, config.SQLite.FileName)
	absPath, err := filepath.Abs(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Ensure the directory exists
	if err := ensureDir(absPath); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", fmt.Sprintf("file:%s?cache=shared&mode=rwc", absPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(config.Pool.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Pool.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.Pool.ConnMaxLifetime)

	return db, nil
}

// ensureDir creates the directory for the database file if it doesn't exist
func ensureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return createDirIfNotExists(dir)
}

// createDirIfNotExists creates a directory if it doesn't exist
func createDirIfNotExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}
