package data

import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

func (j *JsonDatabase) GetSchemaVersion(ctx basecontext.ApiContext) (string, error) {
	if !j.IsConnected() {
		return "", ErrDatabaseNotConnected
	}

	return j.data.Schema.Version, nil
}

func (j *JsonDatabase) UpdateSchemaVersion(ctx basecontext.ApiContext, version string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	j.data.Schema.Version = version
	j.Save(ctx)

	return nil
}
