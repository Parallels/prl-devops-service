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
	startup.Init(ctx)

	startup.Start(ctx)
	cfg := config.Get()

	if cfg.EncryptionPrivateKey() == "" {
		common.Logger.Warn("No security key found, database will be unencrypted")
	}

	currentUser, err := serviceprovider.Get().System.GetCurrentUser(ctx)
	if err != nil {
		panic(err)
	}

	if err := os.Setenv(constants.CURRENT_USER_ENV_VAR, currentUser); err != nil {
		panic(err)
	}

	currentUserEnv := cfg.GetKey(constants.CURRENT_USER_ENV_VAR)
	if currentUserEnv != "" {
		ctx.LogInfof("Running with user %s", currentUser)
	}

	// updating the root password if the environment variable is set
	if cfg.GetKey(constants.ROOT_PASSWORD_ENV_VAR) != "" {
		db := serviceprovider.Get().JsonDatabase
		_ = db.Connect(ctx)
		rootUser, _ := db.GetUser(ctx, "root")
		rootPassword := cfg.GetKey(constants.ROOT_PASSWORD_ENV_VAR)
		if rootUser != nil {
			passwdSvc := password.Get()
			if err := passwdSvc.Compare(rootPassword, rootUser.ID, rootUser.Password); err != nil {
				ctx.LogInfof("Updating root password")
				if err := db.UpdateRootPassword(ctx, cfg.GetKey(constants.ROOT_PASSWORD_ENV_VAR)); err != nil {
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
		startup.Start(ctx)
	}

	if cfg.IsOrchestrator() {
		ctx := basecontext.NewRootBaseContext()
		orchestratorBackgroundService := orchestrator.NewOrchestratorService(ctx)
		orchestratorBackgroundService.Stop()
	}
}

func processApiHelp() {
	fmt.Println("Usage:")
	fmt.Printf(" %v api [FLAGS]\n", constants.ExecutableName)
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  API_PORT\t\t\t\t\t The port that the service will listen on, defaults to 80")
	fmt.Println("  API_PREFIX\t\t\t\t\t The prefix that will be used for the api endpoints")
	fmt.Println("  LOG_LEVEL\t\t\t\t\t The log level of the service")
	fmt.Println("  HMAC_SECRET\t\t\t\t\t The secret that will be used to sign the jwt tokens")
	fmt.Println("  ENCRYPTION_PRIVATE_KEY\t\t\t The private key that will be used to encrypt the database at rest, you can generate one with the `gen-rsa` command")
	fmt.Println("  TLS_ENABLED\t\t\t\t\t If the service should use tls")
	fmt.Println("  TLS_PORT\t\t\t\t\t The port that the service will listen on for tls, defaults to 443")
	fmt.Println("  TLS_CERTIFICATE\t\t\t\t A base64 encoded certificate string")
	fmt.Println("  TLS_PRIVATE_KEY\t\t\t\t A base64 encoded private key string")
	fmt.Println("  DISABLE_CATALOG_CACHING\t\t\t If the service should disable the catalog caching")
	fmt.Println("  MODE\t\t\t\t\t\t The mode that the service will run in, this can be either `api` or `orchestrator` and defaults to `api`")
	fmt.Println("  USE_ORCHESTRATOR_RESOURCES\t\t\t If the service is running in orchestrator mode, this will allow the service to use the resources of the orchestrator")
	fmt.Println("  ORCHESTRATOR_PULL_FREQUENCY_SECOND\t\t The frequency in seconds that the orchestrator will sync with the other hosts in seconds")
	fmt.Println("  DATABASE_FOLDER\t\t\t\t The folder that the database will be stored in")
	fmt.Println("  CATALOG_CACHE_FOLDER\t\t\t\t The folder that the catalog cache will be stored in")
	fmt.Println()
	fmt.Println("Security Flags:")
	fmt.Println("  ROOT_PASSWORD\t\t\t\t\t The root password that will be used to update the root password of the virtual machine")
	fmt.Println("  JWT_SIGN_ALGORITHM\t\t\t\t The algorithm that will be used to sign the jwt tokens, this can be either `HS256`, `RS256`, `HS384`, `RS384`, `HS512`, `RS512`")
	fmt.Println("  JWT_PRIVATE_KEY\t\t\t\t The private key that will be used to sign the jwt tokens, this is only required if you are using `RS256`, `RS384` or `RS512`")
	fmt.Println("  JWT_HMACS_SECRET\t\t\t\t The secret that will be used to sign the jwt tokens, this is only required if you are using `HS256`, `HS384` or `HS512`")
	fmt.Println("  JWT_DURATION\t\t\t\t\t The duration that the jwt token will be valid for, you can use the following format, for example, 5 minutes would be `5m` or 1 hour would be `1h` ")
	fmt.Println()
	fmt.Println("Password Complexity Flags:")
	fmt.Println("  SECURITY_PASSWORD_MIN_PASSWORD_LENGTH\t\t The minimum length that the password should be, min is 8, defaults to 12")
	fmt.Println("  SECURITY_PASSWORD_MAX_PASSWORD_LENGTH\t\t The maximum length that the password should be, max is 40, defaults to 40")
	fmt.Println("  SECURITY_PASSWORD_REQUIRE_UPPERCASE\t\t If the password should require at least one uppercase character, defaults to true")
	fmt.Println("  SECURITY_PASSWORD_REQUIRE_LOWERCASE\t\t If the password should require at least one lowercase character, defaults to true")
	fmt.Println("  SECURITY_PASSWORD_REQUIRE_NUMBER\t\t If the password should require at least one number, defaults to true")
	fmt.Println("  SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR\t If the password should require at least one special character, defaults to true")
	fmt.Println("  SECURITY_PASSWORD_SALT_PASSWORD\t\t If the password should be salted, defaults to true")
	fmt.Println()
	fmt.Println("Brute Force Guard Flags:")
	fmt.Println("  BRUTE_FORCE_MAX_LOGIN_ATTEMPTS\t\t The maximum number of login attempts before the account is locked, defaults to 5")
	fmt.Println("  BRUTE_FORCE_LOCK_DURATION\t\t\t The duration that the account will be locked for, you can use the following format, for example, 5 minutes would be `5m` or 1 hour would be `1h`, defaults to 5 seconds")
	fmt.Println("  BRUTE_FORCE_INCREMENTAL_WAIT\t\t\t If the wait period should be incremental, if set to false, the wait period will be the same for each failed attempt, defaults to true")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Printf("  %v api --ROOT_PASSWORD=VeryStrongPassw0rd!", constants.ExecutableName)
	fmt.Println()
}
