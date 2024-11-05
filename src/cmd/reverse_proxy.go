package cmd

import (
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/reverse_proxy"
	"github.com/cjlapao/common-go/helper"
)

func processReverseProxy(ctx basecontext.ApiContext, cmd string) {
	// processing the command help
	if helper.GetFlagSwitch(constants.HELP_FLAG, false) || helper.GetCommandAt(1) == "help" {
		processHelp(constants.REVERSE_PROXY_COMMAND)
		os.Exit(0)
	}
	_ = os.Setenv(constants.SOURCE_ENV_VAR, constants.REVERSE_PROXY_COMMAND)

	// Loading configuration
	cfg := config.New(ctx)
	cfg.Load()
	processTelemetry(cmd)

	service := reverse_proxy.New(ctx)
	if service == nil {
		ctx.LogErrorf("Error creating reverse proxy service")
		os.Exit(1)
	}
	if err := service.Start(); err != nil {
		ctx.LogErrorf(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func processReverseProxyHelp() {
	fmt.Println("Starts a Reverse Proxy server for the API service.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %v %v <options>\n", constants.ExecutableName, constants.REVERSE_PROXY_COMMAND)
	fmt.Println()
	fmt.Println("Example:")
	fmt.Printf("  %v %v\n", constants.ExecutableName, constants.REVERSE_PROXY_COMMAND)
	fmt.Println()
}
