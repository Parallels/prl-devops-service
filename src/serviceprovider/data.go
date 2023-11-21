package serviceprovider

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/data"
)

func GetDatabaseService(ctx basecontext.ApiContext) (*data.JsonDatabase, error) {
	provider := Get()
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
