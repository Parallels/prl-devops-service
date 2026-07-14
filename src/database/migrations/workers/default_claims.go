package workers

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/stores"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"
	"gorm.io/gorm"
)

// DefaultClaimsWorker seeds default system claims
type DefaultClaimsWorker struct {
	db *gorm.DB
}

func NewDefaultClaimsWorker(db *gorm.DB) *DefaultClaimsWorker {
	return &DefaultClaimsWorker{db: db}
}

func (w *DefaultClaimsWorker) GetName() string {
	return "default-claims"
}

func (w *DefaultClaimsWorker) GetDescription() string {
	return "Seeds all default system claims from constants"
}

func (w *DefaultClaimsWorker) GetOrder() int {
	return 10 // Run early, before roles
}

func (w *DefaultClaimsWorker) Run(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("default_claims_migration")
	logger := logging.Get()

	// Initialize claim store with BaseDataStore pattern
	claimStore := &stores.ClaimDataStore{
		BaseDataStore: *common.NewBaseDataStore(w.db),
	}

	logger.Info("Seeding default claims...")

	allSystemClaims := constants.AllSystemClaims
	created := 0
	skipped := 0

	for _, claimID := range allSystemClaims {
		// Check if claim already exists
		existing, getDiag := claimStore.GetClaimByNameOrID(ctx, claimID)
		if getDiag.HasErrors() {
			diag.Append(getDiag)
			continue
		}

		if existing != nil {
			logger.Debug("Claim '%s' already exists, skipping", claimID)
			skipped++
			continue
		}

		// Get metadata
		meta, hasMeta := constants.ClaimCategoryMap[claimID]
		group := constants.ClaimGroupCustom
		resource := ""
		action := ""
		if hasMeta {
			group = meta.Group
			resource = meta.Resource
			action = meta.Action
		}

		description := constants.ClaimDescriptionMap[claimID]

		// Create claim
		claim := &models.Claim{
			Name:        claimID,
			Internal:    true,
			Description: description,
			Group:       group,
			Resource:    resource,
			Action:      action,
		}

		_, createDiag := claimStore.CreateClaim(ctx, claim)
		if createDiag.HasErrors() {
			logger.Error("Failed to create claim '%s': %v", claimID, createDiag.GetSummary())
			diag.Append(createDiag)
			continue
		}

		logger.Debug("Created claim: %s", claimID)
		created++
	}

	logger.Info("Claims seeding completed: %d created, %d skipped, %d total", created, skipped, len(allSystemClaims))

	return diag
}

func (w *DefaultClaimsWorker) Rollback(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	// Rollback not supported for claims (they might be referenced by roles/users)
	return nil
}
