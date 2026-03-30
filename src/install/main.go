package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/logs"
	"github.com/cjlapao/common-go/helper"
)

// normalizeModules validates and normalizes a comma-separated modules string.
// "api" is always included. Unknown modules are silently dropped.
// Returns a deduplicated, sorted comma-separated string, or "api" if nothing valid was provided.
func normalizeModules(input string) string {
	validSet := make(map[string]struct{}, len(constants.VALID_MODULES))
	for _, m := range constants.VALID_MODULES {
		validSet[m] = struct{}{}
	}

	seen := make(map[string]struct{})
	seen[constants.API_MODE] = struct{}{} // api is always included

	for _, raw := range strings.Split(input, ",") {
		m := strings.TrimSpace(strings.ToLower(raw))
		if m == "" {
			continue
		}
		if _, ok := validSet[m]; ok {
			seen[m] = struct{}{}
		}
	}

	result := make([]string, 0, len(seen))
	// Emit in VALID_MODULES order for determinism
	for _, m := range constants.VALID_MODULES {
		if _, ok := seen[m]; ok {
			result = append(result, m)
		}
	}
	return strings.Join(result, ",")
}

func getRawFlagValue(flagName string) string {
	longFlag := "--" + strings.TrimSpace(flagName)

	for i, arg := range os.Args {
		if arg == longFlag && i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
			return os.Args[i+1]
		}
		if strings.HasPrefix(arg, longFlag+"=") {
			return strings.TrimPrefix(arg, longFlag+"=")
		}
	}

	return ""
}

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
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	unit, err := generateSystemdUnit(execPath, config)
	if err != nil {
		return fmt.Errorf("failed to generate systemd unit: %w", err)
	}

	unitPath := filepath.Join(LINUX_SYSTEMD_UNIT_DIR, LINUX_SYSTEMD_UNIT_NAME)

	// Stop and disable any existing instance before replacing the unit file.
	if helper.FileExists(unitPath) {
		if err := uninstallServiceOnLinux(ctx, false); err != nil {
			return err
		}
	}

	if err := helper.WriteToFile(unit, unitPath); err != nil {
		return fmt.Errorf("failed to write systemd unit to %s: %w", unitPath, err)
	}

	chmodCmd := helpers.Command{
		Command: "chmod",
		Args:    []string{"644", unitPath},
	}
	daemonReloadCmd := helpers.Command{
		Command: "systemctl",
		Args:    []string{"daemon-reload"},
	}
	enableCmd := helpers.Command{
		Command: "systemctl",
		Args:    []string{"enable", LINUX_SYSTEMD_UNIT_NAME},
	}
	startCmd := helpers.Command{
		Command: "systemctl",
		Args:    []string{"start", LINUX_SYSTEMD_UNIT_NAME},
	}

	for _, cmd := range []helpers.Command{chmodCmd, daemonReloadCmd, enableCmd, startCmd} {
		if _, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout); err != nil {
			return fmt.Errorf("command %q failed: %w", cmd.Command+" "+strings.Join(cmd.Args, " "), err)
		}
	}

	ctx.LogInfof("Service installed successfully")
	return nil
}

func uninstallServiceOnLinux(ctx basecontext.ApiContext, removeDatabase bool) error {
	unitPath := filepath.Join(LINUX_SYSTEMD_UNIT_DIR, LINUX_SYSTEMD_UNIT_NAME)

	stopCmd := helpers.Command{
		Command: "systemctl",
		Args:    []string{"stop", LINUX_SYSTEMD_UNIT_NAME},
	}
	disableCmd := helpers.Command{
		Command: "systemctl",
		Args:    []string{"disable", LINUX_SYSTEMD_UNIT_NAME},
	}
	daemonReloadCmd := helpers.Command{
		Command: "systemctl",
		Args:    []string{"daemon-reload"},
	}

	// Best-effort stop/disable — don't fail if the unit wasn't running.
	for _, cmd := range []helpers.Command{stopCmd, disableCmd, daemonReloadCmd} {
		if _, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout); err != nil {
			ctx.LogWarnf("systemctl %s warning (continuing): %v", strings.Join(cmd.Args, " "), err)
		}
	}

	if helper.FileExists(unitPath) {
		if err := os.Remove(unitPath); err != nil {
			ctx.LogWarnf("Could not remove unit file %s: %v", unitPath, err)
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

func getConfigFromEnv() ApiServiceConfig {
	cfg := config.Get()
	config := ApiServiceConfig{}
	if apiPort := getRawFlagValue("api-port"); apiPort != "" {
		config.Port = apiPort
	} else if cfg.GetKey(constants.API_PORT_ENV_VAR) != "" {
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
		config.TokenDurationMinutes = cfg.GetKey(constants.TOKEN_DURATION_MINUTES_ENV_VAR)
	}
	if modules := getRawFlagValue("modules"); modules != "" {
		config.EnabledModules = normalizeModules(modules)
	} else if cfg.GetKey(constants.ENABLED_MODULES_ENV_VAR) != "" {
		config.EnabledModules = normalizeModules(cfg.GetKey(constants.ENABLED_MODULES_ENV_VAR))
	} else if cfg.GetKey(constants.MODE_ENV_VAR) != "" {
		// Backward compatibility: MODE contains a single mode value
		config.EnabledModules = normalizeModules(cfg.GetKey(constants.MODE_ENV_VAR))
	}
	// If neither is set, leave EnabledModules empty so the plist omits the key
	// entirely and the runtime auto-detects modules based on OS and PD availability.
	if cfg.GetKey(constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR) != "" {
		config.UseOrchestratorResources = cfg.GetKey(constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR) == "true"
	}

	// CORS — enabled when the cors module is present, or when ENABLE_CORS is
	// explicitly set, or when no modules are specified (auto-detect at runtime
	// will include cors by default).
	corsInModules := strings.Contains(","+strings.ToLower(config.EnabledModules)+",", ","+constants.CORS_MODE+",")
	if corsInModules || cfg.GetBoolKey(constants.ENABLE_CORS_ENV_VAR) || config.EnabledModules == "" {
		config.EnableCors = true
		if cfg.GetKey(constants.CORS_ALLOWED_ORIGINS_ENV_VAR) != "" {
			config.CorsAllowedOrigins = cfg.GetKey(constants.CORS_ALLOWED_ORIGINS_ENV_VAR)
		} else {
			config.CorsAllowedOrigins = "*"
		}
		if cfg.GetKey(constants.CORS_ALLOWED_METHODS_ENV_VAR) != "" {
			config.CorsAllowedMethods = cfg.GetKey(constants.CORS_ALLOWED_METHODS_ENV_VAR)
		} else {
			config.CorsAllowedMethods = "GET,POST,PUT,DELETE,PATCH"
		}
		if cfg.GetKey(constants.CORS_ALLOWED_HEADERS_ENV_VAR) != "" {
			config.CorsAllowedHeaders = cfg.GetKey(constants.CORS_ALLOWED_HEADERS_ENV_VAR)
		} else {
			config.CorsAllowedHeaders = "X-Requested-With,Accept,Authorization,Content-Type,Content-Length,Accept-Encoding,X-CSRF-Token,Origin,Access-Control-Request-Method,Access-Control-Request-Headers,x-source-id,X-Source-Id"
		}
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

	// Normalise modules: prefer EnabledModules; fall back to Mode for backward compat.
	// If neither is set, leave empty so the plist omits the key and the runtime
	// auto-detects modules based on OS and PD availability.
	if result.EnabledModules != "" {
		result.EnabledModules = normalizeModules(result.EnabledModules)
	} else if result.Mode != "" {
		result.EnabledModules = normalizeModules(result.Mode)
	}

	// CORS — enabled when the cors module is present, or when EnableCors is
	// already set in the config file, or when no modules are specified.
	corsInModules := strings.Contains(","+strings.ToLower(result.EnabledModules)+",", ","+constants.CORS_MODE+",")
	if corsInModules || result.EnableCors || result.EnabledModules == "" {
		result.EnableCors = true
		if result.CorsAllowedOrigins == "" {
			result.CorsAllowedOrigins = "*"
		}
		if result.CorsAllowedMethods == "" {
			result.CorsAllowedMethods = "GET,POST,PUT,DELETE,PATCH"
		}
		if result.CorsAllowedHeaders == "" {
			result.CorsAllowedHeaders = "X-Requested-With,Accept,Authorization,Content-Type,Content-Length,Accept-Encoding,X-CSRF-Token,Origin,Access-Control-Request-Method,Access-Control-Request-Headers,x-source-id,X-Source-Id"
		}
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

// PersistEnabledModules rewrites the service config file (plist on macOS) with
// the current ENABLED_MODULES env var value so the module list survives a service
// restart.  The running service is NOT stopped or reloaded — the change takes
// effect the next time the service starts.
//
// This is called from startup when the host module is auto-enabled because
// Parallels Desktop is available but was missing from the stored module list
// (e.g. because the service started before PD finished installing).
func PersistEnabledModules(ctx basecontext.ApiContext) {
	switch runtime.GOOS {
	case "darwin":
		path, err := getExecutablePath()
		if err != nil {
			ctx.LogWarnf("[startup] Could not determine executable path for plist update: %v", err)
			return
		}
		cfg := getConfigFromEnv()
		plist, err := generatePlist(path, cfg)
		if err != nil {
			ctx.LogWarnf("[startup] Could not generate updated plist: %v", err)
			return
		}
		daemonPath := filepath.Join(MAC_PLIST_DAEMON_PATH, MAC_PLIST_DAEMON_NAME)
		if err := helper.WriteToFile(plist, daemonPath); err != nil {
			ctx.LogWarnf("[startup] Could not write updated plist to %s: %v", daemonPath, err)
			return
		}
		ctx.LogInfof("[startup] Persisted updated module list to %s", daemonPath)
	}
}
