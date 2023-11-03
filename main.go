package main

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/config"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/security"
	"Parallels/pd-api-service/serviceprovider"
	"Parallels/pd-api-service/startup"
	"os"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/version"
)

var versionSvc = version.Get()

func main() {
	versionSvc.Author = "Carlos Lapao"
	versionSvc.Name = "Parallels Desktop API Service"
	versionSvc.License = "MIT"
	versionSvc.Major = 0
	versionSvc.Minor = 1
	versionSvc.Build = 32

	if helper.GetFlagSwitch("version", false) {
		println(versionSvc.String())
		os.Exit(0)
	}

	versionSvc.PrintAnsiHeader()
	ctx := basecontext.NewRootBaseContext()
	cfg := config.NewConfig()

	if helper.GetFlagSwitch(constants.GENERATE_SECURITY_KEY, false) {
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

	if cfg.GetSecurityKey() == "" {
		common.Logger.Warn("No security key found, database will be unencrypted")
	}

	startup.Start()

	if helper.GetFlagSwitch(constants.UPDATE_ROOT_PASSWORD, false) {
		ctx.LogInfo("Updating root password")
		rootPassword := helper.GetFlagValue("password", "")
		if rootPassword != "" {
			db := serviceprovider.Get().JsonDatabase
			ctx.LogInfo("Database connection found, updating password")
			db.Connect(ctx)
			if db != nil {
				err := db.UpdateRootPassword(ctx, rootPassword)
				if err != nil {
					panic(err)
				}
				db.Disconnect(ctx)
			} else {
				panic(data.ErrDatabaseNotConnected)
			}
		} else {
			panic("No password provided")
		}
		ctx.LogInfo("Root password updated")

		os.Exit(0)
	}

	// Serve the API
	currentUser, err := serviceprovider.Get().System.GetCurrentUser(ctx)
	if err != nil {
		panic(err)
	}
	os.Setenv(constants.CURRENT_USER_ENV_VAR, currentUser)
	ctx.LogInfo("Running with user %s", os.Getenv(constants.CURRENT_USER_ENV_VAR))

	for {
		listener := startup.InitApi()
		restartChannel := restapi.GetRestartChannel()
		listener.Start(versionSvc.Name, versionSvc.String())

		needsRestart := <-restartChannel
		if !needsRestart {
			break
		}
		startup.Start()
	}
}
