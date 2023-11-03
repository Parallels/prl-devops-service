package seeds

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/serviceprovider"
)

func SeedDefaultRoles() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	err := db.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	defer db.Disconnect(ctx)

	allSystemRoles := constants.AllSystemRoles

	for _, role := range allSystemRoles {
		if exists, _ := db.GetRole(ctx, role); exists == nil {
			if err := db.CreateRole(ctx, models.Role{
				ID:       role,
				Name:     role,
				Internal: true,
			}); err != nil {
				common.Logger.Error("Error adding role: %s", err.Error())
				return err
			}
		} else {
			ctx.LogDebug("Role already exists: %s", role)
		}
	}

	db.Disconnect(ctx)

	return nil
}
