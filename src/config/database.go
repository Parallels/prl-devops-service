package config

// Cohesive getter group moved from main.go (pure move, no behaviour
// change) - see .ai-agents/features/417-gorm-database-layer/420-database-configuration/plan.md D-6

import (
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
)

func (c *Config) DbNumberBackupFiles() int {
	numberFiles := c.GetIntKey(constants.DATABASE_NUMBER_BACKUP_FILES_ENV_VAR)
	if numberFiles == 0 {
		return 10
	}

	return numberFiles
}

func (c *Config) DbBackupInterval() time.Duration {
	interval := c.GetIntKey(constants.DATABASE_BACKUP_INTERVAL_ENV_VAR)
	if interval == 0 {
		return 2 * time.Hour
	}

	return time.Duration(interval) * time.Minute
}

func (c *Config) DbSaveInterval() time.Duration {
	interval := c.GetIntKey(constants.DATABASE_SAVE_INTERVAL_ENV_VAR)
	if interval == 0 {
		return 2 * time.Minute
	}

	return time.Duration(interval) * time.Minute
}

// DbGhostJobTimeoutMinutes returns the configured timeout in minutes for
// detecting ghost (stuck) jobs. Falls back to DefaultGhostJobTimeoutMinutes
// if GHOST_JOB_TIMEOUT_MINUTES is missing, nonpositive, invalid, or would
// overflow a time.Duration.
func (c *Config) DbGhostJobTimeoutMinutes() int {
	timeout := c.GetIntKey(constants.GhostJobTimeoutMinutesEnvVar)
	if timeout <= 0 || int64(timeout) > int64((1<<63-1)/time.Minute) {
		return constants.DefaultGhostJobTimeoutMinutes
	}
	return timeout
}

func (c *Config) IsDatabaseAutoRecover() bool {
	envVar := c.GetKey(constants.SYSTEM_AUTO_RECOVER_DATABASE_ENV_VAR)
	if envVar == "" ||
		envVar == "true" ||
		envVar == "1" ||
		envVar == "yes" ||
		envVar == "y" ||
		envVar == "t" ||
		envVar == "on" {
		return true
	}

	return false
}

func (c *Config) DatabaseFolder() string {
	return c.GetKey(constants.DATABASE_FOLDER_ENV_VAR)
}

// DatabaseType returns the raw DB_TYPE setting ("" = unset → caller defaults to sqlite).
// Config file key db_type resolves through the same case-insensitive GetKey lookup.
func (c *Config) DatabaseType() string { return c.GetKey(constants.DB_TYPE_ENV_VAR) }

// DatabasePath returns the raw DB_PATH setting ("" = unset → AD-3 default resolution).
func (c *Config) DatabasePath() string { return c.GetKey(constants.DB_PATH_ENV_VAR) }

// DatabaseDSN returns the raw DB_DSN setting ("" = unset). Treat as a secret:
// callers must never embed it in error/log text unredacted (#420 plan D-8).
func (c *Config) DatabaseDSN() string { return c.GetKey(constants.DB_DSN_ENV_VAR) }

// DatabaseAutoMigrate parses DB_AUTO_MIGRATE. Unset → true (AC: "Defaults to true").
// Vocabulary extends the repo's tolerant-bool sets (helpers.StringToBool,
// src/helpers/strings.go:144; IsDatabaseAutoRecover) with an explicit false-set —
// and, deliberately unlike both, returns a clear error for garbage instead of
// silently defaulting (AC "Validation and errors").
func (c *Config) DatabaseAutoMigrate() (bool, error) {
	raw := strings.ToLower(strings.TrimSpace(c.GetKey(constants.DB_AUTO_MIGRATE_ENV_VAR)))
	switch raw {
	case "":
		return true, nil
	case "true", "1", "t", "yes", "y", "on", "enabled":
		return true, nil
	case "false", "0", "f", "no", "n", "off", "disabled":
		return false, nil
	default:
		// Echo capped at 64 bytes (same guard idea as database.quoteCapped, D-7.1): a
		// mis-set value is echoed to its owner, but a paste-gone-wrong can't spam logs.
		echo := raw
		if len(echo) > 64 {
			echo = echo[:64] + "…"
		}
		return false, errors.NewWithCode(
			fmt.Sprintf("invalid DB_AUTO_MIGRATE %q (config key db_auto_migrate): expected true or false", echo), 400)
	}
}
