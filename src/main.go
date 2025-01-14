package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/cmd"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/Parallels/prl-devops-service/telemetry"

	"github.com/cjlapao/common-go/version"
)

var ver = "0.9.12"

//	@title			Parallels Desktop DevOps Service
//	@version		0.9.12
//	@description	Parallels Desktop DevOps Service
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Parallels Desktop DevOps Support
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
	// catching all of the exceptions
	defer func() {
		// Saving the database before exiting

		if err := recover(); err != nil {
			sp := serviceprovider.Get()
			if sp != nil {
				db := sp.JsonDatabase
				if db != nil {
					ctx := basecontext.NewRootBaseContext()
					_ = db.SaveNow(ctx)
					_ = db.SaveAs(ctx, fmt.Sprintf("data.json.%s.panic", strings.ReplaceAll(time.Now().Format("20060102150405"), "-", "_")))
				}
			}
			fmt.Fprintf(os.Stderr, "Exception: %v\n", err)
			os.Exit(1)
		}
	}()
	versionSvc := system.VersionSvc

	versionSvc.Author = "Carlos Lapao"
	versionSvc.Name = constants.Name
	versionSvc.License = "Fair Source (https://fair.io)"

	// Reading the version from a string
	strVer, err := version.FromString(ver)

	if err == nil {
		versionSvc.Major = strVer.Major
		versionSvc.Minor = strVer.Minor
		versionSvc.Build = strVer.Build
		versionSvc.Rev = strVer.Rev
	}

	// setting a build id for debug builds
	if strings.Contains(os.Args[0], "__debug") {
		versionSvc.Build = int(time.Now().Unix())
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()
	cfg := config.New(ctx)
	cfg.Load()
	enableLogToFile := cfg.GetKey(constants.LOG_TO_FILE_ENV_VAR)
	if enableLogToFile == "true" {
		logFilename := "prldevops.log"
		filePath := cfg.GetKey(constants.LOG_FILE_PATH_ENV_VAR)
		if filePath != "" {
			baseFolder := filepath.Dir(filePath)
			if _, err := os.Stat(baseFolder); os.IsNotExist(err) {
				ctx.LogErrorf("[Core] Log file path does not exist: %s, using executable path", filePath)
			} else {
				logFilename = filepath.Join(filePath, logFilename)
			}
		} else {
			logPath := filepath.Dir(os.Args[0])
			logFilename = filepath.Join(logPath, logFilename)
		}

		executable, err := os.Executable()
		if err == nil && !strings.Contains(executable, "__debug") {
			logFilename = executable + ".log"
		}
		ctx.EnableLogFile(logFilename)
	}

	go func() {
		<-c
		cfg := config.Get()
		if cfg.GetRunningCommand() == constants.API_COMMAND || cfg.GetRunningCommand() == "" {
			sp := serviceprovider.Get()
			if sp != nil {
				db := sp.JsonDatabase
				if db != nil {
					cleanup(ctx, db)
					retries := 0
					maxRetries := 10
					for {
						retries++
						if !db.IsConnected() || retries > maxRetries {
							break
						}
						ctx.LogInfof("[Core] Waiting for database to disconnect...")
						time.Sleep(5 * time.Second)
					}
				}
			}
		}
		os.Exit(0)
	}()

	go func() {
		for {
			select {
			case <-c:
				ctx.LogInfof("[Core] Exiting")
				return
			default:
				// Call home every 30 minutes
				time.Sleep(30 * time.Minute)
				ctx.LogInfof("[Core] Sending heartbeat")
				sendHeartbeat()
			}
		}
	}()

	cmd.Process()
}

func cleanup(ctx basecontext.ApiContext, db *data.JsonDatabase) {
	if db != nil {
		ctx.LogInfof("[Core] Saving database")
		if err := db.SaveNow(ctx); err != nil {
			ctx.LogErrorf("[Core] Error saving database: %v", err)
		} else {
			ctx.LogInfof("[Core] Database saved")
		}
		_ = db.Disconnect(ctx)
	}
}

func sendHeartbeat() {
	if telemetry.Get() == nil {
		return
	}
	ctx := basecontext.NewRootBaseContext()
	telemetry.TrackEvent(telemetry.NewTelemetryItem(ctx, telemetry.HeartbeatEvent, nil, nil))
}
