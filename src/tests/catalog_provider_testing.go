package tests

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/tester"
	"github.com/Parallels/pd-api-service/config"
)

func TestCatalogProviders(ctx basecontext.ApiContext) error {
	cfg := config.Get()
	if cfg.GetKey("ARTIFACTORY_TEST_CONNECTION") != "" {
		ctx.LogInfof("Testing connection to Artifactory")
		test := tester.NewTestProvider(ctx, cfg.GetKey("ARTIFACTORY_TEST_CONNECTION"))
		err := test.Test()
		if err != nil {
			ctx.LogErrorf(err.Error())
			return err
		} else {
			ctx.LogInfof("Connection to Artifactory successful")
		}
	}

	if cfg.GetKey("AZURE_SA_TEST_CONNECTION") != "" {
		ctx.LogInfof("Testing %v", cfg.GetKey("AZURE_SA_TEST_CONNECTION"))
		ctx.LogInfof("Testing connection to Azure Storage Account")
		test := tester.NewTestProvider(ctx, cfg.GetKey("AZURE_SA_TEST_CONNECTION"))
		err := test.Test()
		if err != nil {
			ctx.LogErrorf(err.Error())
			return err
		} else {
			ctx.LogInfof("Connection to Azure Storage Account successful")
		}
	}

	if cfg.GetKey("AWS_S3_TEST_CONNECTION") != "" {
		ctx.LogInfof("Testing %v", cfg.GetKey("AWS_S3_TEST_CONNECTION"))
		ctx.LogInfof("Testing connection to AWS S3")
		test := tester.NewTestProvider(ctx, cfg.GetKey("AWS_S3_TEST_CONNECTION"))
		err := test.Test()
		if err != nil {
			ctx.LogErrorf(err.Error())
			return err
		} else {
			ctx.LogInfof("Connection to AWS S3 successful")
		}
	}

	return nil
}
