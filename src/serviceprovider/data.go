package serviceprovider

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider/dbservice"
)

// GetDatabaseService returns the GORM-based database service
// This is the new implementation using database stores
func GetDatabaseService(ctx basecontext.ApiContext) (*dbservice.DatabaseService, *apperrors.Diagnostics) {
	db := dbservice.GetDatabaseService()
	if db == nil {
		diag := apperrors.NewDiagnostics("get_database_service")
		diag.AddError("not_connected", "database not connected - call InitDatabase first", "serviceprovider", nil)
		return nil, diag
	}
	return db, nil
}

// GetJsonDatabaseService returns the legacy JSON database service
// Deprecated: Use GetDatabaseService instead
func GetJsonDatabaseService(ctx basecontext.ApiContext) (*data.JsonDatabase, error) {
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
