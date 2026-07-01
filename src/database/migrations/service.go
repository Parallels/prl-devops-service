// Package migrations provides database seeding and migration functionality
package migrations

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/migrations/interfaces"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	instance *MigrationService
	once     sync.Once
)

// MigrationRecord tracks which migrations have been applied
type MigrationRecord struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(64)"`
	Name        string    `json:"name" gorm:"column:name;type:varchar(255);not null;uniqueIndex"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	AppliedAt   time.Time `json:"applied_at" gorm:"column:applied_at;autoCreateTime"`
	Status      string    `json:"status" gorm:"column:status;type:varchar(50);not null;default:'applied'"`
	Error       string    `json:"error,omitempty" gorm:"column:error;type:text"`
}

func (MigrationRecord) TableName() string {
	return "_migrations"
}

// MigrationService manages database migrations
type MigrationService struct {
	db      *gorm.DB
	workers []interfaces.MigrationWorker
	mu      sync.RWMutex
	applied map[string]bool
}

// Initialize initializes the migration service (singleton)
func Initialize(db *gorm.DB) *MigrationService {
	once.Do(func() {
		instance = NewMigrationService(db)
	})
	return instance
}

// GetInstance returns the singleton instance
func GetInstance() *MigrationService {
	return instance
}

// NewMigrationService creates a new migration service
func NewMigrationService(db *gorm.DB) *MigrationService {
	logger := logging.Get()
	service := &MigrationService{
		db:      db,
		workers: make([]interfaces.MigrationWorker, 0),
		applied: make(map[string]bool),
	}

	// Initialize migrations tracking table
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		logger.Error("Failed to create migrations table: %v", err)
		return service
	}

	// Load already-applied migrations
	service.loadAppliedMigrations()

	return service
}

// loadAppliedMigrations loads the list of already applied migrations
func (s *MigrationService) loadAppliedMigrations() {
	var records []MigrationRecord
	if err := s.db.Find(&records).Error; err != nil {
		return // Table might not exist yet
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, record := range records {
		if record.Status == "applied" {
			s.applied[record.Name] = true
		}
	}
}

// Register registers a migration worker
func (s *MigrationService) Register(worker interfaces.MigrationWorker) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already registered
	for _, existing := range s.workers {
		if existing.GetName() == worker.GetName() {
			return
		}
	}

	s.workers = append(s.workers, worker)
}

// RunAll executes all registered migrations that haven't been applied yet
func (s *MigrationService) RunAll(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("run_all_migrations")
	logger := logging.Get()

	s.mu.RLock()
	workers := make([]interfaces.MigrationWorker, len(s.workers))
	copy(workers, s.workers)
	s.mu.RUnlock()

	// Sort by execution order
	sort.Slice(workers, func(i, j int) bool {
		return workers[i].GetOrder() < workers[j].GetOrder()
	})

	logger.Info("Running database migrations...")
	logger.Info("Total migrations to check: %d", len(workers))

	for i, worker := range workers {
		name := worker.GetName()

		// Skip if already applied
		if s.isApplied(name) {
			logger.Debug("Migration '%s' already applied, skipping", name)
			continue
		}

		logger.Info("[%d/%d] Running migration: %s", i+1, len(workers), name)
		logger.Info("  Description: %s", worker.GetDescription())

		// Run the migration
		runDiag := worker.Run(ctx)
		if runDiag.HasErrors() {
			logger.Error("Migration '%s' failed: %v", name, runDiag.GetSummary())
			s.recordMigration(name, worker.GetDescription(), "failed", runDiag.GetSummary())
			diag.Append(runDiag)
			return diag
		}

		// Record success
		s.recordMigration(name, worker.GetDescription(), "applied", "")
		s.markAsApplied(name)
		logger.Info("✓ Migration '%s' completed successfully", name)
	}

	logger.Info("All migrations completed successfully")
	return diag
}

// isApplied checks if a migration has been applied
func (s *MigrationService) isApplied(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.applied[name]
}

// markAsApplied marks a migration as applied
func (s *MigrationService) markAsApplied(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applied[name] = true
}

// recordMigration records a migration result in the database
func (s *MigrationService) recordMigration(name, description, status, errMsg string) {
	// Remove any existing record
	s.db.Where("name = ?", name).Delete(&MigrationRecord{})

	record := MigrationRecord{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Status:      status,
		Error:       errMsg,
	}

	if err := s.db.Create(&record).Error; err != nil {
		logger := logging.Get()
		logger.Error("Failed to record migration '%s': %v", name, err)
	}
}

// GetAppliedMigrations returns a list of applied migration names
func (s *MigrationService) GetAppliedMigrations() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	applied := make([]string, 0, len(s.applied))
	for name := range s.applied {
		applied = append(applied, name)
	}
	return applied
}

// GetPendingMigrations returns a list of pending migration names
func (s *MigrationService) GetPendingMigrations() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pending := make([]string, 0)
	for _, worker := range s.workers {
		name := worker.GetName()
		if !s.applied[name] {
			pending = append(pending, name)
		}
	}
	return pending
}

// Rollback rolls back a specific migration
func (s *MigrationService) Rollback(ctx basecontext.BaseContext, migrationName string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("rollback_migration")
	logger := logging.Get()

	// Find the worker
	s.mu.RLock()
	var worker interfaces.MigrationWorker
	for _, w := range s.workers {
		if w.GetName() == migrationName {
			worker = w
			break
		}
	}
	s.mu.RUnlock()

	if worker == nil {
		diag.AddError("worker_not_found", fmt.Sprintf("migration worker '%s' not found", migrationName), "migration_service", nil)
		return diag
	}

	// Check if applied
	if !s.isApplied(migrationName) {
		diag.AddError("migration_not_applied", fmt.Sprintf("migration '%s' has not been applied", migrationName), "migration_service", nil)
		return diag
	}

	logger.Info("Rolling back migration: %s", migrationName)

	// Run rollback
	rollbackDiag := worker.Rollback(ctx)
	if rollbackDiag != nil && rollbackDiag.HasErrors() {
		logger.Error("Rollback of '%s' failed: %v", migrationName, rollbackDiag.GetSummary())
		diag.Append(rollbackDiag)
		return diag
	}

	// Remove from database and memory
	s.db.Where("name = ?", migrationName).Delete(&MigrationRecord{})

	s.mu.Lock()
	delete(s.applied, migrationName)
	s.mu.Unlock()

	logger.Info("✓ Rollback of '%s' completed successfully", migrationName)
	return diag
}
