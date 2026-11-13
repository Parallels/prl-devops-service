package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"

	"gopkg.in/yaml.v3"

	log "github.com/cjlapao/common-go-logger"
	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/security"
)

var globalConfig *Config

var (
	extensions = []string{
		".local.yaml",
		".local.yml",
		".local.json",
		".yaml",
		".yml",
		".json",
	}
)

func findConfigFile(folder, baseName string) string {
	for _, ext := range extensions {
		candidate := filepath.Join(folder, baseName+ext)
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			return candidate
		}
	}
	return ""
}

func resolveConfigFile(configFileName, execPath string) string {
	if configFileName != "" {
		if _, err := os.Stat(configFileName); !os.IsNotExist(err) {
			return configFileName
		}
		return ""
	}

	searchLocations := []string{
		filepath.Dir(execPath),
		filepath.Join(filepath.Dir(execPath), "prldevops_config"),
		constants.ServiceDefaultDirectory,
	}
	baseNames := []string{"prldevops_config", "config"}
	for _, location := range searchLocations {
		for _, baseName := range baseNames {
			if fileName := findConfigFile(location, baseName); fileName != "" {
				return fileName
			}
		}
	}
	return ""
}

type Config struct {
	ctx                 basecontext.ApiContext
	mode                string
	command             string
	includeOwnResources bool
	fileFormat          string
	filename            string
	isLoaded            bool
	config              ConfigFile
}

func New(ctx basecontext.ApiContext) *Config {
	globalConfig = &Config{
		mode:       "api",
		ctx:        ctx,
		fileFormat: "yaml",
		filename:   "prldevops_config",
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
	if c.isLoaded {
		return true
	}

	execPath, err := os.Executable()
	if err != nil {
		c.ctx.LogErrorf("Error getting executable path: %s", err.Error())
		c.isLoaded = false
		return false
	}

	fileName := resolveConfigFile(
		helper.GetFlagValue(constants.CONFIG_FILE_FLAG, ""),
		execPath,
	)

	if fileName == "" {
		c.ctx.LogInfof("No configuration file found")
		c.isLoaded = false
		return false
	}

	c.ctx.LogInfof("Loading configuration from %s", fileName)
	content, err := helper.ReadFromFile(fileName)
	if err != nil {
		c.ctx.LogErrorf("Error reading configuration file: %s", err.Error())
		c.isLoaded = false
		return false
	}

	if content == nil {
		c.ctx.LogErrorf("Error reading configuration file: content is nil")
		c.isLoaded = false
		return false
	}

	if strings.HasSuffix(fileName, ".json") {
		err = json.Unmarshal(content, &c.config)
		if err != nil {
			c.ctx.LogErrorf("Error reading configuration file: %s", err.Error())
			c.isLoaded = false
			return false
		}
		c.fileFormat = "json"
	} else {
		err = yaml.Unmarshal(content, &c.config)
		if err != nil {
			c.ctx.LogErrorf("Error reading configuration file: %s", err.Error())
			c.isLoaded = false
			return false
		}
		c.fileFormat = "yaml"
	}

	c.LogLevel(false)
	c.filename = fileName
	c.isLoaded = true
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

	filename := strings.TrimSuffix(c.filename, filepath.Ext(c.filename))
	err = helper.WriteToFile(string(content), fmt.Sprintf("%s.%s", filename, c.fileFormat))
	if err != nil {
		c.ctx.LogErrorf("Error saving configuration file: %s", err.Error())
		return false
	}

	return true
}

func (c *Config) Refresh() bool {
	c.isLoaded = false
	return c.Load()
}

func (c *Config) CleanReverseProxyDataFromConfig() error {
	c.config.ReverseProxy = nil
	return nil
}

func (c *Config) LogLevel(silent bool) string {
	logLevel := c.GetKey(constants.LOG_LEVEL_ENV_VAR)
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
	default:
		common.Logger.LogLevel = log.Info
		if logLevel != "" && !silent {
			common.Logger.Warn("Invalid log level %s, defaulting to INFO", logLevel)
		}
	}
	if logLevel != "" && !silent {
		common.Logger.Info("Log Level set to %v", common.Logger.LogLevel)
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

func (c *Config) RootPassword() string {
	password := c.GetKey(constants.ROOT_PASSWORD_ENV_VAR)
	if password == "" {
		return ""
	}
	return password
}

func (c *Config) Source() string {
	source := c.GetKey(constants.SOURCE_ENV_VAR)
	if source == "" {
		source = c.Mode()
	}

	return source
}

func (c *Config) SetRunningCommand(command string) {
	c.command = command
}

func (c *Config) GetRunningCommand() string {
	return c.command
}

// OrchestratorVmRefetchInterval returns how often the orchestrator live-fetches VM
// data from its hosts (bypassing host caches). Defaults to 5 minutes.
func (c *Config) OrchestratorVmRefetchInterval() time.Duration {
	intervalEnvValue := c.GetKey(constants.ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS_ENV_VAR)
	if intervalEnvValue == "" {
		return time.Duration(constants.DEFAULT_ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS) * time.Second
	}
	return time.Duration(c.GetIntKey(constants.ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS_ENV_VAR)) * time.Second
}

func (c *Config) GetKey(key string) string {
	value := ""
	for k, v := range c.config.Environment {
		if strings.EqualFold(k, key) {
			value = v
			break
		}
	}

	if flagValue := helper.GetFlagValue(key, ""); flagValue != "" {
		value = flagValue
	}

	if envValue, exists := os.LookupEnv(key); exists {
		value = envValue
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

func (c *Config) SetKey(key string, value string) {
	if c.config.Environment == nil {
		c.config.Environment = make(map[string]string)
	}
	c.config.Environment[strings.ToLower(key)] = value
}

func (c *Config) SetIntKey(key string, value int) {
	c.SetKey(key, strconv.Itoa(value))
}

func (c *Config) SetBoolKey(key string, value bool) {
	c.SetKey(key, strconv.FormatBool(value))
}

func (c *Config) GetConfig() *ConfigFile {
	return &c.config
}

func (c *Config) GetFilename() string {
	return c.filename
}

func (c *Config) GetConfigContent() ([]byte, error) {
	switch c.fileFormat {
	case "json":
		return json.Marshal(c.config)
	case "yaml":
		return yaml.Marshal(c.config)
	default:
		return yaml.Marshal(c.config)
	}
}
