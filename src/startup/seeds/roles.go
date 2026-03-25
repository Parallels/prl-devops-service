package seeds

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func SeedDefaultRoles() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	err := db.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	allSystemRoles := constants.AllSystemRoles

	for _, role := range allSystemRoles {
		description := constants.RoleDescriptionMap[role] // "" if not found
		if exists, _ := db.GetRole(ctx, role); exists == nil {
			if _, err := db.CreateRole(ctx, models.Role{
				ID:          role,
				Name:        role,
				Description: description,
				Internal:    true,
			}); err != nil {
				common.Logger.Error("Error adding role: %s", err.Error())
				return err
			}
		} else {
			// Backfill description on already-seeded roles that are missing it.
			if exists.Description == "" && description != "" {
				if err := db.UpdateRoleDescription(ctx, role, description); err != nil {
					common.Logger.Error("Error backfilling description for role %s: %s", role, err.Error())
					return err
				}
			}
		}
	}

	_ = db.Disconnect(ctx)

	return nil
}
