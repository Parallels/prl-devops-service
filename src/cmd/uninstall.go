package cmd

import (
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/install"
	"github.com/cjlapao/common-go/helper"
)

func processUninstall(ctx basecontext.ApiContext) {
	ctx.ToggleLogTimestamps(false)
	removeDatabase := helper.GetFlagSwitch("full", false)
	_ = os.Setenv(constants.SOURCE_ENV_VAR, "uninstall")

	if err := install.UninstallService(ctx, removeDatabase); err != nil {
		ctx.LogErrorf(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
