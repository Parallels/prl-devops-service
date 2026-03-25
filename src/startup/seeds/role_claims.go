package seeds

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

// SeedDefaultRoleClaims idempotently ensures that each built-in role has its
// canonical claims attached. It is safe to run on every startup: claims already
// present on a role are silently skipped.
func SeedDefaultRoleClaims() error {
	ctx := basecontext.NewRootBaseContext()
	db := serviceprovider.Get().JsonDatabase
	if err := db.Connect(ctx); err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}
	defer db.Disconnect(ctx)

	hasUpdates := false

	for roleName, claimNames := range constants.RoleClaimsMap {
		role, err := db.GetRole(ctx, roleName)
		if err != nil {
			common.Logger.Warn("Role %s not found during role-claims seeding, skipping", roleName)
			continue
		}

		for _, claimName := range claimNames {
			err := db.AddClaimToRole(ctx, role.ID, claimName)
			if err == nil {
				hasUpdates = true
				continue
			}
			// ErrRoleAlreadyContainsClaim is expected on re-runs — not a failure.
			if err == data.ErrRoleAlreadyContainsClaim {
				continue
			}
			// Claim doesn't exist yet (shouldn't happen if SeedDefaultClaims ran first).
			if err == data.ErrClaimNotFound {
				common.Logger.Warn("Claim %s not found while seeding role %s, skipping", claimName, roleName)
				continue
			}
			common.Logger.Error("Error adding claim %s to role %s: %s", claimName, roleName, err.Error())
			return err
		}
	}

	if hasUpdates {
		if err := db.SaveNow(ctx); err != nil {
			common.Logger.Error("Error saving database after role-claims seeding: %s", err.Error())
			return err
		}
	}

	return nil
}
