package main

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/service_provider"
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
	versionSvc.Build = 31

	if helper.GetFlagSwitch("version", false) {
		println(versionSvc.String())
		os.Exit(0)
	}

	versionSvc.PrintAnsiHeader()
	startup.Start()

	if helper.GetFlagSwitch("update-root-pass", false) {
		common.Logger.Info("Updating root password")
		rootPassword := helper.GetFlagValue("password", "")
		if rootPassword != "" {
			db := service_provider.Get().JsonDatabase
			common.Logger.Info("Database connection found, updating password")
			db.Connect()
			if db != nil {
				err := db.UpdateRootPassword(rootPassword)
				if err != nil {
					panic(err)
				}
			} else {
				panic("No database connection")
			}
		} else {
			panic("No password provided")
		}
		common.Logger.Info("Root password updated")
		os.Exit(0)
	}

	port := helper.GetFlagValue("port", "")

	if port == "" {
		port = os.Getenv("PORT")
	}

	if port == "" {
		port = "8080"
	}
	// Serve the API

	for {
		listener := startup.InitApi()
		listener.Options.HttpPort = port
		restartChannel := restapi.GetRestartChannel()
		listener.Start(versionSvc.Name, versionSvc.String())

		needsRestart := <-restartChannel
		if !needsRestart {
			break
		}
		startup.Start()
	}
}
