package controllers

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
)

func RegisterV1Handlers(ctx basecontext.ApiContext) error {
	version := "v1"
	registerAuthorizationHandlers(ctx, version)
	registerUsersHandlers(ctx, version)
	registerApiKeysHandlers(ctx, version)
	registerClaimsHandlers(ctx, version)
	registerRolesHandlers(ctx, version)
	if config.Get().IsCatalogManager() {
		registerCatalogManagerHandlers(ctx, version)
	}
	registerJobsHandlers(ctx, version)

	if config.Get().IsHost() {
		registerCacheHandlers(ctx, version)
		// Host-only orchestrator push channel (POST /settings/apply).
		registerHostSettingsHandlers(ctx, version)
	}
	if config.Get().IsCatalog() {
		registerCatalogManifestHandlers(ctx, version)
	}
	if config.Get().IsModuleEnabled(constants.HOST_MODE) {
		registerPackerTemplatesHandlers(ctx, version)
		registerVirtualMachinesHandlers(ctx, version)
	}
	registerConfigHandlers(ctx, version)
	registerUserConfigsHandlers(ctx, version)
	if config.Get().IsOrchestrator() {
		registerOrchestratorHostsHandlers(ctx, version)
		registerOrchestratorMachineCheckHandlers(ctx, version)
		// Self console (definitions + values) — identical paths on hosts,
		// where "self" means the local layer instead.
		registerSettingsHandlers(ctx, version)
		// Privileged console for configuring OTHER nodes.
		registerOrchestratorSettingsHandlers(ctx, version)
	} else if config.Get().IsHost() {
		// Host running as a service (directly-connected UI): self console,
		// no privileged routes — a host never manages other nodes.
		registerSettingsHandlers(ctx, version)
	}
	registerSshHandlers(ctx, version)
	registerPerformanceHandlers(ctx, version)
	if config.Get().IsReverseProxyEnabled() {
		registerReverseProxyHandlers(ctx, version)
	}
	registerEventHandlers(ctx, version)

	return nil
}
