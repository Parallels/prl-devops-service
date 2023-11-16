package main

import (
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/install"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/security"
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/Parallels/pd-api-service/startup"
	"github.com/Parallels/pd-api-service/tests"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/version"
)

var ver = "0.2.0"
var versionSvc = version.Get()

//	@title			Parallels Desktop API
//	@version		1.0
//	@description	Parallels Desktop API Service
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Parallels Desktop API Support
//	@contact.url	https://forum.parallels.com/
//	@contact.email	carlos.lapao@parallels.com

//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@BasePath					/api
//	@securityDefinitions.apikey	ApiKeyAuth
//	@description				Type the api key in the input below.
//	@in							header
//	@name						X-Api-Key

// @securityDefinitions.apikey	BearerAuth
// @description				Type "Bearer" followed by a space and JWT token.
// @in							header
// @name						Authorization
func main() {
	versionSvc.Author = "Carlos Lapao"
	versionSvc.Name = "Parallels Desktop API Service"
	versionSvc.License = "MIT"
	// Reading the version from a string
	strVer, err := version.FromString(ver)
	if err == nil {
		versionSvc.Major = strVer.Major
		versionSvc.Minor = strVer.Minor
		versionSvc.Build = strVer.Build
		versionSvc.Rev = strVer.Rev
	}

	if helper.GetFlagSwitch("version", false) {
		println(versionSvc.String())
		os.Exit(0)
	}

	versionSvc.PrintAnsiHeader()
	ctx := basecontext.NewRootBaseContext()
	cfg := config.NewConfig()

	// Checking if we just want to test
	if helper.GetFlagSwitch(constants.TEST_FLAG, false) {
		if helper.GetFlagSwitch(constants.TEST_CATALOG_PROVIDERS_FLAG, false) {
			if err := tests.TestCatalogProviders(ctx); err != nil {
				ctx.LogError(err.Error())
				os.Exit(1)
			}
		}

		os.Exit(0)
	}

	if helper.GetFlagSwitch(constants.GENERATE_SECURITY_KEY_FLAG, false) {
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

	if helper.GetFlagSwitch(constants.INSTALL_SERVICE_FLAG, false) {
		if err := install.InstallService(ctx); err != nil {
			ctx.LogError(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if helper.GetFlagSwitch(constants.UNINSTALL_SERVICE_FLAG, false) {
		if err := install.UninstallService(ctx); err != nil {
			ctx.LogError(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if cfg.GetSecurityKey() == "" {
		common.Logger.Warn("No security key found, database will be unencrypted")
	}

	startup.Start()

	if helper.GetFlagSwitch(constants.UPDATE_ROOT_PASSWORD_FLAG, false) {
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

	currentUser, err := serviceprovider.Get().System.GetCurrentUser(ctx)
	if err != nil {
		panic(err)
	}
	os.Setenv(constants.CURRENT_USER_ENV_VAR, currentUser)
	currentUserEnv := os.Getenv(constants.CURRENT_USER_ENV_VAR)
	if currentUserEnv != "" {
		ctx.LogInfo("Running with user %s", currentUser)
	}

	// updating the root password if the environment variable is set
	if os.Getenv(constants.ROOT_PASSWORD_ENV_VAR) != "" {
		rootPassword := helpers.Sha256Hash(os.Getenv(constants.ROOT_PASSWORD_ENV_VAR))
		db := serviceprovider.Get().JsonDatabase
		db.Connect(ctx)
		rootUser, _ := db.GetUser(ctx, "root")
		if rootUser != nil {
			if rootUser.Password != rootPassword {
				ctx.LogInfo("Updating root password")
				db.UpdateRootPassword(ctx, os.Getenv(constants.ROOT_PASSWORD_ENV_VAR))
			}
		}
	}

	// Serve the API
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
