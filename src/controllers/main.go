package controllers

import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

func RegisterV1Handlers(ctx basecontext.ApiContext) error {
	version := "v1"
	registerAuthorizationHandlers(ctx, version)
	registerUsersHandlers(ctx, version)
	registerApiKeysHandlers(ctx, version)
	registerClaimsHandlers(ctx, version)
	registerRolesHandlers(ctx, version)
	registerCatalogManifestHandlers(ctx, version)
	registerPackerTemplatesHandlers(ctx, version)
	registerVirtualMachinesHandlers(ctx, version)
	registerConfigHandlers(ctx, version)
	registerOrchestratorHostsHandlers(ctx, version)
	registerPerformanceHandlers(ctx, version)
	registerReverseProxyHandlers(ctx, version)
	registerEventHandlers(ctx, version)

	return nil
}
