package seeds

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/serviceprovider"
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
		if exists, _ := db.GetRole(ctx, role); exists == nil {
			if _, err := db.CreateRole(ctx, models.Role{
				ID:       role,
				Name:     role,
				Internal: true,
			}); err != nil {
				common.Logger.Error("Error adding role: %s", err.Error())
				return err
			}
		} else {
			ctx.LogDebugf("Role already exists: %s", role)
		}
	}

	_ = db.Disconnect(ctx)

	return nil
}
