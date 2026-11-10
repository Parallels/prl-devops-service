package interfaces

// Sentinels relocated from data/claims.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrClaimEmptyNameOrId  = errors.NewWithCode("no claim specified", 500)
	ErrClaimEmptyName      = errors.NewWithCode("claim name cannot be empty", 500)
	ErrClaimNotFound       = errors.NewWithCode("claim not found", 404)
	ErrRemoveInternalClaim = errors.NewWithCode("claim is internal and cannot be removed", 400)
	ErrUpdateInternalClaim = errors.NewWithCode("claim is internal and cannot be updated", 400)
	ErrClaimHasUsers       = errors.NewWithCode("claim has users assigned and cannot be removed", 400)
)
