package cmd

import (
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/tests"
	"github.com/cjlapao/common-go/helper"
)

func processTestProviders() {
	ctx := basecontext.NewRootBaseContext()
	// Checking if we just want to test
	if helper.GetFlagSwitch(constants.TEST_CATALOG_PROVIDERS_FLAG, false) {
		if err := tests.TestCatalogProviders(ctx); err != nil {
			ctx.LogError(err.Error())
			os.Exit(1)
		}
	}

	os.Exit(0)
}
