package migrations

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
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
	result.DbService = serviceprovider.Get().JsonDatabase
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
