package cmd

import (
	"fmt"
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/tests"
	"github.com/cjlapao/common-go/helper"
)

func processTestProviders(ctx basecontext.ApiContext) {
	subcommand := helper.GetCommandAt(1)
	switch subcommand {
	case constants.TEST_CATALOG_PROVIDERS_FLAG:
		if err := tests.TestCatalogProviders(ctx); err != nil {
			ctx.LogErrorf(err.Error())
			os.Exit(1)
		}
	default:
		processHelp(constants.TEST_COMMAND)

	}

	os.Exit(0)
}

func processTestHelp() {
	fmt.Println("Usage: pd-api-service test <command>")
	fmt.Println("  catalog-providers\t\t\tRun a test on all the catalog providers")
}
