package startup

import (
	"encoding/base64"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/install"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/jobs/tracker"
	"github.com/Parallels/prl-devops-service/logs"
	"github.com/Parallels/prl-devops-service/orchestrator"
	"github.com/Parallels/prl-devops-service/reverse_proxy"
	bruteforceguard "github.com/Parallels/prl-devops-service/security/brute_force_guard"
	"github.com/Parallels/prl-devops-service/security/jwt"
	"github.com/Parallels/prl-devops-service/security/password"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	diskspace "github.com/Parallels/prl-devops-service/serviceprovider/diskSpace"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
	"github.com/Parallels/prl-devops-service/serviceprovider/health"
	providerlogs "github.com/Parallels/prl-devops-service/serviceprovider/logs"
	"github.com/Parallels/prl-devops-service/serviceprovider/stats"
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

	logs.SetupFileLogger(constants.DEFAULT_LOG_FILE_NAME, ctx)

	_ = tracker.NewProgressService(ctx)

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

	// Keep HOST_MODE in sync with actual Parallels Desktop availability.
	// We check both directions so that a freshly deployed agent that just had
	// PD installed (and may not have included "host" in ENABLED_MODULES, or
	// where PD was not ready at the previous boot) gets the module enabled
	// automatically.
	if system.GetOperatingSystem() == "macos" || system.GetOperatingSystem() == "darwin" {
		provider := serviceprovider.Get()
		pdAvailable := provider.IsParallelsDesktopAvailable()
		if cfg.IsModuleEnabled(constants.HOST_MODE) && !pdAvailable {
			ctx.LogWarnf("Parallels Desktop is not available, disabling host module")
			cfg.DisableModule(constants.HOST_MODE)
		} else if !cfg.IsModuleEnabled(constants.HOST_MODE) && pdAvailable {
			ctx.LogInfof("Parallels Desktop is available; auto-enabling host module")
			cfg.EnableModule(constants.HOST_MODE)
			// Persist the updated module list to the service config file so
			// the setting survives a service restart.
			install.PersistEnabledModules(ctx)
		}
	}

	ctx.LogInfof("Enabled Modules: %v", cfg.GetEnabledModules())

	telemetry.TrackEvent(telemetry.NewTelemetryItem(ctx, telemetry.EventStartApi, nil, nil))

	// Initialize EventEmitter service (for API and Orchestrator modes)
	if cfg.IsApi() || cfg.IsOrchestrator() {
		emitter := eventemitter.NewEventEmitter(ctx)
		if diag := emitter.Initialize(); diag.HasErrors() {
			ctx.LogErrorf("Failed to initialize EventEmitter: %v", diag.GetErrors())
		} else {
			// Register handlers
			health.NewHealthService(emitter)
			eventemitter.NewSystemHandler(emitter)

			// Start Stats Service
			statsService := stats.NewStatsService(emitter)
			go statsService.Run(ctx, 5*time.Second)

			// Start Log Service
			logService := providerlogs.NewLogService(emitter)
			go logService.Run(ctx)
		}
	}

	// Initialize DiskSpace Service (for Host and Orchestrator modes)
	if cfg.IsHost() || cfg.IsOrchestrator() {
		ds := diskspace.New(ctx)
		provider := serviceprovider.Get()
		if provider.ParallelsDesktopService != nil {
			ds.SetParallelsHomePathProvider(provider.ParallelsDesktopService.GetUserHome)
		}
		ds.Start()
	}

	// Seeding defaults
	ctx.LogInfof("Seeding defaults")
	if err := SeedDefaults(); err != nil {
		panic(err)
	}

	// Clean up expired/used enrollment tokens at startup
	if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
		if err := dbService.DeleteExpiredEnrollmentTokens(ctx); err != nil {
			ctx.LogWarnf("Could not purge expired enrollment tokens: %v", err)
		}
	}

	ctx.LogInfof("Applying migrations")
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
				createdKey := false
				// Find the local host by description or IsLocal flag — URL matching via
				// GetOrchestratorHost is unreliable here because the stored host includes
				// PathPrefix in GetHost() while cfg.Localhost() does not.
				var localhost *models.OrchestratorHost
				if allHosts, hErr := dbService.GetOrchestratorHosts(ctx, ""); hErr == nil {
					for i := range allHosts {
						if allHosts[i].IsLocal || allHosts[i].Description == constants.LOCAL_ORCHESTRATOR_DESCRIPTION {
							localhost = &allHosts[i]
							break
						}
					}
				}
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
						Enabled:     true,
						IsLocal:     true,
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
					// Ensure existing local host has the correct flags, regardless of how it was created
					localhost.Enabled = true
					localhost.IsLocal = true
					if createdKey {
						secret, err := cryptorand.GetAlphaNumericRandomString(32)
						if err != nil {
							ctx.LogErrorf("Error generating secret: %v", err)
							panic(err)
						}

						localhost.Authentication = &models.OrchestratorHostAuthentication{
							ApiKey: base64.StdEncoding.EncodeToString([]byte(ORCHESTRATOR_KEY_NAME + ":" + secret)),
						}
					}
					_, _ = dbService.UpdateOrchestratorHost(
						ctx,
						localhost,
					)
				}
			}
		} else {
			// checking if we need to remove the current host from the orchestrator hosts
			if dbService, err := serviceprovider.GetDatabaseService(ctx); err == nil {
				if allHosts, hErr := dbService.GetOrchestratorHosts(ctx, ""); hErr == nil {
					for _, h := range allHosts {
						if h.IsLocal || h.Description == constants.LOCAL_ORCHESTRATOR_DESCRIPTION {
							ctx.LogInfof("Removing local orchestrator host")
							_ = dbService.DeleteOrchestratorHost(ctx, h.ID)
							break
						}
					}
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

	ctx.LogInfof("Starting Job Manager Service")
	jobManagerService := jobs.New(ctx)

	// Wire up the NotificationService → JobManager callbacks.
	// We use the combined OnUpdateJobProgressAndSteps callback so that progress
	// and the step snapshot are written to the DB in a single operation, which
	// prevents the race condition where a separate UpdateJobProgress read could
	// see stale (empty) steps that were written by a concurrent UpdateJobSteps.
	if ns := tracker.GetProgressService(); ns != nil {
		ns.OnUpdateJobProgressAndSteps = func(jobId string, percent int, state string, steps []data_models.JobStep) {
			jobManagerService.UpdateJobProgressAndSteps(jobId, percent, constants.JobState(state), steps)
		}
		// Keep the legacy callbacks as fallbacks (called only when OnUpdateJobProgressAndSteps is nil)
		ns.OnUpdateJobSteps = func(jobId string, steps []data_models.JobStep) {
			jobManagerService.UpdateJobSteps(jobId, steps)
		}
		ns.OnUpdateJobProgress = func(jobId string, percent int, status string) {
			jobManagerService.UpdateJobProgress(jobId, percent, constants.JobState(status))
		}
		ns.OnUpdateJobMessage = func(jobId string, message string) {
			jobManagerService.UpdateJobMessage(jobId, message)
		}
		ns.OnUpdateJobResultRecord = func(jobId string, recordId string, recordName string, recordType string, recordLinkId string) {
			jobManagerService.UpdateJobResultRecord(jobId, recordId, recordName, recordType, recordLinkId)
		}
		ns.OnInitJob = func(jobId string) {
			jobManagerService.InitJob(jobId)
		}
	}

	go func() {
		if err := jobManagerService.Start(); err != nil {
			ctx.LogErrorf("Error starting job manager service: %v", err)
		}
	}()
}

func Restart() {
	listener.Restart()
}
