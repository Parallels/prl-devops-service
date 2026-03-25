package seeds

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
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

	for _, claimID := range allSystemClaims {
		meta, hasMeta := constants.ClaimCategoryMap[claimID]
		group := constants.ClaimGroupCustom
		resource := ""
		action := ""
		if hasMeta {
			group = meta.Group
			resource = meta.Resource
			action = meta.Action
		}
		description := constants.ClaimDescriptionMap[claimID] // "" if not found

		if exists, _ := db.GetClaim(ctx, claimID); exists == nil {
			if _, err := db.CreateClaim(ctx, models.Claim{
				ID:          claimID,
				Name:        claimID,
				Internal:    true,
				Description: description,
				Group:       group,
				Resource:    resource,
				Action:      action,
			}); err != nil {
				common.Logger.Error("Error adding claim: %s", err.Error())
				return err
			}
		} else {
			// Backfill metadata on already-seeded claims that are missing it.
			if exists.Group == "" || exists.Description == "" {
				if err := db.UpdateClaimMetadata(ctx, claimID, description, group, resource, action); err != nil {
					common.Logger.Error("Error backfilling claim metadata for %s: %s", claimID, err.Error())
					return err
				}
			}
		}
	}

	_ = db.Disconnect(ctx)

	return nil
}
