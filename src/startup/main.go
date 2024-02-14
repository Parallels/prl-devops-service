package startup

import (
	"encoding/base64"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/orchestrator"
	bruteforceguard "github.com/Parallels/prl-devops-service/security/brute_force_guard"
	"github.com/Parallels/prl-devops-service/security/jwt"
	"github.com/Parallels/prl-devops-service/security/password"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
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

	system := system.SystemService{}
	if system.GetOperatingSystem() != "macos" {
		serviceprovider.InitCatalogServices(ctx)
	} else {
		serviceprovider.InitServices(ctx)
	}

	// Seeding defaults
	if err := SeedDefaults(); err != nil {
		panic(err)
	}

	if cfg.IsOrchestrator() {
		ctx := basecontext.NewRootBaseContext()
		ctx.LogInfof("Starting Orchestrator Background Service")
		// Checking if we need to add the current host to the orchestrator hosts
		if cfg.UseOrchestratorResources() {
			if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
				hostName := cfg.Localhost()
				createdKey := false
				localhost, _ := dbService.GetOrchestratorHost(ctx, hostName)
				apiKey, err := dbService.GetApiKey(ctx, ORCHESTRATOR_KEY_NAME)
				secret := serviceprovider.Get().HardwareSecret
				if err != nil {
					if errors.GetSystemErrorCode(err) != 404 {
						ctx.LogErrorf("Error getting orchestrator key: %v", err)
						panic(err)
					}
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
						Description: "Local Orchestrator",
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
						secret := serviceprovider.Get().HardwareSecret
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
}

func Restart() {
	listener.Restart()
}
