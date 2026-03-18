package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	str "github.com/stretchr/testify/assert"
)

// onDarwin reports whether the current platform is macOS.  The darwin-only
// modules (host, cache, reverse_proxy) are stripped on every other OS, so
// tests that check for them must guard with this helper.
var onDarwin = runtime.GOOS == "darwin"

func TestGetEnabledModules_Fallback(t *testing.T) {
	// Setup
	os.Unsetenv(constants.ENABLED_MODULES_ENV_VAR)
	os.Setenv(constants.MODE_ENV_VAR, "api")
	ctx := basecontext.NewBaseContext()
	cfg := New(ctx)

	// Test Fallback to MODE=api
	modules := cfg.GetEnabledModules()
	str.Contains(t, modules, "api")
	str.Contains(t, modules, constants.CORS_MODE)
	if onDarwin {
		str.Contains(t, modules, "host")
	} else {
		str.NotContains(t, modules, "host")
	}
	str.NotContains(t, modules, "catalog")
	str.NotContains(t, modules, "orchestrator")

	// Test Fallback to MODE=catalog
	os.Setenv(constants.MODE_ENV_VAR, "catalog")
	cfg = New(ctx)
	modules = cfg.GetEnabledModules()
	str.Contains(t, modules, "api")
	str.Contains(t, modules, "catalog")
	if onDarwin {
		str.Contains(t, modules, "host")
	} else {
		str.NotContains(t, modules, "host")
	}
	str.NotContains(t, modules, "orchestrator")

	// Test Fallback to MODE=orchestrator with different case
	os.Setenv(constants.MODE_ENV_VAR, "Orchestrator")
	cfg = New(ctx)
	modules = cfg.GetEnabledModules()
	str.Contains(t, modules, "api")
	str.Contains(t, modules, "orchestrator")
	if onDarwin {
		str.Contains(t, modules, "host")
	} else {
		str.NotContains(t, modules, "host")
	}
	str.NotContains(t, modules, "catalog")
}

func TestGetEnabledModules_ReverseProxyFallback(t *testing.T) {
	// Setup
	os.Unsetenv(constants.ENABLED_MODULES_ENV_VAR)
	os.Setenv(constants.MODE_ENV_VAR, "api")
	os.Unsetenv(constants.DISABLE_REVERSE_PROXY_ENV_VAR) // Proxy enabled by default
	ctx := basecontext.NewBaseContext()
	cfg := New(ctx)

	// reverse_proxy is auto-added on darwin when not disabled; stripped on other OSes.
	modules := cfg.GetEnabledModules()
	str.Contains(t, modules, "api")
	if onDarwin {
		str.Contains(t, modules, constants.REVERSE_PROXY_MODE)
		str.True(t, cfg.IsReverseProxyEnabled())
	} else {
		str.NotContains(t, modules, constants.REVERSE_PROXY_MODE)
		str.False(t, cfg.IsReverseProxyEnabled())
	}

	// Disable reverse proxy explicitly — should never appear regardless of OS.
	os.Setenv(constants.DISABLE_REVERSE_PROXY_ENV_VAR, "true")
	cfg = New(ctx)
	modules = cfg.GetEnabledModules()
	str.NotContains(t, modules, constants.REVERSE_PROXY_MODE)
	str.False(t, cfg.IsReverseProxyEnabled())
}

func TestGetEnabledModules_Explicit(t *testing.T) {
	// Setup — explicitly request api, catalog, and reverse_proxy.
	// On non-darwin systems, reverse_proxy is stripped even when explicitly set.
	os.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api,catalog,reverse_proxy")
	ctx := basecontext.NewBaseContext()
	cfg := New(ctx)

	modules := cfg.GetEnabledModules()
	str.Contains(t, modules, "api")
	str.Contains(t, modules, "catalog")
	str.NotContains(t, modules, "host")
	str.NotContains(t, modules, "orchestrator")
	if onDarwin {
		str.Contains(t, modules, constants.REVERSE_PROXY_MODE)
		str.True(t, cfg.IsReverseProxyEnabled())
	} else {
		str.NotContains(t, modules, constants.REVERSE_PROXY_MODE)
		str.False(t, cfg.IsReverseProxyEnabled())
	}
}

func TestGetEnabledModules_EnsureApi(t *testing.T) {
	// Setup
	os.Setenv(constants.ENABLED_MODULES_ENV_VAR, "catalog")
	ctx := basecontext.NewBaseContext()
	cfg := New(ctx)

	// Test Ensure API
	modules := cfg.GetEnabledModules()
	str.Contains(t, modules, "api")
	str.Contains(t, modules, "catalog")
}

func TestGetEnabledModules_Validation(t *testing.T) {
	// Setup
	os.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api,catalog,invalid_module")
	ctx := basecontext.NewBaseContext()
	cfg := New(ctx)

	// Test Validation — invalid module is dropped
	modules := cfg.GetEnabledModules()
	str.Contains(t, modules, "api")
	str.Contains(t, modules, "catalog")
	str.NotContains(t, modules, "invalid_module")
}

func TestIsModuleEnabled(t *testing.T) {
	os.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api,catalog")
	ctx := basecontext.NewBaseContext()
	cfg := New(ctx)

	str.True(t, cfg.IsModuleEnabled("api"))
	str.True(t, cfg.IsModuleEnabled("catalog"))
	str.False(t, cfg.IsModuleEnabled("host"))
}

func TestDisableModule(t *testing.T) {
	os.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api,catalog")
	ctx := basecontext.NewBaseContext()
	cfg := New(ctx)

	cfg.DisableModule("catalog")
	str.False(t, cfg.IsModuleEnabled("catalog"))
	str.True(t, cfg.IsModuleEnabled("api"))
	str.Contains(t, os.Getenv(constants.ENABLED_MODULES_ENV_VAR), "api")
	str.NotContains(t, os.Getenv(constants.ENABLED_MODULES_ENV_VAR), "catalog")
}

func TestEnableModule(t *testing.T) {
	os.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api")
	ctx := basecontext.NewBaseContext()
	cfg := New(ctx)

	cfg.EnableModule("catalog")
	str.True(t, cfg.IsModuleEnabled("catalog"))
	str.True(t, cfg.IsModuleEnabled("api"))
	str.Contains(t, os.Getenv(constants.ENABLED_MODULES_ENV_VAR), "catalog")
}
