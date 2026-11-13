package migrations

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

type MigrationService struct {
	Context       basecontext.ApiContext
	DbService     *data.JsonDatabase
	schemaVersion string
}

func Init() (*MigrationService, error) {
	result := MigrationService{}
	result.Context = basecontext.NewRootBaseContext()
	result.DbService = serviceprovider.Get().Legacy() // #421: stays concrete (AD-4: legacy-store-only)
	// #421 task 18 (D-7): SQL mode runs with a NIL legacy handle. The startup
	// migrator loop gates these out already; this guard is defense-in-depth
	// so a stray caller errors instead of nil-panicking a SQL boot.
	if result.DbService == nil {
		return nil, errors.NewWithCode("legacy schema migrator requires the legacy database backend", 400)
	}
	err := result.DbService.Connect(result.Context)
	if err != nil {
		return nil, err
	}

	schemaVersion, err := result.DbService.GetSchemaVersion(result.Context)
	if err != nil {
		return nil, err
	}
	if schemaVersion == "" {
		schemaVersion = "0.0.0"
	}

	result.schemaVersion = schemaVersion

	return &result, nil
}
