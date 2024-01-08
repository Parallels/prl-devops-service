package cmd

import (
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/install"
)

func processUninstall() {
	ctx := basecontext.NewRootBaseContext()

	if err := install.UninstallService(ctx); err != nil {
		ctx.LogError(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
