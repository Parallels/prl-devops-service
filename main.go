package main

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/services"
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
	versionSvc.Minor = 1
	versionSvc.Major = 0
	versionSvc.Build = 29

	if helper.GetFlagSwitch("version", false) {
		println(versionSvc.String())
		os.Exit(0)
	}

	versionSvc.PrintAnsiHeader()
	services.InitServices()

	if helper.GetFlagSwitch("update-root-pass", false) {
		common.Logger.Info("Updating root password")
		rootPassword := helper.GetFlagValue("password", "")
		if rootPassword != "" {
			db := services.GetServices().JsonDatabase
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
	} else {
		common.Logger.Info("Not updating root password")
	}

	port := helper.GetFlagValue("port", "")

	if port == "" {
		port = os.Getenv("PORT")
	}

	if port == "" {
		port = "8080"
	}

	listener := startup.InitControllers()
	listener.Options.HttpPort = port

	// Seeding defaults
	err := startup.SeedVirtualMachineTemplateDefaults()
	if err != nil {
		panic(err)
	}
	err = startup.SeedDefaultRolesAndClaims()
	if err != nil {
		panic(err)
	}

	err = startup.SeedAdminUser()
	if err != nil {
		panic(err)
	}

	// Serve the API
	listener.Start(versionSvc.Name, versionSvc.String())
}
