package seeds

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	cryptorand "github.com/cjlapao/common-go-cryptorand"
)

func SeedDefaultUsers() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	err := db.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	suRole, err := db.GetRole(ctx, constants.SUPER_USER_ROLE)
	if err != nil {
		return err
	}

	claims, err := db.GetClaims(ctx, "")
	if err != nil {
		return err
	}

	// Update root user with any missing claims
	if exists, _ := db.GetUser(ctx, "root"); exists != nil {
		userClaims := make(map[string]bool)
		for _, claim := range exists.Claims {
			userClaims[claim.ID] = true
		}

		needsUpdate := false
		for _, claim := range claims {
			if _, ok := userClaims[claim.ID]; !ok {
				exists.Claims = append(exists.Claims, claim)
				needsUpdate = true
			}
		}

		if needsUpdate {
			updateUser := *exists
			updateUser.Password = ""
			if err := db.UpdateUser(ctx, updateUser); err != nil {
				common.Logger.Error("Error updating root user claims: %s", err.Error())
				return err
			}
			common.Logger.Info("Updated root user with new claims")
		}

		_ = db.Disconnect(ctx)
		return nil
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
