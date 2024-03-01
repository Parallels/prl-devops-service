package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/install"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper"
)

type InstallServiceResult struct {
	Success bool   `json:"success,omitempty" yaml:"success,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

func processInstall(ctx basecontext.ApiContext) {
	subcommand := helper.GetCommandAt(1)
	// processing the command help
	if helper.GetFlagSwitch(constants.HELP_FLAG, false) || helper.GetCommandAt(1) == "help" {
		processHelp(constants.INSTALL_SERVICE_COMMAND)
		os.Exit(0)
	}
	ctx.ToggleLogTimestamps(false)
	ctx.DisableLog()
	serviceprovider.InitServices(ctx)
	providerSvc := serviceprovider.Get()
	ctx.EnableLog()
	if providerSvc == nil {
		ctx.LogErrorf("There was an error during initialization, exiting...")
		os.Exit(1)
	}
	currentUser, err := providerSvc.System.GetCurrentUser(ctx)
	if err != nil {
		ctx.LogErrorf("There was anm error getting the current user, err: %v", err.Error())
		os.Exit(1)
	}

	versionFlag := helper.GetFlagValue("version", "")
	userFlag := helper.GetFlagValue("user", currentUser)
	flags := make(map[string]string)

	switch subcommand {
	case "service":
		filePath := helper.GetFlagValue(constants.FILE_FLAG, "")
		ctx.ToggleLogTimestamps(false)
		if filePath != "" {
			if err := install.InstallService(ctx, filePath); err != nil {
				ctx.LogErrorf(err.Error())
				os.Exit(1)
			}
		} else {
			if err := install.InstallService(ctx, ""); err != nil {
				ctx.LogErrorf(err.Error())
				os.Exit(1)
			} else {
				cmdResult := InstallServiceResult{Success: true, Message: "Service installed successfully"}
				_ = json.NewEncoder(os.Stdout).Encode(cmdResult)
			}
		}
	case "brew":
		result := providerSvc.InstallTool(userFlag, "brew", versionFlag, flags)
		if !result.Result {
			ctx.LogErrorf("There was an error installing brew: %v", result.Message)
			os.Exit(1)
		} else {
			cmdResult := InstallServiceResult{Success: true, Message: "brew installed successfully"}
			if result.Message != "" {
				cmdResult.Message = result.Message
			}
			_ = json.NewEncoder(os.Stdout).Encode(cmdResult)
		}
	case "parallels-desktop":
		result := providerSvc.InstallTool(userFlag, "parallels-desktop", versionFlag, flags)
		if !result.Result {
			ctx.LogErrorf("There was an error installing parallels-desktop: %v", result.Message)
			os.Exit(1)
		} else {
			cmdResult := InstallServiceResult{Success: true, Message: "parallels-desktop installed successfully"}
			if result.Message != "" {
				cmdResult.Message = result.Message
			}
			_ = json.NewEncoder(os.Stdout).Encode(cmdResult)
		}
	case "git":
		result := providerSvc.InstallTool(userFlag, "git", versionFlag, flags)
		if !result.Result {
			ctx.LogErrorf("There was an error installing git: %v", result.Message)
			os.Exit(1)
		} else {
			cmdResult := InstallServiceResult{Success: true, Message: "git installed successfully"}
			if result.Message != "" {
				cmdResult.Message = result.Message
			}
			_ = json.NewEncoder(os.Stdout).Encode(cmdResult)
		}
	case "packer":
		result := providerSvc.InstallTool(userFlag, "packer", versionFlag, flags)
		if !result.Result {
			ctx.LogErrorf("There was an error installing packer: %v", result.Message)
			os.Exit(1)
		} else {
			cmdResult := InstallServiceResult{Success: true, Message: "packer installed successfully"}
			if result.Message != "" {
				cmdResult.Message = result.Message
			}
			json.NewEncoder(os.Stdout).Encode(cmdResult)
		}
	case "vagrant":
		result := providerSvc.InstallTool(userFlag, "vagrant", versionFlag, flags)
		if !result.Result {
			ctx.LogErrorf("There was an error installing vagrant: %v", result.Message)
			os.Exit(1)
		} else {
			cmdResult := InstallServiceResult{Success: true, Message: "vagrant installed successfully"}
			if result.Message != "" {
				cmdResult.Message = result.Message
			}
			_ = json.NewEncoder(os.Stdout).Encode(cmdResult)
		}
	default:
		processHelp(constants.INSTALL_SERVICE_COMMAND)
	}

	os.Exit(0)
}

func processInstallHelp() {
	fmt.Println("Usage:")
	fmt.Printf("  %v %v <command> <flags>\n", constants.ExecutableName, constants.INSTALL_SERVICE_COMMAND)
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  service\t\tInstalls the RestAPI service as a daemon that will start on reboot")
	fmt.Println("  brew\t\t\tInstalls the brew tool in the system")
	fmt.Println("  git\t\t\tInstalls the git tool in the system")
	fmt.Println("  parallels-desktop \tInstalls the parallels-desktop in the system")
	fmt.Println("  packer \t\tInstalls Hashicorp packer in the system")
	fmt.Println("  vagrant \t\tInstalls Hashicorp vagrant in the system")
	fmt.Println()
	fmt.Println("flags:")
	fmt.Println("  user\t\twhat user would be used to install the service, by default the current user is defined")
	fmt.Println("  version\t\tRequest a specific version to be installed")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Printf("  %v %v git --version=latest", constants.ExecutableName, constants.INSTALL_SERVICE_COMMAND)
	fmt.Println()
}
