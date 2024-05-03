package cmd

import (
	"os"
	"strings"

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
		subcommand := helper.GetCommandAt(1)
		filepath := helper.GetCommandAt(2)
		if !strings.HasSuffix(filepath, ".pdfile") {
			filepath = ""
		}
		processCatalog(ctx, subcommand, filepath)
	case constants.CATALOG_PUSH_COMMAND:
		filepath := helper.GetCommandAt(1)
		if !strings.HasSuffix(filepath, ".pdfile") {
			filepath = ""
		}
		processCatalog(ctx, "push", filepath)
	case constants.CATALOG_PULL_COMMAND:
		filepath := helper.GetCommandAt(1)
		if !strings.HasSuffix(filepath, ".pdfile") {
			filepath = ""
		}
		processCatalog(ctx, "pull", filepath)
	case constants.UPDATE_ROOT_PASSWORD_COMMAND:
		processRootPassword(ctx)
	case constants.REVERSE_PROXY_COMMAND:
		processReverseProxy(ctx)
	case constants.START_COMMAND,
		constants.STOP_COMMAND,
		constants.CLONE_COMMAND,
		constants.DELETE_COMMAND,
		constants.EXEC_COMMAND:
		processParallelsDesktop(ctx)
	default:
		if helper.GetFlagSwitch("help", false) {
			processHelp("")
			os.Exit(0)
		}
		processApi(ctx)
	}

	os.Exit(0)
}
