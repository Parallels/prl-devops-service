package cmd

import (
	"os"
	"strconv"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/security"
	"github.com/cjlapao/common-go/helper"
)

func processGenerateSecurityKey(ctx basecontext.ApiContext) {
	filename := "private.key"

	if helper.GetFlagValue(constants.FILE_FLAG, "") != "" {
		filename = helper.GetFlagValue(constants.FILE_FLAG, "")
	}
	keySize := 2048
	if helper.GetFlagValue(constants.SIZE_FLAG, "") != "" {
		size, err := strconv.Atoi(helper.GetFlagValue(constants.SIZE_FLAG, "0"))
		if err != nil {
			ctx.LogError("Error parsing size flag: %s", err.Error())
		} else {
			keySize = size
		}
	}

	ctx.LogInfo("Generating security key, with size %v", keySize)

	err := security.GenPrivateRsaKey(filename, keySize)
	if err != nil {
		panic(err)
	}

	os.Exit(0)
}
