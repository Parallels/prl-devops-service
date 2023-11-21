package startup

import (
	"encoding/base64"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/orchestrator"
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/Parallels/pd-api-service/serviceprovider/system"
)

func Start() {
	config := config.NewConfig()
	config.GetLogLevel()

	system := system.New()
	if system.GetOperatingSystem() != "macos" {
		serviceprovider.InitCatalogServices()
	} else {
		serviceprovider.InitServices()
	}

	// Seeding defaults
	if err := SeedDefaults(); err != nil {
		panic(err)
	}

	if config.IsOrchestrator() {
		ctx := basecontext.NewRootBaseContext()
		ctx.LogInfo("Starting Orchestrator Background Service")
		// Checking if we need to add the current host to the orchestrator hosts
		if config.UseOrchestratorResources() {
			if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
				hostName := config.GetLocalhost()
				localhost, _ := dbService.GetOrchestratorHost(ctx, hostName)
				_, err := dbService.CreateApiKey(ctx, models.ApiKey{
					Key:    "orchestrator_key",
					Name:   "orchestrator_key",
					Secret: serviceprovider.Get().HardwareSecret,
				})
				if err != nil {
					ctx.LogError("Error creating orchestrator key: %v", err)
					panic(err)
				}

				if localhost == nil {
					ctx.LogInfo("Creating local orchestrator host")
					dbService.CreateOrchestratorHost(ctx, models.OrchestratorHost{
						ID:          helpers.GenerateId(),
						Host:        "localhost",
						Description: "Local Orchestrator",
						Tags:        []string{"localhost", "local"},
						PathPrefix:  config.GetApiPrefix(),
						Schema:      "http",
						Port:        config.GetApiPort(),
						Authentication: &models.OrchestratorHostAuthentication{
							ApiKey: base64.StdEncoding.EncodeToString([]byte("orchestrator_key:" + serviceprovider.Get().HardwareSecret)),
						},
					})
				}
			}
		} else {
			// checking if we need to remove the current host from the orchestrator hosts
			if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
				hostName := config.GetLocalhost()
				localhost, _ := dbService.GetOrchestratorHost(ctx, hostName)
				if localhost != nil {
					ctx.LogInfo("Removing local orchestrator host")
					dbService.DeleteOrchestratorHost(ctx, localhost.ID)
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
