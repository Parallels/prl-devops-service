package connection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/database/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// initializePostgreSQL initializes PostgreSQL database connection
func initializePostgreSQL(config *common.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	// test the connection to the postgres server
	if err := testConnection(config); err != nil {
		return nil, fmt.Errorf("failed to test PostgreSQL connection: %w", err)
	}

	// Check if database exists
	exists, err := checkDatabaseExists(config)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	// If database doesn't exist, create it
	if !exists {
		if err := createDatabase(config); err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	// Build connection string for the connection with the database name
	dsn := buildPostgresConnectionString(config, config.PostgreSQL.Database)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(config.Pool.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Pool.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.Pool.ConnMaxLifetime)

	return db, nil
}

// buildPostgresConnectionString creates a PostgreSQL connection string
func buildPostgresConnectionString(config *common.Config, dbName string) string {
	if config.PostgreSQL.Host == "" {
		config.PostgreSQL.Host = "localhost"
	}
	if config.PostgreSQL.Port == 0 {
		config.PostgreSQL.Port = 5432
	}

	sslMode := "disable"
	if config.PostgreSQL.SSLMode {
		sslMode = "prefer"
	}
	conn := fmt.Sprintf("host=%s user=%s password=%s port=%d",
		config.PostgreSQL.Host,
		config.PostgreSQL.Username,
		config.PostgreSQL.Password,
		config.PostgreSQL.Port,
	)
	if sslMode != "disable" {
		conn += fmt.Sprintf(" sslmode=%s", sslMode)
	}
	if dbName != "" {
		conn += fmt.Sprintf(" database=%s", dbName)
	}
	return conn
}

// checkDatabaseExists checks if the specified database exists
func checkDatabaseExists(config *common.Config) (bool, error) {
	var dialector gorm.Dialector
	switch config.Type {
	case common.SQLite:
		return true, nil
	case common.PostgreSQL:
		dsn := buildPostgresConnectionString(config, "postgres")
		dialector = postgres.Open(dsn)
	default:
		return false, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	srv, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return false, fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}

	var exists int
	switch config.Type {
	case common.PostgreSQL:
		srv.Raw("SELECT 1 FROM pg_database WHERE datname = ?", config.PostgreSQL.Database).Scan(&exists)
	case common.SQLite:
		exists = 1
	default:
		return false, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	sqlDB, err := srv.DB()
	if err != nil {
		return false, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	return exists == 1, nil
}

// createDatabase creates the specified database
func createDatabase(config *common.Config) error {
	var dial gorm.Dialector
	switch config.Type {
	case common.SQLite:
		return nil
	case common.PostgreSQL:
		dsn := buildPostgresConnectionString(config, "postgres")
		dial = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	srv, err := gorm.Open(dial, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}

	var stmt string
	switch config.Type {
	case common.PostgreSQL:
		stmt = fmt.Sprintf("CREATE DATABASE \"%s\"", config.PostgreSQL.Database)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err := srv.Exec(stmt).Error; err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}

	var privs string
	switch config.Type {
	case common.PostgreSQL:
		privs = "GRANT ALL PRIVILEGES ON DATABASE " + config.PostgreSQL.Database + " TO " + config.PostgreSQL.Username
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err := srv.Exec(privs).Error; err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	return nil
}

// testConnection tests the connection to the postgres server
func testConnection(config *common.Config) error {
	var dial gorm.Dialector
	switch config.Type {
	case common.SQLite:
		return nil
	case common.PostgreSQL:
		dsn := buildPostgresConnectionString(config, "postgres")
		dial = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	srv, err := gorm.Open(dial, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}

	sqlDB, err := srv.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	return nil
}
