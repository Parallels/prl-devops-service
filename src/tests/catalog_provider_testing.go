package tests

import (
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/tester"
)

func TestCatalogProviders(ctx basecontext.ApiContext) error {
	if os.Getenv("ARTIFACTORY_TEST_CONNECTION") != "" {
		ctx.LogInfo("Testing connection to Artifactory")
		test := tester.NewTestProvider(ctx, os.Getenv("ARTIFACTORY_TEST_CONNECTION"))
		err := test.Test()
		if err != nil {
			ctx.LogError(err.Error())
			return err
		} else {
			ctx.LogInfo("Connection to Artifactory successful")
		}
	}

	if os.Getenv("AZURE_SA_TEST_CONNECTION") != "" {
		ctx.LogInfo("Testing connection to Azure Storage Account")
		test := tester.NewTestProvider(ctx, os.Getenv("AZURE_SA_TEST_CONNECTION"))
		err := test.Test()
		if err != nil {
			ctx.LogError(err.Error())
			return err
		} else {
			ctx.LogInfo("Connection to Azure Storage Account successful")
		}
	}

	if os.Getenv("AWS_S3_TEST_CONNECTION") != "" {
		ctx.LogInfo("Testing connection to AWS S3")
		test := tester.NewTestProvider(ctx, os.Getenv("AWS_S3_TEST_CONNECTION"))
		err := test.Test()
		if err != nil {
			ctx.LogError(err.Error())
			return err
		} else {
			ctx.LogInfo("Connection to AWS S3 successful")
		}
	}

	return nil
}
