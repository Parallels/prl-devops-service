package cmd

import (
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/install"
	"github.com/cjlapao/common-go/helper"
)

func processInstall(ctx basecontext.ApiContext) {
	filePath := helper.GetFlagValue(constants.FILE_FLAG, "")
	if filePath != "" {
		if err := install.InstallService(ctx, filePath); err != nil {
			ctx.LogErrorf(err.Error())
			os.Exit(1)
		}
	} else {
		if err := install.InstallService(ctx, ""); err != nil {
			ctx.LogErrorf(err.Error())
			os.Exit(1)
		}
	}
	os.Exit(0)
}
