package cmd

import (
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/tests"
	"github.com/cjlapao/common-go/helper"
)

func processTestProviders(ctx basecontext.ApiContext, cmd string) {
	subcommand := helper.GetCommandAt(1)
	processTelemetry(cmd)
	_ = os.Setenv(constants.SOURCE_ENV_VAR, "test")
	switch subcommand {
	case constants.TEST_CATALOG_PROVIDERS_FLAG:
		if err := tests.TestCatalogProviders(ctx); err != nil {
			ctx.LogErrorf(err.Error())
			os.Exit(1)
		}
	case "unzip":
		rootctx := basecontext.NewBaseContext()
		filename := helper.GetFlagValue("zip-file", "")
		destination := helper.GetFlagValue("destination", "")
		if destination == "" {
			destination = "/tmp"
		}
		cSrv := catalog.NewManifestService(rootctx)
		cSrv.Unzip(rootctx, filename, destination)
	case "push-file":
		rootctx := basecontext.NewBaseContext()
		filename := helper.GetFlagValue("file_path", "")
		targetPath := helper.GetFlagValue("target_path", "")
		targetFilename := helper.GetFlagValue("target_filename", "")
		if err := tests.TestCatalogProvidersPushFile(rootctx, filename, targetPath, targetFilename); err != nil {
			ctx.LogErrorf(err.Error())
			os.Exit(1)
		}
	case "catalog-cache":
		cacheSubcommand := helper.GetCommandAt(2)
		switch cacheSubcommand {
		case "is-cached":
			if err := tests.TestIsCached(); err != nil {
				ctx.LogErrorf(err.Error())
				os.Exit(1)
			}
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
