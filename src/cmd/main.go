package cmd

import (
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/cjlapao/common-go/helper"
)

func Process() {
	command := helper.GetCommandAt(0)
	ctx := basecontext.NewRootBaseContext()
	// backwards compatibility with the --version flag
	if helper.GetFlagSwitch("version", false) {
		processVersion()
		os.Exit(0)
	}

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
	case constants.REVERSE_PROXY_COMMAND:
		processReverseProxy(ctx)
	default:
		if helper.GetFlagSwitch("help", false) {
			processHelp("")
			os.Exit(0)
		}
		processApi(ctx)
	}

	os.Exit(0)
}
