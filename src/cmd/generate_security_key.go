package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/security"
	"github.com/cjlapao/common-go/helper"
)

func processGenerateSecurityKey(ctx basecontext.ApiContext) {
	// processing the command help
	if helper.GetFlagSwitch(constants.HELP_FLAG, false) || helper.GetCommandAt(1) == "help" {
		processHelp(constants.GENERATE_SECURITY_KEY_COMMAND)
		os.Exit(0)
	}

	filename := "private.key"
	ctx.ToggleLogTimestamps(false)
	cfg := config.Get()

	if cfg.GetKey(constants.FILE_FLAG) != "" {
		filename = helper.GetFlagValue(constants.FILE_FLAG, "")
	} else {
		ctx.LogInfof("No output file specified, using default: %s", filename)
	}
	keySize := 2048
	if cfg.GetKey(constants.RSA_KEY_SIZE) != "" {
		size, err := strconv.Atoi(cfg.GetKey(constants.RSA_KEY_SIZE))
		if err != nil {
			ctx.LogErrorf("Error parsing size flag: %s", err.Error())
		} else {
			keySize = size
		}
	} else {
		ctx.LogInfof("No key size specified, using default: %v", keySize)
	}

	ctx.LogInfof("Generating security key, with size %v", keySize)

	err := security.GenPrivateRsaKey(filename, keySize)
	if err != nil {
		panic(err)
	}

	os.Exit(0)
}

func processGenerateSecurityKeyHelp() {
	fmt.Println("Generates an RSA private key for the API service. you can use this to encrypt the database at rest.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %v %v <options>\n", constants.ExecutableName, constants.GENERATE_SECURITY_KEY_COMMAND)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Printf("  %s\t\t output file where the key will be saved to\n", constants.FILE_FLAG)
	fmt.Printf("  %s\t size of the key to be generated\n", constants.RSA_KEY_SIZE)
	fmt.Println()
	fmt.Println("Example:")
	fmt.Printf("  %v %v --%s=private.key --%s=4096\n", constants.ExecutableName, constants.GENERATE_SECURITY_KEY_COMMAND, constants.FILE_FLAG, constants.RSA_KEY_SIZE)
	fmt.Println()
}
