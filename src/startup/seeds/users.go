package seeds

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/serviceprovider"
)

func SeedDefaultUsers() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	err := db.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	if exists, _ := db.GetUser(ctx, "root"); exists != nil {
		return nil
	}

	suRole, err := db.GetRole(ctx, constants.SUPER_USER_ROLE)
	if err != nil {
		return err
	}

	claims, err := db.GetClaims(ctx, "")
	if err != nil {
		return err
	}

	if _, err := db.CreateUser(ctx, models.User{
		ID:       serviceprovider.Get().HardwareId,
		Name:     "Root",
		Username: "root",
		Email:    "root@localhost",
		Password: serviceprovider.Get().HardwareSecret,
		Roles: []models.Role{
			*suRole,
		},
		Claims: claims,
	}); err != nil {
		common.Logger.Error("Error adding root user: %s", err.Error())
		return err
	}

	_ = db.Disconnect(ctx)

	return nil
}
