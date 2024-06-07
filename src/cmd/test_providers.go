package cmd

import (
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/tests"
	"github.com/cjlapao/common-go/helper"
)

func processTestProviders(ctx basecontext.ApiContext) {
	subcommand := helper.GetCommandAt(1)
	_ = os.Setenv(constants.SOURCE_ENV_VAR, "test")
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
	fmt.Println("Usage: prl-devops-service test <command>")
	fmt.Println("  catalog-providers\t\t\tRun a test on all the catalog providers")
}
