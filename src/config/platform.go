package config

// Cohesive getter group moved from main.go (pure move, no behaviour
// change) - see .ai-agents/features/417-gorm-database-layer/420-database-configuration/plan.md D-6

import (
	"path/filepath"

	"strconv"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	appversion "github.com/Parallels/prl-devops-service/version"
)

func (c *Config) RootFolder() (string, error) {
	ctx := basecontext.NewRootBaseContext()
	srv := system.Get()
	currentUser, err := srv.GetCurrentUser(ctx)
	if err != nil {
		currentUser = "root"
	}

	if currentUser == "root" {
		folder := constants.ServiceDefaultDirectory
		err := helpers.CreateDirIfNotExist(folder)
		if err != nil {
			return "", err
		}

		return folder, nil
	} else {
		userHome, err := srv.GetUserHome(ctx, currentUser)
		if err != nil {
			return "", err
		}
		folder := userHome + "/.parallels-api-service"
		err = helpers.CreateDirIfNotExist(folder)
		if err != nil {
			return "", err
		}

		return folder, nil
	}
}

func (c *Config) CatalogCacheFolder() (string, error) {
	rootFolder, err := c.RootFolder()
	if err != nil {
		return "", err
	}

	cacheFolder := filepath.Join(rootFolder, constants.DEFAULT_CATALOG_CACHE_FOLDER)
	if c.GetKey(constants.CATALOG_CACHE_FOLDER_ENV_VAR) != "" {
		cacheFolder = c.GetKey(constants.CATALOG_CACHE_FOLDER_ENV_VAR)
	}

	err = helpers.CreateDirIfNotExist(cacheFolder)
	if err != nil {
		return "", err
	}

	return cacheFolder, nil
}

func (c *Config) IsCatalogCachingEnable() bool {
	envVar := c.GetKey(constants.DISABLE_CATALOG_CACHING_ENV_VAR)
	if envVar == "true" || envVar == "1" {
		return false
	}

	return true
}

func (c *Config) CacheKeepFreeDiskSpace() int64 {
	val := c.GetKey(constants.CATALOG_CACHE_KEEP_FREE_DISK_SPACE_ENV_VAR)
	if val == "" {
		return constants.DEFAULT_CATALOG_CACHE_KEEP_FREE_DISK_SPACE
	}

	return int64(c.GetIntKey(constants.CATALOG_CACHE_KEEP_FREE_DISK_SPACE_ENV_VAR))
}

func (c *Config) CacheMaxSize(freeDiskSpace int64) int64 {
	val := c.GetKey(constants.CATALOG_CACHE_MAX_SIZE_ENV_VAR)
	if val == "" {
		maxSize := freeDiskSpace - c.CacheKeepFreeDiskSpace()
		if maxSize < 0 {
			maxSize = 0
		}
		return maxSize
	}

	return int64(c.GetIntKey(constants.CATALOG_CACHE_MAX_SIZE_ENV_VAR))
}

func (c *Config) AllowCacheAboveFreeDiskSpace() bool {
	return c.GetBoolKey(constants.CATALOG_CACHE_ALLOW_CACHE_ABOVE_FREE_DISK_SPACE_ENV_VAR)
}

func (c *Config) Mode() string {
	c.mode = c.GetKey(constants.MODE_ENV_VAR)
	if c.mode != "" {
		return c.mode
	}

	return c.mode
}

func (c *Config) SystemReservedCpu() int {
	val := c.GetIntKey(constants.SYSTEM_RESERVED_CPU_ENV_VAR)
	if val > 0 {
		return val
	}

	return constants.DEFAULT_SYSTEM_RESERVED_CPU
}

func (c *Config) SystemReservedMemory() int {
	val := c.GetIntKey(constants.SYSTEM_RESERVED_MEMORY_ENV_VAR)
	if val > 0 {
		return val
	}

	return constants.DEFAULT_SYSTEM_RESERVED_MEMORY
}

func (c *Config) SystemReservedDisk() int {
	val := c.GetIntKey(constants.SYSTEM_RESERVED_DISK_ENV_VAR)
	if val > 0 {
		return val
	}

	return constants.DEFAULT_SYSTEM_RESERVED_DISK
}

// MaxVmsPerHost returns the hard maximum number of virtual machines this host may
// host. It is a count limit set on the host itself and is applied even when the host
// still has free CPU, memory and disk. An absent key, an unparsable value and any
// value <= 0 all mean "unlimited".
// (Ported verbatim from upstream #577, which added it in the pre-split main.go;
// the sync-merge lands it in the split file its SystemReservedDisk sibling moved to.)
func (c *Config) MaxVmsPerHost() int64 {
	raw := strings.TrimSpace(c.GetKey(constants.MAX_VMS_PER_HOST_ENV_VAR))
	if raw == "" {
		return constants.DEFAULT_MAX_VMS_PER_HOST
	}

	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		if c.ctx != nil {
			c.ctx.LogWarnf("Invalid %s value %q; using unlimited", constants.MAX_VMS_PER_HOST_ENV_VAR, raw)
		}
		return constants.DEFAULT_MAX_VMS_PER_HOST
	}
	if val <= 0 {
		return constants.DEFAULT_MAX_VMS_PER_HOST
	}

	return val
}

// MaxVmsPerHostMode returns which virtual machines the per-host limit counts:
// "any" (every VM registered on the host, including stopped ones) or "running"
// (only VMs occupying a hypervisor slot). Unknown values fall back to "any".
func (c *Config) MaxVmsPerHostMode() string {
	mode := strings.ToLower(strings.TrimSpace(c.GetKey(constants.MAX_VMS_PER_HOST_MODE_ENV_VAR)))
	switch mode {
	case constants.MAX_VMS_PER_HOST_MODE_ANY, constants.MAX_VMS_PER_HOST_MODE_RUNNING:
		return mode
	case "":
		return constants.DEFAULT_MAX_VMS_PER_HOST_MODE
	}

	if c.ctx != nil {
		c.ctx.LogWarnf("Invalid %s value %q; using %s",
			constants.MAX_VMS_PER_HOST_MODE_ENV_VAR, mode, constants.DEFAULT_MAX_VMS_PER_HOST_MODE)
	}

	return constants.DEFAULT_MAX_VMS_PER_HOST_MODE
}

func (c *Config) OrchestratorPullFrequency() int {
	frequency := c.GetKey(constants.ORCHESTRATOR_PULL_FREQUENCY_SECONDS_ENV_VAR)
	if frequency == "" {
		return constants.DEFAULT_ORCHESTRATOR_PULL_FREQUENCY_SEC
	}

	intVal, err := strconv.Atoi(frequency)
	if err != nil {
		return constants.DEFAULT_ORCHESTRATOR_PULL_FREQUENCY_SEC
	}

	return intVal
}

func (c *Config) OrchestratorPublicUrl() string {
	url := c.GetKey(constants.ORCHESTRATOR_PUBLIC_URL)
	if url == "" {
		return "localhost"
	}
	return url
}

func (c *Config) IsRemoteProviderStreamEnabled() bool {
	streamingDisabled := c.GetKey(constants.DISABLE_CATALOG_PROVIDER_STREAMING_ENV_VAR)
	if streamingDisabled == "" {
		return true
	}

	return !helpers.StringToBool(streamingDisabled)
}

func (c *Config) IsCanaryEnabled() bool {
	if appversion.Get().IsCanary() {
		return true
	}
	return c.GetBoolKey(constants.ENABLE_CANARY_ENV_VAR)
}

func (c *Config) CatalogCompressVM() bool {
	compressVM := c.GetBoolKey(constants.CATALOG_COMPRESS_VM_ENV_VAR)
	return compressVM
}

func (c *Config) CatalogCompressRatio() int {
	compressRatio := strings.ToLower(c.GetKey(constants.CATALOG_COMPRESS_VM_RATIO_ENV_VAR))
	returnValue := 9

	ratio, err := helpers.ConvertCompressRatioFromString(compressRatio)
	if err == nil {
		returnValue = ratio
	}

	return returnValue
}

func (c *Config) IsBetaEnabled() bool {
	if appversion.Get().IsBeta() {
		return true
	}
	return c.GetBoolKey(constants.ENABLE_BETA_ENV_VAR)
}

func (c *Config) IsCacheRefreshEnabled() bool {
	disableRefreshEnvValue := c.GetBoolKey(constants.DISABLE_VM_CACHE_REFRESH_ENV_VAR)
	if disableRefreshEnvValue {
		return false
	}
	return true
}

func (c *Config) ForceCacheRefreshInterval() time.Duration {
	cacheRefreshIntervalEnvValue := c.GetKey(constants.VM_CACHE_REFRESH_INTERVAL_SECONDS_ENV_VAR)
	if cacheRefreshIntervalEnvValue == "" {
		return time.Duration(constants.DEFAULT_VM_CACHE_REFRESH_INTERVAL_SECONDS) * time.Second
	}
	return time.Duration(c.GetIntKey(constants.VM_CACHE_REFRESH_INTERVAL_SECONDS_ENV_VAR)) * time.Second
}

func (c *Config) EnableCredentialsObfuscation() bool {
	enableObfuscationEnvValue := c.GetKey(constants.CATALOG_ENABLE_PROVIDER_CREDENTIALS_OBFUSCATION_ENV_VAR)
	if enableObfuscationEnvValue == "" {
		return constants.ENABLE_CREDENTIALS_OBFUSCATION_DEFAULT_VALUE
	}

	return c.GetBoolKey(constants.CATALOG_ENABLE_PROVIDER_CREDENTIALS_OBFUSCATION_ENV_VAR)
}
