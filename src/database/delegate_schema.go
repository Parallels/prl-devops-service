package database

// Schema version bookkeeping keeps delegating to legacy until the SQL schema
// tracker lands. TODO(#421 task 08): SQL-backed schema version.

import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

func (g *GormDatabase) GetSchemaVersion(ctx basecontext.ApiContext) (string, error) {
	return g.legacy.GetSchemaVersion(ctx)
}

func (g *GormDatabase) UpdateSchemaVersion(ctx basecontext.ApiContext, version string) error {
	return g.legacy.UpdateSchemaVersion(ctx, version)
}
