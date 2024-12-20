package tests

import (
	"encoding/base64"
	"encoding/json"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	"github.com/Parallels/prl-devops-service/catalog/cacheservice"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
)

const (
	TEST_CACHE_CATALOG_ID           = "TEST_CACHE_CATALOG_ID"
	TEST_CACHE_CATALOG_VERSION      = "TEST_CACHE_CATALOG_VERSION"
	TEST_CACHE_CATALOG_ARCH         = "TEST_CACHE_CATALOG_ARCH"
	TEST_CACHE_CATALOG_MACHINE_NAME = "TEST_CACHE_CATALOG_MACHINE_NAME"
	TEST_CACHE_CATALOG_CONNECTION   = "TEST_CACHE_CATALOG_CONNECTION"
	TEST_CACHE_REMOTE_FILENAME      = "TEST_CACHE_REMOTE_FILENAME"
	TEST_CACHE_REMOTE_METADATA_NAME = "TEST_CACHE_REMOTE_METADATA_NAME"
	TEST_ENCODED_CATALOG_MANIFEST   = "TEST__ENCODED_CATALOG_MANIFEST"
)

func TestIsCached() error {
	ctx := basecontext.NewRootBaseContext()
	ctx.LogInfof("Testing if catalog is cached functionality")
	catalogSvc := catalog.NewManifestService(ctx)
	cfg := config.Get()
	var apiManifest models.CatalogManifest
	encodedManifest := cfg.GetKey(TEST_ENCODED_CATALOG_MANIFEST)
	decodedManifest, err := base64.StdEncoding.DecodeString(encodedManifest)
	if err != nil {
		ctx.LogErrorf("Error decoding catalog manifest: %v", err)
		return err
	}

	err = json.Unmarshal(decodedManifest, &apiManifest)
	if err != nil {
		ctx.LogErrorf("Error unmarshalling catalog manifest: %v", err)
		return err
	}
	m := mappers.ApiCatalogManifestToCatalogManifest(apiManifest)

	rss, err := catalogSvc.GetProviderFromConnection(m.Provider.String())
	if err != nil {
		ctx.LogErrorf("Error getting provider from connection: %v", err)
		return err
	}

	cr := cacheservice.NewCacheRequest(ctx, &m, rss)

	cacheSvc, err := cacheservice.NewCacheService(ctx)
	if err != nil {
		ctx.LogErrorf("Error creating cache request: %v", err)
		return err
	}
	cacheSvc.WithRequest(cr)

	if cacheSvc.IsCached() {
		ctx.LogInfof("Catalog is cached")
	} else {
		ctx.LogInfof("Catalog is not cached, caching it")
		if err := cacheSvc.Cache(); err != nil {
			ctx.LogErrorf("Error caching catalog: %v", err)
			return err
		}
	}

	return nil
}
