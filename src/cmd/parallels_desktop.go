package cmd

import (
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/security"
	"github.com/Parallels/prl-devops-service/serviceprovider/parallelsdesktop"
	"github.com/cjlapao/common-go/helper"
)

func processParallelsDesktop(ctx basecontext.ApiContext) {
	command := helper.GetCommandAt(0)
	// processing the command help
	if helper.GetFlagSwitch(constants.HELP_FLAG, false) || helper.GetCommandAt(1) == "help" {
		processHelp(command)
		os.Exit(0)
	}
	os.Setenv(constants.SOURCE_ENV_VAR, "parallels_desktop")
	ctx.ToggleLogTimestamps(false)
	pdSvc := parallelsdesktop.New(ctx)

	switch command {
	case constants.START_COMMAND:
		machineId := helper.GetCommandAt(1)
		if machineId == "" {
			ctx.LogErrorf("No machine id or name provided")
			os.Exit(1)
		}
		err := pdSvc.StartVm(ctx, machineId)
		if err != nil {
			ctx.LogErrorf("Error starting vm: %v", err)
			os.Exit(1)
		}
		ctx.LogInfof("VM %s started", machineId)
	case constants.STOP_COMMAND:
		machineId := helper.GetCommandAt(1)
		if machineId == "" {
			ctx.LogErrorf("No machine id or name provided")
			os.Exit(1)
		}
		err := pdSvc.StopVm(ctx, machineId)
		if err != nil {
			ctx.LogErrorf("Error stopping vm: %v", err)
			os.Exit(1)
		}
		ctx.LogInfof("VM %s stopped", machineId)
	case constants.EXEC_COMMAND:
		machineId := helper.GetCommandAt(1)
		if machineId == "" {
			ctx.LogErrorf("No machine id or name provided")
			os.Exit(1)
		}
		command := helper.GetCommandAt(2)
		if command == "" {
			ctx.LogErrorf("No command provided")
			os.Exit(1)
		}
		ctx.LogInfof("Executing command on VM %s", machineId)
		request := models.VirtualMachineExecuteCommandRequest{
			Command: command,
		}
		response, err := pdSvc.ExecuteCommandOnVm(ctx, machineId, &request)
		if err != nil {
			ctx.LogErrorf("Error executing on vm: %v", err)
			os.Exit(1)
		}
		if response != nil {
			if response.ExitCode != 0 {
				ctx.LogErrorf("Command failed with exit code %v", response.ExitCode)
				ctx.LogErrorf(response.Stderr)
			}
			ctx.LogInfof("Command executed successfully with exit code %v", response.ExitCode)
			ctx.LogInfof(response.Stdout)
		}
		ctx.LogInfof("Command executed successfully on VM %s", machineId)
	case constants.CLONE_COMMAND:
		machineId := helper.GetCommandAt(1)
		cloneName := helper.GetCommandAt(2)
		if machineId == "" {
			ctx.LogErrorf("No machine id or name provided")
			os.Exit(1)
		}
		if cloneName == "" {
			name, err := security.GenerateCryptoRandomString(20)
			if err != nil {
				ctx.LogErrorf("Error generating random clone name: %v", err)
				os.Exit(1)
			}
			cloneName = name
		}

		err := pdSvc.CloneVm(ctx, machineId, cloneName)
		if err != nil {
			ctx.LogErrorf("Error cloning vm: %v", err)
			os.Exit(1)
		}
		ctx.LogInfof("VM %s Cloned to %s", machineId, cloneName)

	default:
		processHelp(command)
	}

	os.Exit(0)
}

func processParallelsDesktopHelp(command string) {
	switch command {
	case constants.CLONE_COMMAND:
		processParallelsDesktopCloneCommand()
	case constants.START_COMMAND:
		processParallelsDesktopStartCommand()
	default:
		fmt.Println("Usage:")
		fmt.Printf("  %v %v <machine_id>\n", constants.ExecutableName, command)
	}
}

func processParallelsDesktopCloneCommand() {
	fmt.Println("Usage:")
	fmt.Printf("  %v clone <machine_id> <clone_name>\n", constants.ExecutableName)
	fmt.Println()
}

func processParallelsDesktopStartCommand() {
	fmt.Println("Usage:")
	fmt.Printf("  %v start <machine_id>\n", constants.ExecutableName)
	fmt.Println()
}
