package cmd

import (
	"os"
	"strconv"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/security"
	"github.com/cjlapao/common-go/helper"
)

func processGenerateSecurityKey(ctx basecontext.ApiContext) {
	filename := "private.key"
	cfg := config.Get()

	if cfg.GetKey(constants.FILE_FLAG) != "" {
		filename = helper.GetFlagValue(constants.FILE_FLAG, "")
	}
	keySize := 2048
	if cfg.GetKey(constants.RSA_KEY_SIZE) != "" {
		size, err := strconv.Atoi(cfg.GetKey(constants.RSA_KEY_SIZE))
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
