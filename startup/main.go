package startup

import (
	"Parallels/pd-api-service/service_provider"
)

func Start() {
	service_provider.InitServices()

	// Seeding defaults
	if err := SeedDefaults(); err != nil {
		panic(err)
	}
}

func Restart() {
	listener.Restart()
}
