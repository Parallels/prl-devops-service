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
		processApi(ctx)
	case constants.GENERATE_SECURITY_KEY_COMMAND:
		processGenerateSecurityKey(ctx)
	case constants.INSTALL_SERVICE_COMMAND:
		processInstall(ctx)
	case constants.UNINSTALL_SERVICE_COMMAND:
		processUninstall(ctx)
	case constants.TEST_COMMAND:
		processTestProviders(ctx)
	case constants.VERSION_COMMAND:
		processVersion()
	case constants.HELP_COMMAND:
		processHelp("")
	case constants.CATALOG_COMMAND:
		processCatalog(ctx)
	case constants.UPDATE_ROOT_PASSWORD_COMMAND:
		processRootPassword(ctx)
	default:
		processApi(ctx)
	}

	os.Exit(0)
}
