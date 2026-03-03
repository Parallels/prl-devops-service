package seeds

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func SeedUsersMissingClaimsByRole() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	if err := db.Connect(ctx); err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}
	defer db.Disconnect(ctx)

	users, err := db.GetUsers(ctx, "")
	if err != nil {
		return err
	}

	allClaims, err := db.GetClaims(ctx, "")
	if err != nil {
		return err
	}

	claimsByID := make(map[string]models.Claim, len(allClaims))
	for _, claim := range allClaims {
		claimsByID[claim.ID] = claim
	}

	hasUpdates := false
	for _, user := range users {
		requiredClaims := getRoleBasedClaimSet(user.Roles)
		if len(requiredClaims) == 0 {
			continue
		}

		existingClaims := make(map[string]bool, len(user.Claims))
		for _, userClaim := range user.Claims {
			existingClaims[userClaim.ID] = true
		}

		userUpdated := false
		for claimID := range requiredClaims {
			if existingClaims[claimID] {
				continue
			}

			claim, ok := claimsByID[claimID]
			if !ok {
				common.Logger.Warn("Claim %s was expected from role mapping but does not exist", claimID)
				continue
			}

			user.Claims = append(user.Claims, claim)
			userUpdated = true
		}

		if !userUpdated {
			continue
		}

		updateUser := user
		updateUser.Password = ""
		if err := db.UpdateUser(ctx, updateUser); err != nil {
			common.Logger.Error("Error updating user %s with role-based claims: %s", user.Username, err.Error())
			return err
		}

		hasUpdates = true
		common.Logger.Info("Updated user %s with missing role-based claims", user.Username)
	}

	if hasUpdates {
		if err := db.SaveNow(ctx); err != nil {
			common.Logger.Error("Error saving database after role-based claims update: %s", err.Error())
			return err
		}
	}

	return nil
}

func getRoleBasedClaimSet(roles []models.Role) map[string]bool {
	claims := make(map[string]bool)

	for _, role := range roles {
		roleID := strings.ToUpper(role.ID)
		roleName := strings.ToUpper(role.Name)

		if roleID == constants.USER_ROLE || roleName == constants.USER_ROLE {
			for _, claim := range constants.DefaultClaims {
				claims[claim] = true
			}
		}

		if roleID == constants.SUPER_USER_ROLE || roleName == constants.SUPER_USER_ROLE {
			for _, claim := range constants.AllSuperUserClaims {
				claims[claim] = true
			}
		}
	}

	return claims
}
