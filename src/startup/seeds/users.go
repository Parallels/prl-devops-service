package seeds

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/security/password"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	cryptorand "github.com/cjlapao/common-go-cryptorand"
)

func SeedDefaultUsers() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	passwordService := password.Get()
	cfg := config.Get()
	err := db.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	if exists, _ := db.GetUser(ctx, "root"); exists != nil {
		// if we have the environment variable for the envPassword, update the envPassword
		envPassword := cfg.RootPassword()
		if envPassword != "" && passwordService != nil {
			// if the hashed password is different from the existing password, update it
			if err := passwordService.Compare(envPassword, exists.ID, exists.Password); err != nil {
				if err := db.UpdateRootPassword(ctx, envPassword); err != nil {
					common.Logger.Error("Error updating root user password: %s", err.Error())
					return err
				}
				common.Logger.Info("Root user password updated successfully during booting due to password change detected during startup")
			}
		}
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

	defaultPassword, err := cryptorand.GetAlphaNumericRandomString(32)
	if err != nil {
		return err
	}

	if _, err := db.CreateUser(ctx, models.User{
		ID:       serviceprovider.Get().HardwareId,
		Name:     "Root",
		Username: "root",
		Email:    "root@localhost",
		Password: defaultPassword,
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
