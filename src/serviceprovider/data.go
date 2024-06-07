package serviceprovider

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
)

func GetDatabaseService(ctx basecontext.ApiContext) (*data.JsonDatabase, error) {
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
