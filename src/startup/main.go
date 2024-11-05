package startup

import (
	"encoding/base64"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/orchestrator"
	"github.com/Parallels/prl-devops-service/reverse_proxy"
	bruteforceguard "github.com/Parallels/prl-devops-service/security/brute_force_guard"
	"github.com/Parallels/prl-devops-service/security/jwt"
	"github.com/Parallels/prl-devops-service/security/password"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/Parallels/prl-devops-service/startup/migrations"
	"github.com/Parallels/prl-devops-service/telemetry"
	cryptorand "github.com/cjlapao/common-go-cryptorand"
)

const (
	ORCHESTRATOR_KEY_NAME = "orchestrator_key"
)

func Init(ctx basecontext.ApiContext) {
	cfg := config.New(ctx)
	cfg.Load()

	password.New(ctx)
	jwt.New(ctx)
	bruteforceguard.New(ctx)
}

func Start(ctx basecontext.ApiContext) {
	cfg := config.Get()
	schemaMigrations := make([]migrations.Migration, 0)
	schemaMigrations = append(schemaMigrations, migrations.Version0_6_0{})

	system := system.SystemService{}
	if system.GetOperatingSystem() != "macos" {
		serviceprovider.InitCatalogServices(ctx)
	} else {
		serviceprovider.InitServices(ctx)
	}

	// initializing telemetry with default context
	_ = telemetry.New(ctx)

	telemetry.TrackEvent(telemetry.NewTelemetryItem(ctx, telemetry.EventStartApi, nil, nil))

	// Seeding defaults
	if err := SeedDefaults(); err != nil {
		panic(err)
	}

	for _, migration := range schemaMigrations {
		if err := migration.Apply(); err != nil {
			ctx.LogErrorf("Error applying migration: %v", err)
		}
	}

	// lets import any reverse proxy configuration hosts we have into the db and remove them from the config
	// this is to allow any misconfiguration to be corrected and to import old configurations
	// rpConfig := cfg.GetReverseProxyConfig()
	// if rpConfig != nil {
	// 	for _, host := range rpConfig.Hosts {
	// 		if host != nil {
	// 			if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
	// 				if _, err := dbService.GetReverseProxyHost(ctx, host.Host); err != nil {
	// 					if errors.GetSystemErrorCode(err) == 404 {
	// 						dbHost := models.ReverseProxyHost{
	// 							ID:   helpers.GenerateId(),
	// 							Host: host.Host,
	// 						}

	// 						_, _ = dbService.CreateReverseProxyHost(ctx, host)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	if cfg.IsOrchestrator() {
		ctx := basecontext.NewRootBaseContext()
		ctx.LogInfof("Starting Orchestrator Background Service")
		canUseOwnResources := false
		if system.GetOperatingSystem() == "linux" {
			canUseOwnResources = false
		} else if system.GetOperatingSystem() == "windows" {
			canUseOwnResources = false
		} else if system.GetOperatingSystem() == "macos" || system.GetOperatingSystem() == "darwin" {
			canUseOwnResources = true
		}

		// Checking if we need to add the current host to the orchestrator hosts
		if cfg.UseOrchestratorResources() && canUseOwnResources {
			if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
				hostName := cfg.Localhost()
				createdKey := false
				localhost, _ := dbService.GetOrchestratorHost(ctx, hostName)
				apiKey, err := dbService.GetApiKey(ctx, ORCHESTRATOR_KEY_NAME)
				if err != nil {
					if errors.GetSystemErrorCode(err) != 404 {
						ctx.LogErrorf("Error getting orchestrator key: %v", err)
						panic(err)
					}
				}
				secret, err := cryptorand.GetAlphaNumericRandomString(32)
				if err != nil {
					ctx.LogErrorf("Error generating secret: %v", err)
					panic(err)
				}

				if apiKey == nil {
					_, err := dbService.CreateApiKey(ctx, models.ApiKey{
						Key:    ORCHESTRATOR_KEY_NAME,
						Name:   ORCHESTRATOR_KEY_NAME,
						Secret: secret,
					})
					if err != nil {
						if errors.GetSystemErrorCode(err) != 404 {
							ctx.LogErrorf("Error creating orchestrator key: %v", err)
							panic(err)
						}
					}
					createdKey = true
				}

				if localhost == nil {
					ctx.LogInfof("Creating local orchestrator host")
					_, _ = dbService.CreateOrchestratorHost(ctx, models.OrchestratorHost{
						ID:          helpers.GenerateId(),
						Host:        "localhost",
						Description: constants.LOCAL_ORCHESTRATOR_DESCRIPTION,
						Tags:        []string{"localhost", "local"},
						PathPrefix:  cfg.ApiPrefix(),
						Schema:      "http",
						Port:        cfg.ApiPort(),
						Authentication: &models.OrchestratorHostAuthentication{
							ApiKey: base64.StdEncoding.EncodeToString([]byte(ORCHESTRATOR_KEY_NAME + ":" + secret)),
						},
					})
				} else {
					if createdKey {
						secret, err := cryptorand.GetAlphaNumericRandomString(32)
						if err != nil {
							ctx.LogErrorf("Error generating secret: %v", err)
							panic(err)
						}

						localhost.Authentication = &models.OrchestratorHostAuthentication{
							ApiKey: base64.StdEncoding.EncodeToString([]byte(ORCHESTRATOR_KEY_NAME + ":" + secret)),
						}
						_, _ = dbService.UpdateOrchestratorHost(
							ctx,
							localhost,
						)
					}
				}
			}
		} else {
			// checking if we need to remove the current host from the orchestrator hosts
			if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
				hostName := cfg.Localhost()
				localhost, _ := dbService.GetOrchestratorHost(ctx, hostName)
				if localhost != nil {
					ctx.LogInfof("Removing local orchestrator host")
					_ = dbService.DeleteOrchestratorHost(ctx, localhost.ID)
				}
				apiKey, _ := dbService.GetApiKey(ctx, ORCHESTRATOR_KEY_NAME)
				if apiKey != nil {
					ctx.LogInfof("Removing local orchestrator key")
					_ = dbService.DeleteApiKey(ctx, apiKey.ID)
				}
			}
		}
		orchestratorBackgroundService := orchestrator.NewOrchestratorService(ctx)
		go orchestratorBackgroundService.Start(true)
	}
	if cfg.IsReverseProxyEnabled() {
		ctx.LogInfof("Starting Reverse Proxy Service")
		reverseProxyService := reverse_proxy.New(ctx)
		go func() {
			if err := reverseProxyService.Start(); err != nil {
				ctx.LogErrorf("Error starting reverse proxy service: %v", err)
			}
		}()
	}
}

func Restart() {
	listener.Restart()
}
