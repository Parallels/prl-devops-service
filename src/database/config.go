package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
)

const (
	DBTypeSQLite   = "sqlite"
	DBTypePostgres = "postgres"
	// DBTypeLegacy is the #421 S3 rollback lever consumed BEFORE config parsing
	// (serviceprovider.selectDatabaseBackend): FromApplicationConfig and
	// Validate intentionally reject it — no DatabaseConfig is ever built
	// against the legacy engine.
	DBTypeLegacy = "legacy"
	// DefaultDatabaseFilename preserves the legacy *folder* convention (AD-3)
	// with a distinct filename so data.json and prl-devops.db coexist.
	DefaultDatabaseFilename = "prl-devops.db"
	MemoryPath              = ":memory:"
)

const (
	DefaultPostgresMaxOpenConns    = 25
	DefaultPostgresMaxIdleConns    = 5
	DefaultPostgresConnMaxLifetime = 30 * time.Minute
)

// PostgresPoolConfig carries pool settings with zero-means-default semantics;
// the defaults are applied by the getters, at the use site (task 04).
type PostgresPoolConfig struct {
	MaxOpenConns    int           // 0 → DefaultPostgresMaxOpenConns
	MaxIdleConns    int           // 0 → DefaultPostgresMaxIdleConns
	ConnMaxLifetime time.Duration // 0 → DefaultPostgresConnMaxLifetime
	ConnMaxIdleTime time.Duration // 0 → leave driver default
}

// DatabaseConfig is the neutral contract consumed by the connection factory
// (tasks 03/04). It is deliberately not config.Config (D-3): the factory must
// not read env vars itself, so config semantics cannot splinter before #420.
type DatabaseConfig struct {
	Type        string
	Path        string
	DSN         string
	AutoMigrate bool // stored for #421/#424; never read in #418
	Postgres    PostgresPoolConfig
}

// DefaultDatabaseConfig is the AD-3 default: SQLite, file-based, AutoMigrate on.
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Type:        DBTypeSQLite,
		Path:        defaultDatabasePath(),
		AutoMigrate: true,
	}
}

// LegacyDataFolder is the pre-unification root data folder. No longer a
// default; kept only as a legacy-migration store discovery candidate, so
// #424 imports data.json on boxes built before the folders were unified.
const LegacyDataFolder = "/etc/parallels-devops"

// DefaultDataFolder is the single source of truth for where an installation
// keeps its data — the legacy data.json AND the SQL database. It replaces
// the AD-3 duplication (two provider inits in serviceprovider/main.go,
// this file, startup/legacy_migration.go), which had drifted into three
// different answers for the root case. Resolution order: explicit
// DATABASE_FOLDER (a pin always wins, never falls back) → root:
// constants.ServiceDefaultDirectory (/etc/prl-devops-service — the service
// folder the installer, config resolution, catalog cache and host-agent
// installs already use) → non-root: ~/.parallels-devops → home lookup
// failure: ServiceDefaultDirectory. Nothing here creates directories;
// callers keep that posture (panic on failure / caller-visible fallback).
func DefaultDataFolder() string {
	if folder := config.Get().DatabaseFolder(); folder != "" {
		return folder
	}
	if os.Getuid() == 0 {
		return constants.ServiceDefaultDirectory
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".parallels-devops")
	}
	return constants.ServiceDefaultDirectory
}

// defaultDatabasePath resolves the SQLite file inside DefaultDataFolder
// (AD-3 resolved: the folder logic lived in four places; it now lives here).
func defaultDatabasePath() string {
	return filepath.Join(DefaultDataFolder(), DefaultDatabaseFilename)
}

// Validate checks the config shape without touching the filesystem or network.
// An empty Type is treated as sqlite (defaults-forward-compatible); type
// comparison is case-insensitive, callers store normalised lowercase values.
func (c DatabaseConfig) Validate() error {
	dbType := strings.TrimSpace(c.Type)
	if dbType == "" {
		dbType = DBTypeSQLite
	}

	switch {
	case strings.EqualFold(dbType, DBTypePostgres):
		if strings.TrimSpace(c.DSN) == "" {
			return ErrPostgresDSNRequired
		}
	case strings.EqualFold(dbType, DBTypeSQLite):
		if strings.TrimSpace(c.Path) == "" {
			return ErrSQLitePathRequired
		}
	default:
		return ErrUnsupportedDatabaseType
	}

	return nil
}

func (p PostgresPoolConfig) MaxOpen() int {
	if p.MaxOpenConns == 0 {
		return DefaultPostgresMaxOpenConns
	}
	return p.MaxOpenConns
}

func (p PostgresPoolConfig) MaxIdle() int {
	if p.MaxIdleConns == 0 {
		return DefaultPostgresMaxIdleConns
	}
	return p.MaxIdleConns
}

func (p PostgresPoolConfig) Lifetime() time.Duration {
	if p.ConnMaxLifetime == 0 {
		return DefaultPostgresConnMaxLifetime
	}
	return p.ConnMaxLifetime
}

// IdleTime returns the raw value: zero intentionally means "driver default",
// no service-side default is imposed (0 → pass 0 through to database/sql).
func (p PostgresPoolConfig) IdleTime() time.Duration {
	return p.ConnMaxIdleTime
}

// FromApplicationConfig builds DatabaseConfig from DB_TYPE/DB_PATH/DB_DSN/DB_AUTO_MIGRATE
// resolved through config.Get() — env var > CLI flag > config-file key (Config.GetKey),
// config-file keys flat lowercase (db_type, db_path, db_dsn, db_auto_migrate; #420 D-1).
// Unset values keep DefaultDatabaseConfig()'s AD-3 defaults byte-identically (D-5).
// Errors name the offending setting + allowed values; the DSN never appears in messages (D-8).
// #420 D-4: the (value, error) signature is adopted atomically with this body —
// no caller exists yet (task 04 wires the provider), so widening is free now.
func FromApplicationConfig() (DatabaseConfig, error) {
	c := config.Get()
	cfg := DefaultDatabaseConfig()

	if t := strings.ToLower(strings.TrimSpace(c.DatabaseType())); t != "" {
		cfg.Type = t
	}
	if p := strings.TrimSpace(c.DatabasePath()); p != "" {
		cfg.Path = p // DB_PATH beats DATABASE_FOLDER-derived default (OQ-420-4)
	}
	cfg.DSN = strings.TrimSpace(c.DatabaseDSN())

	autoMigrate, err := c.DatabaseAutoMigrate()
	if err != nil {
		return DatabaseConfig{}, err
	}
	cfg.AutoMigrate = autoMigrate

	switch strings.ToLower(cfg.Type) {
	case DBTypeSQLite, DBTypePostgres: // constants above, same package
	default:
		return DatabaseConfig{}, fmt.Errorf("invalid DB_TYPE %s (config key db_type): %w",
			quoteCapped(cfg.Type, 64), ErrUnsupportedDatabaseType) // sentinel text names sqlite|postgres
	}
	if strings.EqualFold(cfg.Type, DBTypePostgres) && strings.TrimSpace(cfg.DSN) == "" {
		return DatabaseConfig{}, fmt.Errorf("%w — set DB_DSN (env) or db_dsn (config key db_dsn)",
			ErrPostgresDSNRequired)
	}
	if err := cfg.Validate(); err != nil { // #418 checks stay the single source of truth
		return DatabaseConfig{}, err
	}
	return cfg, nil
}

// quoteCapped quotes v for an error message, truncated at n bytes so a huge typo
// can't spam logs. Echoing a user's own setting value back to them is safe; the
// DSN is never routed through this (D-8).
func quoteCapped(v string, n int) string {
	if len(v) > n {
		v = v[:n] + "…"
	}
	return strconv.Quote(v)
}
