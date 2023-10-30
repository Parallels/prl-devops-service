package seeds

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/service_provider"
)

func SeedDefaultUsers() error {
	db := service_provider.Get().JsonDatabase
	err := db.Connect()
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	defer db.Disconnect()

	if exists, _ := db.GetUser("root"); exists != nil {
		return nil
	}

	suRole, err := db.GetRole(constants.SUPER_USER_ROLE)
	if err != nil {
		return err
	}
	claims, err := db.GetClaims()
	if err != nil {
		return err
	}

	if _, err := db.CreateUser(&models.User{
		ID:       service_provider.Get().HardwareId,
		Name:     "Root",
		Username: "root",
		Email:    "root@localhost",
		Password: service_provider.Get().HardwareSecret,
		Roles: []models.Role{
			*suRole,
		},
		Claims: claims,
	}); err != nil {
		common.Logger.Error("Error adding root user: %s", err.Error())
		return err
	}

	db.Disconnect()

	return nil
}
