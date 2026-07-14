package common

import (
	"fmt"
	"net/url"
	"strings"
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
	DSN      string `config:"dsn"`      // Full PostgreSQL connection string (takes precedence if provided)
	Host     string `config:"host"`     // PostgreSQL host
	Port     int    `config:"port"`     // PostgreSQL port
	Database string `config:"name"`     // PostgreSQL database name
	Username string `config:"username"` // PostgreSQL username
	Password string `config:"password"` // PostgreSQL password (NEVER logged in plain text)
	SSLMode  bool   `config:"ssl_mode"` // PostgreSQL SSL mode (disable, require, verify-ca, verify-full)
}

func (s *PostgreSQLConfig) Validate() *errors.Diagnostics {
	diag := errors.NewDiagnostics("database::config::postgresql::validation")

	// If DSN is provided, it takes precedence - no need to validate individual fields
	if s.DSN != "" {
		return diag
	}

	// Validate individual fields with defaults
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

// String returns a safe string representation with password redacted
func (s *PostgreSQLConfig) String() string {
	if s.DSN != "" {
		return "PostgreSQL(DSN=<redacted>)"
	}
	return fmt.Sprintf("PostgreSQL(host=%s, port=%d, database=%s, username=%s, password=<redacted>, ssl=%t)",
		s.Host, s.Port, s.Database, s.Username, s.SSLMode)
}

// SafeDSN returns the DSN with credentials redacted for logging
func (s *PostgreSQLConfig) SafeDSN() string {
	if s.DSN != "" {
		// Redact password from DSN: postgresql://user:password@host -> postgresql://user:***@host
		return redactDSNPassword(s.DSN)
	}
	// Build DSN from individual fields with password redacted
	sslMode := "disable"
	if s.SSLMode {
		sslMode = "require"
	}
	return fmt.Sprintf("postgresql://%s:***@%s:%d/%s?sslmode=%s",
		s.Username, s.Host, s.Port, s.Database, sslMode)
}

func DefaultPostgreSQLConfig() PostgreSQLConfig {
	return PostgreSQLConfig{
		DSN:      "",
		Host:     "localhost",
		Port:     5432,
		Database: "database",
		Username: "postgres",
		Password: "password",
		SSLMode:  false,
	}
}

// redactDSNPassword redacts the password from a PostgreSQL DSN for safe logging
func redactDSNPassword(dsn string) string {
	// Parse the DSN
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return "postgresql://<invalid-dsn>"
	}

	// Redact password if present
	if u.User != nil {
		username := u.User.Username()
		// Create new URL with redacted password
		if _, hasPassword := u.User.Password(); hasPassword {
			u.User = url.User(username + ":***")
		}
	}

	// url.String() will URL-encode the password placeholder, so we need to unescape it
	result := u.String()
	// Replace URL-encoded *** with plain ***
	result = strings.ReplaceAll(result, "%3A%2A%2A%2A", ":***")
	result = strings.ReplaceAll(result, ":%2A%2A%2A", ":***")
	return result
}

// ParseDSN parses a PostgreSQL DSN and returns individual connection parameters
// Format: postgresql://username:password@host:port/database?sslmode=disable
func ParseDSN(dsn string) (host string, port int, database, username, password string, sslMode bool, err error) {
	u, parseErr := url.Parse(dsn)
	if parseErr != nil {
		return "", 0, "", "", "", false, fmt.Errorf("invalid DSN format: %w", parseErr)
	}

	// Check if it's a valid PostgreSQL DSN
	if u.Scheme != "postgresql" && u.Scheme != "postgres" {
		return "", 0, "", "", "", false, fmt.Errorf("invalid DSN scheme: expected postgresql or postgres, got %s", u.Scheme)
	}

	// Extract username and password
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	// Extract host and port
	host = u.Hostname()
	portStr := u.Port()
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	} else {
		port = 5432 // Default PostgreSQL port
	}

	// Extract database name (path without leading /)
	database = strings.TrimPrefix(u.Path, "/")

	// Extract SSL mode from query parameters
	sslModeStr := u.Query().Get("sslmode")
	sslMode = sslModeStr == "require" || sslModeStr == "verify-ca" || sslModeStr == "verify-full"

	return host, port, database, username, password, sslMode, nil
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
