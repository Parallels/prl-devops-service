package seeds

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/service_provider"
)

func SeedDefaultRoles() error {
	db := service_provider.Get().JsonDatabase
	err := db.Connect()
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	defer db.Disconnect()

	if exists, _ := db.GetRole(constants.USER_ROLE); exists == nil {
		if err := db.CreateRole(&models.Role{
			ID:       constants.USER_ROLE,
			Name:     constants.USER_ROLE,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding role: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetRole(constants.SUPER_USER_ROLE); exists == nil {
		if err := db.CreateRole(&models.Role{
			ID:       constants.SUPER_USER_ROLE,
			Name:     constants.SUPER_USER_ROLE,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding role: %s", err.Error())
			return err
		}
	}

	db.Disconnect()

	return nil
}
