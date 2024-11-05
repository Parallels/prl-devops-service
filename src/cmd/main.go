package cmd

import (
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/telemetry"
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

	cfg := config.Get()
	cfg.SetRunningCommand(command)

	switch command {
	case constants.API_COMMAND:
		processApi(ctx, command)
	case constants.GENERATE_SECURITY_KEY_COMMAND:
		processGenerateSecurityKey(ctx, command)
	case constants.INSTALL_SERVICE_COMMAND:
		processInstall(ctx, command)
	case constants.UNINSTALL_SERVICE_COMMAND:
		processUninstall(ctx, command)
	case constants.TEST_COMMAND:
		processTestProviders(ctx, command)
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
		processCatalog(ctx, command, subcommand, filepath)
	case constants.CATALOG_PUSH_COMMAND:
		filepath := helper.GetCommandAt(1)
		if !strings.HasSuffix(filepath, ".pdfile") {
			filepath = ""
		}
		processCatalog(ctx, command, "push", filepath)
	case constants.CATALOG_PULL_COMMAND:
		filepath := helper.GetCommandAt(1)
		if !strings.HasSuffix(filepath, ".pdfile") {
			filepath = ""
		}
		processCatalog(ctx, command, "pull", filepath)
	case constants.UPDATE_ROOT_PASSWORD_COMMAND:
		processRootPassword(ctx, command)
	case constants.REVERSE_PROXY_COMMAND:
		processReverseProxy(ctx, command)
	case constants.START_COMMAND,
		constants.STOP_COMMAND,
		constants.CLONE_COMMAND,
		constants.DELETE_COMMAND,
		constants.EXEC_COMMAND:
		processParallelsDesktop(ctx, command)
	default:
		if helper.GetFlagSwitch("help", false) {
			processHelp("")
			os.Exit(0)
		}
		processApi(ctx, command)
	}

	os.Exit(0)
}

func processTelemetry(command string) {
	if telemetry.Get() != nil {
		cmd := command
		if cmd == "" {
			cmd = "api"
		}
		telemetry.SendStartEvent(cmd)
	}
}
