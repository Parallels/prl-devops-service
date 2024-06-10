package seeds

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func SeedDefaultConfig() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	err := db.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	if _, err := db.SeedId(ctx); err != nil {
		return err
	}

	_ = db.Disconnect(ctx)

	return nil
}
