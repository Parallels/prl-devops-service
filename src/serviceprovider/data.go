package serviceprovider

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
)

// #421 D-6 S2 (task 17): the consumer-facing surface is the interface.
// Until S3 (task 18) the returned value is still the legacy instance; the
// type system, not a convention, now prevents access to legacy-only
// machinery (D-7 escape hatch: Legacy()).
func GetDatabaseService(ctx basecontext.ApiContext) (datainterfaces.DatabaseService, error) {
	provider := Get()
	if provider == nil {
		return nil, data.ErrDatabaseNotConnected
	}

	dbService := provider.JsonDatabase
	if dbService == nil {
		return nil, data.ErrDatabaseNotConnected
	}

	err := dbService.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return dbService, nil
}
