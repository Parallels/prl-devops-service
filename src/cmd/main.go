package cmd

import (
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/cjlapao/common-go/helper"
)

func Process() {
	command := helper.GetCommandAt(0)
	ctx := basecontext.NewRootBaseContext()

	switch command {
	case constants.API_COMMAND:
		processApi()
	case constants.GENERATE_SECURITY_KEY_COMMAND:
		processGenerateSecurityKey()
	case constants.INSTALL_SERVICE_COMMAND:
		processInstall()
	case constants.UNINSTALL_SERVICE_COMMAND:
		processUninstall()
	case constants.TEST_COMMAND:
		processTestProviders()
	case constants.VERSION_COMMAND:
		processVersion()
	case constants.HELP_COMMAND:
		processHelp("")
	case constants.CATALOG_COMMAND:
		processCatalog(ctx)
	default:
		processApi()
	}
	os.Exit(0)
}
