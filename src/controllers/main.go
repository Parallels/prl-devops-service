package controllers

import (
	"github.com/Parallels/pd-api-service/basecontext"
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

	return nil
}
