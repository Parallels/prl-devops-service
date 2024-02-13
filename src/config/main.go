package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/serviceprovider/system"
	"gopkg.in/yaml.v3"

	log "github.com/cjlapao/common-go-logger"
	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/security"
)

var globalConfig *Config

var extensions = []string{
	".local.yaml",
	".local.yml",
	".local.json",
	".yaml",
	".yml",
	".json",
}

type Config struct {
	ctx                 basecontext.ApiContext
	mode                string
	includeOwnResources bool
	fileFormat          string
	filename            string
	config              ConfigFile
}

func New(ctx basecontext.ApiContext) *Config {
	globalConfig = &Config{
		mode:       "api",
		ctx:        ctx,
		fileFormat: "yaml",
		filename:   "config.yml",
		config:     ConfigFile{},
	}

	globalConfig.LogLevel(false)
	return globalConfig
}

func Get() *Config {
	if globalConfig == nil {
		ctx := basecontext.NewBaseContext()
		return New(ctx)
	}

	return globalConfig
}

func (c *Config) Load() bool {
	fileName := ""
	configFileName := helper.GetFlagValue(constants.CONFIG_FILE_FLAG, "")
	if configFileName != "" {
		if _, err := os.Stat(configFileName); !os.IsNotExist(err) {
			fileName = configFileName
		}
	} else {
		for _, extension := range extensions {
			if _, err := os.Stat(fmt.Sprintf("config%s", extension)); !os.IsNotExist(err) {
				fileName = fmt.Sprintf("config%s", extension)
				break
			}
		}
	}

	if fileName == "" {
		c.ctx.LogInfof("No configuration file found")
		return false
	}

	c.ctx.LogInfof("Loading configuration from %s", fileName)
	content, err := helper.ReadFromFile(fileName)
	if err != nil {
		c.ctx.LogErrorf("Error reading configuration file: %s", err.Error())
		return false
	}

	if content == nil {
		c.ctx.LogErrorf("Error reading configuration file: %s", err.Error())
		return false
	}

	if strings.HasSuffix(fileName, ".json") {
		err = json.Unmarshal(content, &c.config)
		if err != nil {
			c.ctx.LogErrorf("Error reading configuration file: %s", err.Error())
			return false
		}
		c.fileFormat = "json"
	} else {
		err = yaml.Unmarshal(content, &c.config)
		if err != nil {
			c.ctx.LogErrorf("Error reading configuration file: %s", err.Error())
			return false
		}
		c.fileFormat = "yaml"
	}

	c.LogLevel(true)
	c.filename = fileName
	return true
}

func (c *Config) Save() bool {
	var content []byte
	var err error

	switch c.fileFormat {
	case "json":
		content, err = json.Marshal(c.config)
		if err != nil {
			c.ctx.LogErrorf("Error saving configuration file: %s", err.Error())
			return false
		}
	case "yaml":
		content, err = yaml.Marshal(c.config)
		if err != nil {
			c.ctx.LogErrorf("Error saving configuration file: %s", err.Error())
			return false
		}
	}

	err = helper.WriteToFile(string(content), c.filename)
	if err != nil {
		c.ctx.LogErrorf("Error saving configuration file: %s", err.Error())
		return false
	}

	return true
}

func (c *Config) ApiPort() string {
	port := c.GetKey(constants.API_PORT_ENV_VAR)

	if port == "" {
		port = constants.DEFAULT_API_PORT
	}

	return port
}

func (c *Config) ApiPrefix() string {
	apiPrefix := c.GetKey(constants.API_PREFIX_ENV_VAR)
	if apiPrefix == "" {
		apiPrefix = constants.DEFAULT_API_PREFIX
	}

	return apiPrefix
}

func (c *Config) LogLevel(silent bool) string {
	logLevel := c.GetKey(constants.LOG_LEVEL_ENV_VAR)
	if logLevel != "" && !silent {
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

func (c *Config) EncryptionPrivateKey() string {
	securityKey := c.GetKey(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR)
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

func (c *Config) TlsCertificate() string {
	tlsCertificate := c.GetKey(constants.TLS_CERTIFICATE_ENV_VAR)
	decoded, err := security.DecodeBase64String(tlsCertificate)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}
	tlsCertificate = string(decoded)
	return tlsCertificate
}

func (c *Config) TlsPrivateKey() string {
	tlsPrivateKey := c.GetKey(constants.TLS_PRIVATE_KEY_ENV_VAR)
	decoded, err := security.DecodeBase64String(tlsPrivateKey)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}

	tlsPrivateKey = string(decoded)
	return tlsPrivateKey
}

func (c *Config) TlsPort() string {
	tlsPort := c.GetKey(constants.TLS_PORT_ENV_VAR)
	if tlsPort == "" {
		tlsPort = constants.DEFAULT_API_TLS_PORT
	}

	return tlsPort
}

func (c *Config) TlsEnabled() bool {
	TLSEnabled := c.GetKey(constants.TLS_ENABLED_ENV_VAR)
	if TLSEnabled == "" || TLSEnabled == "false" {
		return false
	}
	if c.TlsCertificate() == "" || c.TlsPrivateKey() == "" {
		return false
	}
	return true
}

func (c *Config) RootFolder() (string, error) {
	ctx := basecontext.NewRootBaseContext()
	srv := system.Get()
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

func (c *Config) Mode() string {
	c.mode = c.GetKey(constants.MODE_ENV_VAR)
	if c.mode != "" {
		return c.mode
	}

	return c.mode
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

func (c *Config) DatabaseFolder() string {
	return c.GetKey(constants.DATABASE_FOLDER_ENV_VAR)
}

func (c *Config) Localhost() string {
	schema := "http"
	host := "localhost"
	port := c.ApiPort()
	if c.TlsEnabled() {
		schema = "https"
		port = c.TlsPort()
	}

	return schema + "://" + host + ":" + port
}

func (c *Config) IsOrchestrator() bool {
	return c.Mode() == constants.ORCHESTRATOR_MODE
}

func (c *Config) IsCatalog() bool {
	return c.Mode() == constants.CATALOG_MODE
}

func (c *Config) IsApi() bool {
	return c.Mode() == constants.API_MODE
}

func (c *Config) UseOrchestratorResources() bool {
	ownResources := c.GetKey(constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR)
	if ownResources != "" {
		if ownResources == "true" || ownResources == "1" {
			c.includeOwnResources = true
			return true
		}
	}

	return false
}

func (c *Config) GetReverseProxyConfig() *ReverseProxyConfig {
	return c.config.ReverseProxy
}

func (c *Config) GetKey(key string) string {
	value := helper.GetFlagValue(key, "")
	exists := false

	if value == "" {
		value, exists = os.LookupEnv(key)
		if value == "" && !exists {
			for k, v := range c.config.Environment {
				if strings.EqualFold(k, key) {
					value = v
					break
				}
			}
		}
	}

	return value
}

func (c *Config) GetIntKey(key string) int {
	value := c.GetKey(key)
	if value == "" {
		return 0
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return intVal
}

func (c *Config) GetBoolKey(key string) bool {
	value := c.GetKey(key)
	if value == "" {
		return false
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return boolVal
}
