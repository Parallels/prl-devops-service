package cmd

import (
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/install"
)

func processUninstall(ctx basecontext.ApiContext) {
	if err := install.UninstallService(ctx); err != nil {
		ctx.LogError(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
