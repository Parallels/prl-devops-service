package workers

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/stores"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"
	"gorm.io/gorm"
)

// RoleClaimsWorker assigns claims to roles based on RoleClaimsMap
type RoleClaimsWorker struct {
	db *gorm.DB
}

func NewRoleClaimsWorker(db *gorm.DB) *RoleClaimsWorker {
	return &RoleClaimsWorker{db: db}
}

func (w *RoleClaimsWorker) GetName() string {
	return "role-claims-associations"
}

func (w *RoleClaimsWorker) GetDescription() string {
	return "Assigns claims to roles based on RoleClaimsMap"
}

func (w *RoleClaimsWorker) GetOrder() int {
	return 30 // Run after roles and claims
}

func (w *RoleClaimsWorker) Run(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("role_claims_migration")
	logger := logging.Get()

	// Initialize stores with BaseDataStore pattern
	roleStore := &stores.RoleDataStore{
		BaseDataStore: *common.NewBaseDataStore(w.db),
	}

	claimStore := &stores.ClaimDataStore{
		BaseDataStore: *common.NewBaseDataStore(w.db),
	}

	logger.Info("Assigning claims to roles...")

	totalAssigned := 0

	// Iterate through RoleClaimsMap from constants
	for roleName, claimIDs := range constants.RoleClaimsMap {
		logger.Info("Processing role: %s (%d claims)", roleName, len(claimIDs))

		// Get the role
		role, getRoleDiag := roleStore.GetRoleBySlugOrID(ctx, roleName)
		if getRoleDiag.HasErrors() {
			logger.Error("Failed to get role '%s': %v", roleName, getRoleDiag.GetSummary())
			diag.Append(getRoleDiag)
			continue
		}

		if role == nil {
			logger.Warn("Role '%s' not found, skipping claim assignments", roleName)
			continue
		}

		// Get existing claims for this role
		existingClaims, getClaimsDiag := roleStore.GetRoleClaims(ctx, role.ID)
		if getClaimsDiag.HasErrors() {
			logger.Error("Failed to get existing claims for role '%s': %v", roleName, getClaimsDiag.GetSummary())
			diag.Append(getClaimsDiag)
			continue
		}

		// Build map of existing claim IDs
		existingClaimIDs := make(map[string]bool)
		for _, claim := range existingClaims {
			existingClaimIDs[claim.ID] = true
		}

		assigned := 0
		skipped := 0

		// Assign each claim to the role
		for _, claimID := range claimIDs {
			// Skip if already assigned
			if existingClaimIDs[claimID] {
				skipped++
				continue
			}

			// Get the claim
			claim, getClaimDiag := claimStore.GetClaimByNameOrID(ctx, claimID)
			if getClaimDiag.HasErrors() {
				logger.Error("Failed to get claim '%s': %v", claimID, getClaimDiag.GetSummary())
				continue
			}

			if claim == nil {
				logger.Warn("Claim '%s' not found, skipping", claimID)
				continue
			}

			// Add claim to role
			addDiag := roleStore.AddClaimToRole(ctx, role.ID, claim.ID)
			if addDiag.HasErrors() {
				logger.Error("Failed to add claim '%s' to role '%s': %v", claimID, roleName, addDiag.GetSummary())
				diag.Append(addDiag)
				continue
			}

			assigned++
		}

		logger.Info("Role '%s': %d claims assigned, %d skipped", roleName, assigned, skipped)
		totalAssigned += assigned
	}

	logger.Info("Role-claims association completed: %d total claims assigned", totalAssigned)

	return diag
}

func (w *RoleClaimsWorker) Rollback(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	// Rollback not supported
	return nil
}
