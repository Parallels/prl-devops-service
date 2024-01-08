package cmd

import (
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/security"
	"github.com/cjlapao/common-go/helper"
)

func processGenerateSecurityKey(ctx basecontext.ApiContext) {
	ctx.LogInfo("Generating security key")
	filename := "private.key"

	if helper.GetFlagValue(constants.FILE_FLAG, "") != "" {
		filename = helper.GetFlagValue(constants.FILE_FLAG, "")
	}

	err := security.GenPrivateRsaKey(filename)
	if err != nil {
		panic(err)
	}

	os.Exit(0)
}
