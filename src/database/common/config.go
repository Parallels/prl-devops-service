package common

import (
	"time"

	"github.com/Parallels/prl-devops-service/errors"
)

type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgresql"
)

// PoolConfig represents database connection pool configuration
type PoolConfig struct {
	MaxIdleConns    int           `json:"max_idle_conns" config:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns" config:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" config:"conn_max_lifetime"`
}

func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}
}

// SQLiteConfig represents SQLite configuration
type SQLiteConfig struct {
	StoragePath string `config:"storage_path"` // Path to the SQLite database file
	FileName    string `config:"file_name"`    // Name of the SQLite database file
}

func (s *SQLiteConfig) Validate() *errors.Diagnostics {
	diag := errors.NewDiagnostics("database::config::sqlite::validation")
	if s.StoragePath == "" {
		s.StoragePath = "."
		// Note: Using Info log instead of diagnostics warning for now
	}
	if s.FileName == "" {
		s.FileName = "database.db"
	}
	return diag
}

func DefaultSQLiteConfig() SQLiteConfig {
	return SQLiteConfig{
		StoragePath: ".",
		FileName:    "database.db",
	}
}

// PostgreSQLConfig represents PostgreSQL configuration
type PostgreSQLConfig struct {
	Host     string `config:"host"`     // PostgreSQL host
	Port     int    `config:"port"`     // PostgreSQL port
	Database string `config:"name"`     // PostgreSQL database name
	Username string `config:"username"` // PostgreSQL username
	Password string `config:"password"` // PostgreSQL password
	SSLMode  bool   `config:"ssl_mode"` // PostgreSQL SSL mode (disable, require, verify-ca, verify-full)
}

func (s *PostgreSQLConfig) Validate() *errors.Diagnostics {
	diag := errors.NewDiagnostics("database::config::postgresql::validation")
	if s.Host == "" {
		s.Host = "localhost"
	}
	if s.Port == 0 {
		s.Port = 5432
	}
	if s.Database == "" {
		s.Database = "database"
	}
	if s.Username == "" {
		s.Username = "postgres"
	}
	if s.Password == "" {
		s.Password = "password"
	}
	return diag
}

func DefaultPostgreSQLConfig() PostgreSQLConfig {
	return PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "database",
		Username: "postgres",
		Password: "password",
		SSLMode:  false,
	}
}

// Config represents database configuration
type Config struct {
	Type       DatabaseType     `json:"type" config:"type"`
	SQLite     SQLiteConfig     `json:"sqlite" config:"sqlite"`
	PostgreSQL PostgreSQLConfig `json:"postgres" config:"postgres"`

	// Common configuration
	Debug          bool       `json:"debug" config:"debug"`
	Migrate        bool       `json:"migrate" config:"migrate"`
	MigrationsPath string     `json:"migrations_path" config:"migrations_path"`
	Pool           PoolConfig `json:"pool" config:"pool"`
}

// DefaultConfig returns the default database configuration
func DefaultConfig() Config {
	return Config{
		Migrate: false,
		Pool:    DefaultPoolConfig(),
	}
}

func (d *Config) Validate() *errors.Diagnostics {
	diag := errors.NewDiagnostics("Database.Config.Validate")
	if d.Type == "" {
		diag.AddError("", "database type is required", "")
	}

	if d.Type != SQLite && d.Type != PostgreSQL {
		diag.AddError("", "invalid database type", "")
	}

	if d.Type == SQLite {
		sqliteDiag := d.SQLite.Validate()
		if sqliteDiag.HasErrors() {
			diag.Append(sqliteDiag)
		}
	}

	if d.Type == PostgreSQL {
		pgDiag := d.PostgreSQL.Validate()
		if pgDiag.HasErrors() {
			diag.Append(pgDiag)
		}
	}

	return diag
}
