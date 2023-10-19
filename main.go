package main

import (
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/services"
	"Parallels/pd-api-service/startup"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/version"
)

var versionSvc = version.Get()

func main() {
	versionSvc.Author = "Carlos Lapao"
	versionSvc.Name = "POC Parallels Desktop API Service"
	versionSvc.License = "MIT"
	versionSvc.Minor = 1
	versionSvc.Major = 0
	versionSvc.Build = 20

	if helper.GetFlagSwitch("version", false) {
		println(versionSvc.String())
		os.Exit(0)
	}

	versionSvc.PrintAnsiHeader()
	services.InitServices()

	port := helper.GetFlagValue("port", "")

	if port == "" {
		port = os.Getenv("PORT")
	}

	if port == "" {
		port = "8080"
	}

	r := startup.InitControllers()
	// Seeding defaults
	startup.SeedVirtualMachineTemplateDefaults()
	startup.SeedAdminUser()

	// Serve the API
	services.GetServices().Logger.Info("Serving API on port %s", port)
	services.GetServices().Logger.Info("Api Prefix %s", constants.API_PREFIX)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
