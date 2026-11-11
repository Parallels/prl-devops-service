package seeds

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func SeedDefaultConfig() error {
	ctx := basecontext.NewRootBaseContext()
	// T7 (#424 boot matrix): engine-neutral guard like every #421-18 seed —
	// GetDatabaseService carries Connect; Disconnect is NEVER ours to call
	// (post-flip it closes the provider's shared pool permanently).
	db, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	if _, err := db.SeedId(ctx); err != nil {
		return err
	}

	return nil
}
