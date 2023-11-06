package startup

import (
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/Parallels/pd-api-service/serviceprovider/system"
)

func Start() {
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
}

func Restart() {
	listener.Restart()
}
