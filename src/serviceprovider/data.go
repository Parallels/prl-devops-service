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

	dbService, ok := provider.JsonDatabase.(*data.JsonDatabase) // #421 D-6 S1: field is the interface now; assertion until task 17 flips the signature
	if !ok || dbService == nil {
		return nil, data.ErrDatabaseNotConnected
	}

	err := dbService.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return dbService, nil
}
