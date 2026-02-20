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
	if config.Get().IsCatalog() {
		registerCatalogManifestHandlers(ctx, version)
	}
	if config.Get().IsModuleEnabled(constants.HOST_MODE) {
		registerPackerTemplatesHandlers(ctx, version)
		registerVirtualMachinesHandlers(ctx, version)
	}
	registerConfigHandlers(ctx, version)
	if config.Get().IsOrchestrator() {
		registerOrchestratorHostsHandlers(ctx, version)
	}
	registerSshHandlers(ctx, version)
	registerPerformanceHandlers(ctx, version)
	registerReverseProxyHandlers(ctx, version)
	registerEventHandlers(ctx, version)

	return nil
}
