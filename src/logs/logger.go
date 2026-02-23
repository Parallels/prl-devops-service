package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

var globalLogFilePath string

func SetupFileLogger(logFilename string, ctx basecontext.ApiContext) {
	// Loading the configuration file
	cfg := config.Get()
	cfg.Load()

	sysSvc := system.Get()
	currentUser := "root"
	sysUser, err := sysSvc.GetCurrentUser(ctx)
	if err == nil {
		currentUser = sysUser
	}
	executableFilePath := filepath.Dir(os.Args[0])
	// check if we should log to file
	if cfg.GetKey(constants.LOG_TO_FILE_ENV_VAR) == "true" || cfg.GetKey(constants.PRL_DEVOPS_LOG_TO_FILE_ENV_VAR) == "true" {
		// Setting the default path to the executable path
		logFilePath := executableFilePath
		// Checking if the user is root and the operating system is linux or macos
		// if so we will change the log file path to /var/log
		if currentUser == "root" && (sysSvc.GetOperatingSystem() == "linux" || sysSvc.GetOperatingSystem() == "macos") {
			logFilePath = "/var/log"
		} else {
			// Getting the user home path
			if userHomePath, err := sysSvc.GetUserHome(ctx, currentUser); err == nil {
				logFilePath = fmt.Sprintf("%s/.prl-devops-service/logs", userHomePath)
				// checking if the folder is not created, if not create it
				if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
					if err := os.MkdirAll(logFilePath, os.ModePerm); err != nil {
						ctx.LogErrorf("Error creating log folder: %v", err)
						// if we can't create the folder, we will use the executable path
						logFilePath = filepath.Dir(os.Args[0])
					}
				}
			}
		}

		// checking if a custom path is set in the environment variables
		envFilePath := cfg.GetKey(constants.LOG_FILE_PATH_ENV_VAR)
		if envFilePath == "" {
			envFilePath = cfg.GetKey(constants.PRL_DEVOPS_LOG_FILE_PATH_ENV_VAR)
		}
		if envFilePath != "" && envFilePath != "." {
			baseFolder := filepath.Dir(envFilePath)
			if _, err := os.Stat(baseFolder); os.IsNotExist(err) {
				ctx.LogErrorf("[Core] Log file path does not exist: %s, using executable path", envFilePath)
			} else {
				logFilePath = envFilePath
			}
		}

		logFullPath := filepath.Join(logFilePath, constants.DEFAULT_LOG_FILE_NAME)

		// checking if this is a debug executable, if it is we will change the log file name
		executable, err := os.Executable()
		if err == nil && strings.Contains(executable, "__debug") {
			// Remove any debug log files
			files, err := filepath.Glob(filepath.Join(executableFilePath, "*__debug_bin*.log*"))
			if err == nil {
				for _, file := range files {
					os.Remove(file)
				}
			}
			executableFileName := filepath.Base(executable)
			logFullPath = filepath.Join(executableFilePath, executableFileName+".log")
		}

		logger := ctx.Logger()
		ctx.LogDebugf("Adding file logger: %s", logFullPath)
		logger.AddFileLogger(logFullPath)
		logger.AddChannelLogger() // required to stream logs via OnMessage
		globalLogFilePath = logFullPath
	}
}

func IsLogFileEnabled(ctx basecontext.ApiContext) bool {
	cfg := config.Get()
	logToFile := cfg.GetBoolKey(constants.LOG_TO_FILE_ENV_VAR)
	return logToFile
}

func GetLogFilePath(ctx basecontext.ApiContext) string {
	return globalLogFilePath
}
