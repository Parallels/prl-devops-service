package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/serviceprovider/system"

	log "github.com/cjlapao/common-go-logger"
	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/security"
)

type Config struct {
	mode                string
	includeOwnResources bool
}

func NewConfig() *Config {
	return &Config{
		mode: "api",
	}
}

func (c *Config) GetApiPort() string {
	port := helper.GetFlagValue(constants.API_PORT_FLAG, "")

	if port == "" {
		port = os.Getenv(constants.API_PORT_ENV_VAR)
	}

	if port == "" {
		port = constants.DEFAULT_API_PORT
	}

	return port
}

func (c *Config) GetApiPrefix() string {
	apiPrefix := os.Getenv(constants.API_PREFIX_ENV_VAR)
	if apiPrefix == "" {
		apiPrefix = constants.DEFAULT_API_PREFIX
	}

	return apiPrefix
}

func (c *Config) GetHmacSecret() string {
	hmacSecret := os.Getenv(constants.HMAC_SECRET_ENV_VAR)
	return hmacSecret
}

func (c *Config) GetLogLevel() string {
	logLevel := os.Getenv(constants.LOG_LEVEL_ENV_VAR)
	if logLevel != "" {
		common.Logger.Info("Log Level set to %v", logLevel)
	}
	switch strings.ToLower(logLevel) {
	case "debug":
		logLevel = "DEBUG"
		common.Logger.LogLevel = log.Debug
	case "info":
		logLevel = "INFO"
		common.Logger.LogLevel = log.Info
	case "warn":
		logLevel = "WARN"
		common.Logger.LogLevel = log.Warning
	case "error":
		logLevel = "ERROR"
		common.Logger.LogLevel = log.Error
	}

	return logLevel
}

func (c *Config) GetSecurityKey() string {
	securityKey := os.Getenv(constants.SECURITY_KEY_ENV_VAR)
	if securityKey == "" {
		return ""
	}

	decoded, err := security.DecodeBase64String(securityKey)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}
	securityKey = decoded
	return securityKey
}

func (c *Config) GetTlsCertificate() string {
	tlsCertificate := os.Getenv(constants.TLS_CERTIFICATE_ENV_VAR)
	decoded, err := security.DecodeBase64String(tlsCertificate)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}
	tlsCertificate = string(decoded)
	return tlsCertificate
}

func (c *Config) GetTlsPrivateKey() string {
	tlsPrivateKey := os.Getenv(constants.TLS_PRIVATE_KEY_ENV_VAR)
	decoded, err := security.DecodeBase64String(tlsPrivateKey)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}

	tlsPrivateKey = string(decoded)
	return tlsPrivateKey
}

func (c *Config) GetTLSPort() string {
	tlsPort := os.Getenv(constants.TLS_PORT_ENV_VAR)
	if tlsPort == "" {
		tlsPort = constants.DEFAULT_API_TLS_PORT
	}

	return tlsPort
}

func (c *Config) TLSEnabled() bool {
	TLSEnabled := os.Getenv(constants.TLS_ENABLED_ENV_VAR)
	if TLSEnabled == "" || TLSEnabled == "false" {
		return false
	}
	if c.GetTlsCertificate() == "" || c.GetTlsPrivateKey() == "" {
		return false
	}
	return true
}

func (c *Config) GetTokenDurationMinutes() int {
	tokenDuration := os.Getenv(constants.TOKEN_DURATION_MINUTES_ENV_VAR)
	if tokenDuration != "" {
		return constants.DEFAULT_TOKEN_DURATION_MINUTES
	}

	intVal, err := strconv.Atoi(tokenDuration)
	if err != nil {
		return constants.DEFAULT_TOKEN_DURATION_MINUTES
	}
	return intVal
}

func (c *Config) GetRootFolder() (string, error) {
	srv := system.Get()
	ctx := basecontext.NewRootBaseContext()
	currentUser, err := srv.GetCurrentUser(ctx)
	if err != nil {
		currentUser = "root"
	}

	if currentUser == "root" {
		folder := "/etc/parallels-api-service"
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

func (c *Config) GetCatalogCacheFolder() (string, error) {
	rootFolder, err := c.GetRootFolder()
	if err != nil {
		return "", err
	}

	cacheFolder := filepath.Join(rootFolder, constants.DEFAULT_CATALOG_CACHE_FOLDER)
	if os.Getenv(constants.CATALOG_CACHE_FOLDER_ENV_VAR) != "" {
		cacheFolder = os.Getenv(constants.CATALOG_CACHE_FOLDER_ENV_VAR)
	}

	err = helpers.CreateDirIfNotExist(cacheFolder)
	if err != nil {
		return "", err
	}

	return cacheFolder, nil
}

func (c *Config) IsCatalogCachingEnable() bool {
	envVar := os.Getenv(constants.DISABLE_CATALOG_CACHING_ENV_VAR)
	if envVar == "true" || envVar == "1" {
		return false
	}

	return true
}

func (c *Config) GetSystemMode() string {
	c.mode = os.Getenv(constants.MODE_ENV_VAR)
	if c.mode != "" {
		return c.mode
	}

	c.mode = helper.GetFlagValue(constants.MODE_FLAG, "unknown")
	if c.mode == "" {
		c.mode = "api"
	}

	return c.mode
}

func (c *Config) GetOrchestratorPullFrequency() int {
	frequency := os.Getenv(constants.ORCHESTRATOR_PULL_FREQUENCY_SECONDS_ENV_VAR)
	if frequency == "" {
		return constants.DEFAULT_ORCHESTRATOR_PULL_FREQUENCY_SEC
	}

	intVal, err := strconv.Atoi(frequency)
	if err != nil {
		return constants.DEFAULT_ORCHESTRATOR_PULL_FREQUENCY_SEC
	}

	return intVal
}

func (c *Config) GetDatabaseFolder() string {
	return os.Getenv(constants.DATABASE_FOLDER_ENV_VAR)
}

func (c *Config) GetLocalhost() string {
	schema := "http"
	host := "localhost"
	port := c.GetApiPort()
	if c.TLSEnabled() {
		schema = "https"
		port = c.GetTLSPort()
	}

	return schema + "://" + host + ":" + port
}

func (c *Config) IsOrchestrator() bool {
	return c.GetSystemMode() == constants.ORCHESTRATOR_MODE
}

func (c *Config) IsCatalog() bool {
	return c.GetSystemMode() == constants.CATALOG_MODE
}

func (c *Config) IsApi() bool {
	return c.GetSystemMode() == constants.API_MODE
}

func (c *Config) UseOrchestratorResources() bool {
	ownResources := os.Getenv(constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR)
	if ownResources != "" {
		if ownResources == "true" || ownResources == "1" {
			c.includeOwnResources = true
			return true
		}
	}

	ownResources = helper.GetFlagValue(constants.USE_ORCHESTRATOR_RESOURCES_FLAG, "unknown")
	if ownResources != "" {
		if ownResources == "true" || ownResources == "1" {
			c.includeOwnResources = true
			return true
		}
	}

	return false
}
