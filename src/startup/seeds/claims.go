package seeds

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/serviceprovider"
)

func SeedDefaultClaims() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	err := db.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	allSystemClaims := constants.AllSystemClaims

	for _, claim := range allSystemClaims {
		if exists, _ := db.GetClaim(ctx, claim); exists == nil {
			if _, err := db.CreateClaim(ctx, models.Claim{
				ID:       claim,
				Name:     claim,
				Internal: true,
			}); err != nil {
				common.Logger.Error("Error adding claim: %s", err.Error())
				return err
			}
		} else {
			ctx.LogDebugf("Claim already exists: %s", claim)
		}
	}

	_ = db.Disconnect(ctx)

	return nil
}
