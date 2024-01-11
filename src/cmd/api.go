package cmd

import (
	"fmt"
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/orchestrator"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/security/password"
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/Parallels/pd-api-service/startup"
	"github.com/cjlapao/common-go/helper"
)

func processApi(ctx basecontext.ApiContext) {
	// processing the command help
	if helper.GetFlagSwitch(constants.HELP_FLAG, false) || helper.GetCommandAt(1) == "help" {
		processHelp(constants.API_COMMAND)
		os.Exit(0)
	}

	versionSvc.PrintAnsiHeader()
	cfg := config.NewConfig()

	if cfg.GetSecurityKey() == "" {
		common.Logger.Warn("No security key found, database will be unencrypted")
	}
	startup.Start()
	startup.Init()

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
		db := serviceprovider.Get().JsonDatabase
		_ = db.Connect(ctx)
		rootUser, _ := db.GetUser(ctx, "root")
		rootPassword := os.Getenv(constants.ROOT_PASSWORD_ENV_VAR)
		if rootUser != nil {
			passwdSvc := password.Get()
			if err := passwdSvc.Compare(rootPassword, rootUser.ID, rootUser.Password); err != nil {
				ctx.LogInfo("Updating root password")
				if err := db.UpdateRootPassword(ctx, os.Getenv(constants.ROOT_PASSWORD_ENV_VAR)); err != nil {
					panic(err)
				}
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

	if cfg.IsOrchestrator() {
		ctx := basecontext.NewRootBaseContext()
		orchestratorBackgroundService := orchestrator.NewOrchestratorService(ctx)
		orchestratorBackgroundService.Stop()
	}
}

func processApiHelp() {
	fmt.Println("Usage: pd-api-service api [OPTIONS]")
}
