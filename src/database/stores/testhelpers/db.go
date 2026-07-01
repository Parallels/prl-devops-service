package testhelpers

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/data/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewTestDB creates an in-memory SQLite database for testing
// Each test gets a unique isolated database to prevent test interference
func NewTestDB(t *testing.T) *gorm.DB {
	// Use a unique database name for each test to prevent cross-test pollution
	dbName := fmt.Sprintf("file:test_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), rand.Int())

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Quiet during tests
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		// Core models
		&models.User{},
		&models.Role{},
		&models.Claim{},
		&models.Activity{},
		&models.ActivitySummary{},

		// API & Authentication
		&models.ApiKey{},
		&models.WebAuthnCredential{},

		// Configuration & System
		&models.Configuration{},
		&models.EmailTemplate{},
		&models.Notification{},
		&models.Message{},
		&models.Worker{},

		// Security
		&models.IpBan{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupDB cleans up test database resources
func CleanupDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil && sqlDB != nil {
		sqlDB.Close()
	}
}

// NewTestDBWithModels creates a test database and migrates only the specified models
// Useful for testing specific stores without migrating all tables
func NewTestDBWithModels(t *testing.T, models ...interface{}) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if len(models) > 0 {
		err = db.AutoMigrate(models...)
		if err != nil {
			t.Fatalf("Failed to migrate test database: %v", err)
		}
	}

	return db
}
