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
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"gopkg.in/yaml.v3"

	log "github.com/cjlapao/common-go-logger"
	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/security"
)

var globalConfig *Config

var (
	canaryBuildFlag = "false"
	betaBuildFlag   = "false"
	extensions      = []string{
		".local.yaml",
		".local.yml",
		".local.json",
		".yaml",
		".yml",
		".json",
	}
)

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

	fileName := ""
	configFileName := helper.GetFlagValue(constants.CONFIG_FILE_FLAG, "")

	execPath, err := os.Executable()
	if err != nil {
		c.ctx.LogErrorf("Error getting executable path: %s", err.Error())
		c.isLoaded = false
		return false
	}
	if configFileName != "" {
		if _, err := os.Stat(configFileName); !os.IsNotExist(err) {
			fileName = configFileName
		}
	} else {
		configFolder := filepath.Join(filepath.Dir(execPath), "prldevops_config")
		for _, extension := range extensions {
			if _, err := os.Stat(fmt.Sprintf("%s%s", configFolder, extension)); !os.IsNotExist(err) {
				fileName = fmt.Sprintf("%s%s", configFolder, extension)
				break
			}
		}
		if fileName == "" {
			configFolder = filepath.Join(filepath.Dir(execPath), "config")
			for _, extension := range extensions {
				if _, err := os.Stat(fmt.Sprintf("%s%s", configFolder, extension)); !os.IsNotExist(err) {
					fileName = fmt.Sprintf("%s%s", configFolder, extension)
					break
				}
			}
		}
	}

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

func (c *Config) IsCorsEnabled() bool {
	corsEnabled := c.GetBoolKey(constants.ENABLE_CORS_ENV_VAR)

	return corsEnabled
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

func (c *Config) Source() string {
	source := c.GetKey(constants.SOURCE_ENV_VAR)
	if source == "" {
		source = c.Mode()
	}

	return source
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
	return c.IsModuleEnabled(constants.ORCHESTRATOR_MODE)
}

func (c *Config) IsCatalog() bool {
	return c.IsModuleEnabled(constants.CATALOG_MODE)
}

func (c *Config) IsApi() bool {
	return c.IsModuleEnabled(constants.API_MODE)
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

func (c *Config) SetRunningCommand(command string) {
	c.command = command
}

func (c *Config) GetRunningCommand() string {
	return c.command
}

func (c *Config) DisableTlsValidation() bool {
	val := c.GetBoolKey(constants.TLS_DISABLE_VALIDATION_ENV_VAR)

	return val
}

func (c *Config) IsReverseProxyEnabled() bool {
	return c.IsModuleEnabled(constants.REVERSE_PROXY_MODE)
}

func (c *Config) ReverseProxyHost() string {
	host := c.GetKey(constants.REVERSE_PROXY_HOST_ENV_VAR)
	if host == "" {
		host = constants.DEFAULT_REVERSE_PROXY_HOST
	}

	return host
}

func (c *Config) ReverseProxyPort() string {
	port := c.GetKey(constants.REVERSE_PROXY_PORT_ENV_VAR)
	if port == "" {
		port = constants.DEFAULT_REVERSE_PROXY_PORT
	}

	return port
}

func (c *Config) GetReverseProxyConfig() *ReverseProxyConfig {
	return c.config.ReverseProxy
}

func (c *Config) EnableReverseProxy(value bool) bool {
	c.SetKey(constants.ENABLE_REVERSE_PROXY_ENV_VAR, strconv.FormatBool(value))
	return true
}

func (c *Config) IsRemoteProviderStreamEnabled() bool {
	streamingDisabled := c.GetKey(constants.DISABLE_CATALOG_PROVIDER_STREAMING_ENV_VAR)
	if streamingDisabled == "" {
		return true
	}

	return !helpers.StringToBool(streamingDisabled)
}

func (c *Config) IsCanaryEnabled() bool {
	if canaryBuildFlag == "true" {
		return true
	}

	enableCanary := c.GetBoolKey(constants.ENABLE_CANARY_ENV_VAR)
	return enableCanary
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
	if betaBuildFlag == "true" {
		return true
	}

	enableBeta := c.GetBoolKey(constants.ENABLE_BETA_ENV_VAR)
	return enableBeta
}

func (c *Config) IsForceCacheRefresh() bool {
	forceCacheRefreshEnvValue := c.GetKey(constants.VM_FORCE_CACHE_REFRESH_ENV_VAR)
	if forceCacheRefreshEnvValue == "" {
		return constants.VM_FORCE_CACHE_REFRESH
	}
	return c.GetBoolKey(constants.VM_FORCE_CACHE_REFRESH_ENV_VAR)
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

func (c *Config) GetEnabledModules() []string {
	modules := c.GetKey(constants.ENABLED_MODULES_ENV_VAR)
	modulesList := []string{}

	if modules != "" {
		modulesList = strings.Split(modules, ",")
		for i, module := range modulesList {
			modulesList[i] = strings.TrimSpace(module)
		}
	} else {
		// Fallback to MODE if ENABLED_MODULES is empty
		// API and Host are enabled by default in all modes (Host subject to availability check later)
		modulesList = append(modulesList, "api")
		modulesList = append(modulesList, "host")

		mode := c.Mode()
		if strings.EqualFold(mode, constants.CATALOG_MODE) {
			modulesList = append(modulesList, "catalog")
		} else if strings.EqualFold(mode, constants.ORCHESTRATOR_MODE) {
			modulesList = append(modulesList, "orchestrator")
		}
	}

	reverseProxyEnabled := c.GetKey(constants.ENABLE_REVERSE_PROXY_ENV_VAR)
	if strings.EqualFold(reverseProxyEnabled, "true") || reverseProxyEnabled == "1" {
		found := false
		for _, m := range modulesList {
			if strings.EqualFold(m, constants.REVERSE_PROXY_MODE) {
				found = true
				break
			}
		}
		if !found {
			modulesList = append(modulesList, constants.REVERSE_PROXY_MODE)
		}
	}

	// Validate modules
	validModules := []string{}
	for _, module := range modulesList {
		isValid := false
		for _, validModule := range constants.VALID_MODULES {
			if strings.EqualFold(module, validModule) {
				isValid = true
				break
			}
		}
		if isValid {
			validModules = append(validModules, module)
		}
	}
	modulesList = validModules

	// Ensure API is always enabled
	apiFound := false
	for _, module := range modulesList {
		if strings.EqualFold(module, constants.API_MODE) {
			apiFound = true
			break
		}
	}

	if !apiFound {
		modulesList = append(modulesList, constants.API_MODE)
	}

	return modulesList
}

func (c *Config) IsModuleEnabled(module string) bool {
	modules := c.GetEnabledModules()
	for _, m := range modules {
		if strings.EqualFold(m, module) {
			return true
		}
	}

	return false
}

func (c *Config) DisableModule(module string) {
	modules := c.GetEnabledModules()
	newModules := []string{}
	for _, m := range modules {
		if !strings.EqualFold(m, module) {
			newModules = append(newModules, m)
		}
	}

	newValue := strings.Join(newModules, ",")
	c.SetKey(constants.ENABLED_MODULES_ENV_VAR, newValue)
	_ = os.Setenv(constants.ENABLED_MODULES_ENV_VAR, newValue)
}

func (c *Config) EnableModule(module string) {
	if c.IsModuleEnabled(module) {
		return
	}

	modules := c.GetEnabledModules()
	modules = append(modules, module)
	newValue := strings.Join(modules, ",")
	c.SetKey(constants.ENABLED_MODULES_ENV_VAR, newValue)
	_ = os.Setenv(constants.ENABLED_MODULES_ENV_VAR, newValue)
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
