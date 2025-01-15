package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/logs"
	"github.com/cjlapao/common-go/helper"
)

const (
	MAC_PLIST_DAEMON_PATH = "/Library/LaunchDaemons"
	MAC_PLIST_DAEMON_NAME = "com.parallels.prl-devops-service.plist"
)

func InstallService(ctx basecontext.ApiContext, configFilePath string) error {
	ctx.LogInfof("Installing service...")
	logs.SetupFileLogger(fmt.Sprintf("install_%v.log", time.Now().Unix()), ctx)
	var config ApiServiceConfig
	var err error
	if configFilePath != "" {
		config, err = getConfigFromFile(configFilePath)
		if err != nil {
			return err
		}
	} else {
		config = getConfigFromEnv()
	}

	switch os := runtime.GOOS; os {
	case "darwin":
		return installServiceOnMac(ctx, config)
	case "windows":
		return installServiceOnWindows(ctx, config)
	case "linux":
		return installServiceOnLinux(ctx, config)
	default:
		errMsg := fmt.Sprintf("unsupported operating system: %s", os)
		ctx.LogErrorf(errMsg)
		return errors.New(errMsg)
	}
}

func UninstallService(ctx basecontext.ApiContext, removeDatabase bool) error {
	logs.SetupFileLogger(fmt.Sprintf("uninstall_%v.log", time.Now().Unix()), ctx)
	switch os := runtime.GOOS; os {
	case "darwin":
		return uninstallServiceOnMac(ctx, removeDatabase)
	case "windows":
		return uninstallServiceOnWindows(ctx, removeDatabase)
	case "linux":
		return uninstallServiceOnLinux(ctx, removeDatabase)
	default:
		errMsg := fmt.Sprintf("unsupported operating system: %s", os)
		ctx.LogErrorf(errMsg)
		return errors.New(errMsg)
	}
}

func IsInstalled(ctx basecontext.ApiContext) bool {
	return false
}

func installServiceOnMac(ctx basecontext.ApiContext, config ApiServiceConfig) error {
	path, err := getExecutablePath()
	if err != nil {
		return err
	}

	plist, err := generatePlist(path, config)
	if err != nil {
		return err
	}

	if !helper.FileExists(MAC_PLIST_DAEMON_PATH) {
		return errors.New("daemon path does not exist")
	}

	daemonPath := filepath.Join(MAC_PLIST_DAEMON_PATH, MAC_PLIST_DAEMON_NAME)

	// Unload the daemon if it is already loaded
	if helper.FileExists(daemonPath) {
		if err := uninstallServiceOnMac(ctx, false); err != nil {
			return err
		}
	}

	if err := helper.WriteToFile(plist, daemonPath); err != nil {
		return err
	}

	chownCmd := helpers.Command{
		Command: "sudo",
		Args:    []string{"chown", "root:wheel", daemonPath},
	}
	chmod := helpers.Command{
		Command: "sudo",
		Args:    []string{"chmod", "644", daemonPath},
	}

	launchdLoadCmd := helpers.Command{
		Command: "sudo",
		Args:    []string{"launchctl", "load", daemonPath},
	}

	if _, err := helpers.ExecuteWithNoOutput(ctx.Context(), chownCmd, helpers.ExecutionTimeout); err != nil {
		return err
	}
	if _, err := helpers.ExecuteWithNoOutput(ctx.Context(), chmod, helpers.ExecutionTimeout); err != nil {
		return err
	}
	if _, err := helpers.ExecuteWithNoOutput(ctx.Context(), launchdLoadCmd, helpers.ExecutionTimeout); err != nil {
		return err
	}

	ctx.LogInfof("Service installed successfully")

	return nil
}

func uninstallServiceOnMac(ctx basecontext.ApiContext, removeDatabase bool) error {
	daemonPath := filepath.Join(MAC_PLIST_DAEMON_PATH, MAC_PLIST_DAEMON_NAME)

	cmd := helpers.Command{
		Command: "sudo",
		Args:    []string{"launchctl", "unload", daemonPath},
	}

	if _, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout); err != nil {
		return err
	}

	if helper.FileExists(daemonPath) {
		if err := os.Remove(daemonPath); err != nil {
			ctx.LogWarnf("There was an error removing the daemon file at %v", daemonPath)
		}
	}

	if removeDatabase && helper.FileExists(constants.ServiceDefaultDirectory) {
		if err := os.RemoveAll(constants.ServiceDefaultDirectory); err != nil {
			return err
		}
	}

	ctx.LogInfof("Service uninstalled successfully")
	return nil
}

func installServiceOnWindows(ctx basecontext.ApiContext, config ApiServiceConfig) error {
	return errors.New("not implemented")
}

func uninstallServiceOnWindows(ctx basecontext.ApiContext, removeDatabase bool) error {
	return errors.New("not implemented")
}

func installServiceOnLinux(ctx basecontext.ApiContext, config ApiServiceConfig) error {
	return errors.New("not implemented")
}

func uninstallServiceOnLinux(ctx basecontext.ApiContext, removeDatabase bool) error {
	return errors.New("not implemented")
}

func getConfigFromEnv() ApiServiceConfig {
	cfg := config.Get()
	config := ApiServiceConfig{}
	if cfg.GetKey(constants.API_PORT_ENV_VAR) != "" {
		config.Port = cfg.GetKey(constants.API_PORT_ENV_VAR)
	} else {
		config.Port = constants.DEFAULT_API_PORT
	}
	if cfg.GetKey(constants.API_PREFIX_ENV_VAR) != "" {
		config.Prefix = cfg.GetKey(constants.API_PREFIX_ENV_VAR)
	} else {
		config.Prefix = constants.DEFAULT_API_PREFIX
	}
	if cfg.GetKey(constants.LOG_LEVEL_ENV_VAR) != "" {
		config.LogLevel = cfg.GetKey(constants.LOG_LEVEL_ENV_VAR)
	} else {
		config.LogLevel = "INFO"
	}
	if cfg.GetKey(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR) != "" {
		config.EncryptionRsaKey = cfg.GetKey(constants.ENCRYPTION_SECURITY_KEY_ENV_VAR)
	}
	if cfg.GetKey(constants.HMAC_SECRET_ENV_VAR) != "" {
		config.HmacSecret = cfg.GetKey(constants.HMAC_SECRET_ENV_VAR)
	}
	if cfg.GetKey(constants.TLS_ENABLED_ENV_VAR) != "" {
		config.EnableTLS = cfg.GetKey(constants.TLS_ENABLED_ENV_VAR) == "true"
	} else {
		config.EnableTLS = false
	}
	if cfg.GetKey(constants.TLS_CERTIFICATE_ENV_VAR) != "" {
		config.TLSPrivateKey = cfg.GetKey(constants.TLS_CERTIFICATE_ENV_VAR)
	}
	if cfg.GetKey(constants.TLS_PRIVATE_KEY_ENV_VAR) != "" {
		config.TLSPrivateKey = cfg.GetKey(constants.TLS_PRIVATE_KEY_ENV_VAR)
	}
	if cfg.GetKey(constants.ROOT_PASSWORD_ENV_VAR) != "" {
		config.RootPassword = cfg.GetKey(constants.ROOT_PASSWORD_ENV_VAR)
	}
	if cfg.GetKey(constants.DISABLE_CATALOG_CACHING_ENV_VAR) != "" {
		config.DisableCatalogCaching = cfg.GetKey(constants.ROOT_PASSWORD_ENV_VAR) == "true"
	}
	if cfg.GetKey(constants.TOKEN_DURATION_MINUTES_ENV_VAR) != "" {
		config.DisableCatalogCaching = cfg.GetKey(constants.TOKEN_DURATION_MINUTES_ENV_VAR) == "true"
	}
	if cfg.GetKey(constants.MODE_ENV_VAR) != "" {
		config.Mode = cfg.GetKey(constants.MODE_ENV_VAR)
	}
	if cfg.GetKey(constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR) != "" {
		config.UseOrchestratorResources = cfg.GetKey(constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR) == "true"
	}

	return config
}

func getConfigFromFile(filePath string) (ApiServiceConfig, error) {
	result := ApiServiceConfig{}

	if !helper.FileExists(filePath) {
		return result, errors.New("config file does not exist")
	}

	content, err := helper.ReadFromFile(filePath)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(content, &result); err != nil {
		return result, err
	}

	return result, nil
}

func getExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}
