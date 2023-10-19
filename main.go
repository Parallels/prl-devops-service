package main

import (
	"Parallels/pd-api-service/services"
	"Parallels/pd-api-service/startup"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cjlapao/common-go/version"
)

var versionSvc = version.Get()

func main() {
	versionSvc.Author = "Carlos Lapao"
	versionSvc.Name = "POC Parallels Desktop API Service"
	versionSvc.License = "MIT"
	versionSvc.Minor = 1
	versionSvc.Major = 0
	versionSvc.Build = 18
	versionSvc.PrintAnsiHeader()
	services.InitServices()

	// if the argument is equal to migrations, execute the migrations
	for _, arg := range os.Args {
		if arg == "migrations" {
			startup.ExecuteMigrations()
			os.Exit(0)
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := startup.InitControllers()
	// Seeding defaults
	startup.SeedVirtualMachineTemplateDefaults()

	// Serve the API
	services.GetServices().Logger.Info("Serving API on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
