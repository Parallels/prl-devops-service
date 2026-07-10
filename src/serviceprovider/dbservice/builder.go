package dbservice

import (
	"context"
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/service"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	"gorm.io/gorm"
)

// Builder provides a fluent interface for constructing a DatabaseService
type Builder struct {
	cfg             *config.Config
	ctx             basecontext.ApiContext
	skipMigration   bool
	skipHealthCheck bool
	customModels    []interface{}
}

// NewBuilder creates a new DatabaseService builder
func NewBuilder(cfg *config.Config, ctx basecontext.ApiContext) *Builder {
	return &Builder{
		cfg: cfg,
		ctx: ctx,
	}
}

// SkipMigration disables auto-migration during initialization
func (b *Builder) SkipMigration() *Builder {
	b.skipMigration = true
	return b
}

// SkipHealthCheck disables health check during initialization
func (b *Builder) SkipHealthCheck() *Builder {
	b.skipHealthCheck = true
	return b
}

// WithModels adds custom models for auto-migration
func (b *Builder) WithModels(models ...interface{}) *Builder {
	b.customModels = append(b.customModels, models...)
	return b
}

// Build constructs and initializes the DatabaseService
func (b *Builder) Build() (*DatabaseService, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("database_builder")

	// Step 1: Build database configuration
	dbConfig := buildDatabaseConfig(b.cfg)

	// Step 2: Initialize centralized database service
	dbSvc, err := service.Initialize(dbConfig)
	if err != nil {
		diag.AddError("db_init_failed", fmt.Sprintf("failed to initialize database service: %v", err), "builder", nil)
		return nil, diag
	}

	db := dbSvc.GetDB()
	if db == nil {
		diag.AddError("db_nil", "database connection is nil", "builder", nil)
		return nil, diag
	}

	// Step 3: Health check (optional)
	if !b.skipHealthCheck {
		if err := dbSvc.Health(context.Background()); err != nil {
			diag.AddError("health_check_failed", fmt.Sprintf("database health check failed: %v", err), "builder", nil)
			b.ctx.LogErrorf("Database health check failed: %v", err)
			return nil, diag
		}
	}

	// Step 4: Configure logger
	logLevel := service.ConvertLogLevel(b.cfg.IsDebugEnabled())
	customLogger := service.NewBaseContextLogger(b.ctx, logLevel)
	db.Logger = customLogger

	// Step 5: Auto-migrate (optional)
	if !b.skipMigration {
		if err := b.migrate(db); err != nil {
			diag.AddError("migration_failed", fmt.Sprintf("auto-migration failed: %v", err), "builder", nil)
			return nil, diag
		}
	}

	// Step 6: Initialize store registry
	stores, err := NewStoreRegistry(db)
	if err != nil {
		diag.AddError("store_init_failed", fmt.Sprintf("failed to initialize stores: %v", err), "builder", nil)
		return nil, diag
	}

	// Step 7: Create database service
	dbService := &DatabaseService{
		db:     db,
		stores: stores,
	}

	b.ctx.LogInfof("Database initialized successfully (type: %s)", dbConfig.Type)
	return dbService, nil
}

// migrate runs auto-migration for all models
func (b *Builder) migrate(db *gorm.DB) error {
	// Default models
	models := []interface{}{
		&models.User{},
		&models.Role{},
		&models.Claim{},
		&models.ApiKey{},
		&models.Configuration{},
		&models.Activity{},
	}

	// Add custom models
	models = append(models, b.customModels...)

	return db.AutoMigrate(models...)
}

// buildDatabaseConfig converts app config to database common.Config
func buildDatabaseConfig(cfg *config.Config) common.Config {
	dbType := cfg.DatabaseType()

	dbConfig := common.Config{
		Type:    common.DatabaseType(dbType),
		Debug:   cfg.IsDebugEnabled(),
		Migrate: cfg.IsDatabaseAutoMigrateEnabled(),
		Pool:    common.DefaultPoolConfig(),
	}

	switch dbConfig.Type {
	case common.SQLite:
		dbPath := "data"
		if cfg.DatabaseFolder() != "" {
			dbPath = cfg.DatabaseFolder()
		}
		dbConfig.SQLite = common.SQLiteConfig{
			StoragePath: dbPath,
			FileName:    "database.db",
		}

	case common.PostgreSQL:
		dbConfig.PostgreSQL = common.PostgreSQLConfig{
			Host:     cfg.GetKey("DATABASE_HOST"),
			Port:     cfg.GetIntKey("DATABASE_PORT"),
			Database: cfg.GetKey("DATABASE_NAME"),
			Username: cfg.GetKey("DATABASE_USERNAME"),
			Password: cfg.GetKey("DATABASE_PASSWORD"),
			SSLMode:  cfg.GetBoolKey("DATABASE_SSL_MODE"),
		}
		// Apply defaults
		if dbConfig.PostgreSQL.Host == "" {
			dbConfig.PostgreSQL.Host = "localhost"
		}
		if dbConfig.PostgreSQL.Port == 0 {
			dbConfig.PostgreSQL.Port = 5432
		}
		if dbConfig.PostgreSQL.Database == "" {
			dbConfig.PostgreSQL.Database = "prl_devops"
		}
		if dbConfig.PostgreSQL.Username == "" {
			dbConfig.PostgreSQL.Username = "postgres"
		}
	}

	return dbConfig
}
