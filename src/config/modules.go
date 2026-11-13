package config

// Cohesive getter group moved from main.go (pure move, no behaviour
// change) - see .ai-agents/features/417-gorm-database-layer/420-database-configuration/plan.md D-6

import (
	"os"
	"runtime"
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
)

func (c *Config) IsOrchestrator() bool {
	return c.IsModuleEnabled(constants.ORCHESTRATOR_MODE)
}

func (c *Config) IsCatalog() bool {
	return c.IsModuleEnabled(constants.CATALOG_MODE)
}

func (c *Config) IsCatalogManager() bool {
	return c.IsModuleEnabled(constants.CATALOG_MANAGER_MODE)
}

func (c *Config) IsHost() bool {
	return c.IsModuleEnabled(constants.HOST_MODE)
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

func (c *Config) GetEnabledModules() []string {
	onDarwin := runtime.GOOS == "darwin"

	modules := c.GetKey(constants.ENABLED_MODULES_ENV_VAR)
	modulesList := []string{}

	if modules != "" {
		modulesList = strings.Split(modules, ",")
		for i, module := range modulesList {
			modulesList[i] = strings.TrimSpace(module)
		}
	} else {
		// Auto-detect based on OS and MODE when ENABLED_MODULES is not set.
		modulesList = append(modulesList, constants.API_MODE)
		modulesList = append(modulesList, constants.CORS_MODE) // enabled by default on all platforms
		// host is only meaningful on darwin (requires Parallels Desktop).
		if onDarwin {
			modulesList = append(modulesList, constants.HOST_MODE)
		}
		mode := c.Mode()
		if strings.EqualFold(mode, constants.CATALOG_MODE) {
			modulesList = append(modulesList, constants.CATALOG_MODE)
		} else if strings.EqualFold(mode, constants.ORCHESTRATOR_MODE) {
			modulesList = append(modulesList, constants.ORCHESTRATOR_MODE)
		}
	}

	// Validate modules against the known-valid list.
	validModules := []string{}
	for _, module := range modulesList {
		for _, validModule := range constants.VALID_MODULES {
			if strings.EqualFold(module, validModule) {
				validModules = append(validModules, module)
				break
			}
		}
	}
	modulesList = validModules

	// Ensure API is always enabled.
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

	// On non-darwin systems, remove modules that require the macOS/PD stack.
	// This enforces the rule even when ENABLED_MODULES is set explicitly.
	if !onDarwin {
		darwinOnly := map[string]bool{
			strings.ToLower(constants.HOST_MODE):          true,
			strings.ToLower(constants.CACHE_MODE):         true,
			strings.ToLower(constants.REVERSE_PROXY_MODE): true,
		}
		filtered := []string{}
		for _, module := range modulesList {
			if !darwinOnly[strings.ToLower(module)] {
				filtered = append(filtered, module)
			}
		}
		modulesList = filtered
	}

	// On darwin: auto-add cache unless explicitly disabled via env var.
	if onDarwin && c.IsCatalogCachingEnable() {
		cacheFound := false
		for _, module := range modulesList {
			if strings.EqualFold(module, constants.CACHE_MODE) {
				cacheFound = true
				break
			}
		}
		if !cacheFound {
			modulesList = append(modulesList, constants.CACHE_MODE)
		}
	}

	// On darwin: auto-add reverse_proxy unless explicitly disabled via env var.
	if onDarwin && c.IsReverseProxyEnable() {
		proxyFound := false
		for _, module := range modulesList {
			if strings.EqualFold(module, constants.REVERSE_PROXY_MODE) {
				proxyFound = true
				break
			}
		}
		if !proxyFound {
			modulesList = append(modulesList, constants.REVERSE_PROXY_MODE)
		}
	}

	// catalog_manager is disabled only when the node is a pure catalog server —
	// i.e. catalog is enabled but neither host nor orchestrator is.
	// If host or orchestrator accompanies catalog the node is also an agent/
	// provider and still needs catalog_manager.
	catalogActive := false
	hostOrOrchestratorActive := false
	for _, module := range modulesList {
		switch strings.ToLower(module) {
		case constants.CATALOG_MODE:
			catalogActive = true
		case constants.HOST_MODE, constants.ORCHESTRATOR_MODE:
			hostOrOrchestratorActive = true
		}
	}
	suppressCatalogManager := catalogActive && !hostOrOrchestratorActive
	if suppressCatalogManager {
		// Strip catalog_manager even if it was listed explicitly.
		filtered := modulesList[:0]
		for _, module := range modulesList {
			if !strings.EqualFold(module, constants.CATALOG_MANAGER_MODE) {
				filtered = append(filtered, module)
			}
		}
		modulesList = filtered
	} else {
		catalogManagerFound := false
		for _, module := range modulesList {
			if strings.EqualFold(module, constants.CATALOG_MANAGER_MODE) {
				catalogManagerFound = true
				break
			}
		}
		if !catalogManagerFound {
			modulesList = append(modulesList, constants.CATALOG_MANAGER_MODE)
		}
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
