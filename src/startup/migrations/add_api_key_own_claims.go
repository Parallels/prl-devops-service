package migrations

import (
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
)

type AddApiKeyOwnClaims struct{}

func (v AddApiKeyOwnClaims) Apply() error {
	versionTarget := "0.7.0"
	svc, err := Init()
	if err != nil {
		return err
	}

	compareResult, err := compareVersions(svc.schemaVersion, versionTarget)
	if err != nil {
		svc.Context.LogErrorf("Error comparing versions: %s", err.Error())
		return err
	}

	if compareResult == VersionEqualToTarget || compareResult == VersionHigherThanTarget {
		svc.Context.LogDebugf("Schema version is already %s, skipping migration", versionTarget)
		return nil
	}

	svc.Context.LogInfof("Applying migration to version %s", versionTarget)

	users, err := svc.DbService.GetUsers(svc.Context, "")
	if err != nil {
		return err
	}

	ownClaims := []string{
		constants.LIST_OWN_API_KEY_CLAIM,
		constants.CREATE_OWN_API_KEY_CLAIM,
		constants.DELETE_OWN_API_KEY_CLAIM,
		constants.UPDATE_OWN_API_KEY_CLAIM,
	}

	for _, user := range users {
		roleNames := make(map[string]bool)
		for _, role := range user.Roles {
			roleNames[role.Name] = true
		}

		if roleNames[constants.USER_ROLE] && !roleNames[constants.SUPER_USER_ROLE] && !roleNames[constants.ADMIN_ROLE] {
			userClaimIDs := make(map[string]bool)
			for _, claim := range user.Claims {
				userClaimIDs[claim.ID] = true
			}

			needsUpdate := false
			for _, claimName := range ownClaims {
				if !userClaimIDs[claimName] {
					err := svc.DbService.AddClaimToUser(svc.Context, user.ID, claimName)
					if err == nil {
						svc.Context.LogInfof("Added claim %s to user %s", claimName, user.Username)
						needsUpdate = true
					} else if err != data.ErrUserAlreadyContainsClaim {
						svc.Context.LogErrorf("Error adding claim %s to user %s: %s", claimName, user.Username, err.Error())
						return err
					}
				}
			}

			if needsUpdate {
				if err := svc.DbService.SaveNow(svc.Context); err != nil {
					svc.Context.LogErrorf("Error saving database after adding claims to user %s: %s", user.Username, err.Error())
					return err
				}
			}
		}
	}

	err = svc.DbService.UpdateSchemaVersion(svc.Context, versionTarget)
	if err != nil {
		return err
	}

	return nil
}
