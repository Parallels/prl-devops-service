package settings

// Curated definitions catalog — the whitelist of keys that can ever be set
// through the settings tables. Only runtime-safe keys resolved through
// Config.GetKey belong here; boot-critical settings (DB_*, TLS_*, API_PORT,
// MODE, ENABLED_MODULES, BASE_URL, log file paths) and every secret (keys,
// passwords, HMAC/JWT material) are permanently excluded by construction:
// they are simply never catalogued, so no write path can store them.
//
// Keys that only take effect on restart ARE catalogued but flagged
// RequiresRestart=true so the UI can offer an explicit restart instead of
// silently ignoring the change.
//
// Seeded idempotently (insert-if-absent) at boot on every SQL-mode node;
// releases grow the list here. Defaults mirror the compiled constants in the
// constants package so the catalog and the getters cannot drift silently.

import (
	"strconv"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
)

func itoa(v int) string {
	return strconv.Itoa(v)
}

// Catalog returns the compiled-in definitions. Keys are upper-case env-var
// names — the canonical identity used across the settings domain.
func Catalog() []models.SettingDefinition {
	return []models.SettingDefinition{
		{
			Key:             constants.LOG_LEVEL_ENV_VAR,
			Type:            models.SettingValueTypeString,
			DefaultValue:    "info",
			ValidValues:     "debug,info,warn,error",
			Description:     "Service log verbosity level.",
			AppliesTo:       models.SettingAppliesToAll,
			RequiresRestart: true,
		},
		{
			Key:          constants.ENABLE_CANARY_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Enable canary features.",
			AppliesTo:    models.SettingAppliesToAll,
		},
		{
			Key:          constants.ENABLE_BETA_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Enable beta features.",
			AppliesTo:    models.SettingAppliesToAll,
		},
		{
			Key:          constants.MAX_VMS_PER_HOST_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_MAX_VMS_PER_HOST),
			Description:  "Maximum number of virtual machines per host (0 = no cap).",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.MAX_VMS_PER_HOST_MODE_ENV_VAR,
			Type:         models.SettingValueTypeString,
			DefaultValue: constants.DEFAULT_MAX_VMS_PER_HOST_MODE,
			ValidValues:  "any,macos,linux",
			Description:  "Which VM flavors count against MAX_VMS_PER_HOST.",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.SYSTEM_RESERVED_CPU_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_SYSTEM_RESERVED_CPU),
			Description:  "CPU cores reserved for the host system.",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.SYSTEM_RESERVED_MEMORY_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_SYSTEM_RESERVED_MEMORY),
			Description:  "Memory in MB reserved for the host system.",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.SYSTEM_RESERVED_DISK_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_SYSTEM_RESERVED_DISK),
			Description:  "Disk space in MB reserved for the host system.",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.ORCHESTRATOR_PULL_FREQUENCY_SECONDS_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_ORCHESTRATOR_PULL_FREQUENCY_SEC),
			Description:  "How often the orchestrator polls host health (seconds).",
			AppliesTo:    models.SettingAppliesToOrchestrator,
		},
		{
			Key:          constants.ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS),
			Description:  "How often the orchestrator live-refreshes VM data from hosts (seconds).",
			AppliesTo:    models.SettingAppliesToOrchestrator,
		},
		{
			Key:          constants.DISABLE_VM_CACHE_REFRESH_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Disable the background VM cache refresh loop.",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.VM_CACHE_REFRESH_INTERVAL_SECONDS_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_VM_CACHE_REFRESH_INTERVAL_SECONDS),
			Description:  "Forced VM cache refresh interval (seconds).",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.DISABLE_CATALOG_CACHING_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Disable catalog download caching.",
			AppliesTo:    models.SettingAppliesToAll,
		},
		{
			Key:          constants.CATALOG_CACHE_KEEP_FREE_DISK_SPACE_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_CATALOG_CACHE_KEEP_FREE_DISK_SPACE),
			Description:  "Free disk space to keep free from the cache folder (MB).",
			AppliesTo:    models.SettingAppliesToAll,
		},
		{
			Key:          constants.CATALOG_CACHE_MAX_SIZE_ENV_VAR,
			Type:         models.SettingValueTypeInt,
			DefaultValue: itoa(constants.DEFAULT_CATALOG_CACHE_MAX_SIZE),
			Description:  "Maximum catalog cache size (MB); empty uses free-space derived sizing.",
			AppliesTo:    models.SettingAppliesToAll,
		},
		{
			Key:          constants.CATALOG_CACHE_ALLOW_CACHE_ABOVE_FREE_DISK_SPACE_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Allow the cache to grow beyond the keep-free-disk-space budget.",
			AppliesTo:    models.SettingAppliesToAll,
		},
		{
			Key:          constants.CATALOG_COMPRESS_VM_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Compress VM images on catalog push.",
			AppliesTo:    models.SettingAppliesToOrchestrator,
		},
		{
			Key:          constants.CATALOG_COMPRESS_VM_RATIO_ENV_VAR,
			Type:         models.SettingValueTypeString,
			DefaultValue: "best_compression",
			ValidValues:  "best_speed,balanced,best_compression,no_compression,default",
			Description:  "Compression level label used when catalog push compression is enabled.",
			AppliesTo:    models.SettingAppliesToOrchestrator,
		},
		{
			Key:          constants.ENABLE_PACKER_PLUGIN_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Enable the Packer templates feature.",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.ENABLE_VAGRANT_PLUGIN_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Enable the Vagrant feature.",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Let a host count orchestrator-owned resources as its own.",
			AppliesTo:    models.SettingAppliesToHost,
		},
		{
			Key:          constants.DISABLE_CATALOG_PROVIDER_STREAMING_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "false",
			Description:  "Disable catalog provider streaming (force plain downloads).",
			AppliesTo:    models.SettingAppliesToAll,
		},
		{
			Key:          constants.CATALOG_ENABLE_PROVIDER_CREDENTIALS_OBFUSCATION_ENV_VAR,
			Type:         models.SettingValueTypeBool,
			DefaultValue: "true",
			Description:  "Obfuscate catalog provider credentials at rest.",
			AppliesTo:    models.SettingAppliesToAll,
		},
	}
}
