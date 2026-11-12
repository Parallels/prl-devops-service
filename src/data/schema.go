package data

import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v1.3.0.
func (j *JsonDatabase) GetSchemaVersion(ctx basecontext.ApiContext) (string, error) {
	if !j.IsConnected() {
		return "", ErrDatabaseNotConnected
	}

	return j.data.Schema.Version, nil
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v1.3.0.
func (j *JsonDatabase) UpdateSchemaVersion(ctx basecontext.ApiContext, version string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	j.data.Schema.Version = version

	return nil
}
