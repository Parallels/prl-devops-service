package cmd

import (
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/install"
	"github.com/cjlapao/common-go/helper"
)

func processUninstall(ctx basecontext.ApiContext) {
	removeDatabase := false
	if helper.GetFlagSwitch("full", false) {
		removeDatabase = true
	}
	if err := install.UninstallService(ctx, removeDatabase); err != nil {
		ctx.LogErrorf(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
