package config

// Cohesive getter group moved from main.go (pure move, no behaviour
// change) - see .ai-agents/features/417-gorm-database-layer/420-database-configuration/plan.md D-6

import (
	"time"

	"github.com/Parallels/prl-devops-service/constants"
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
