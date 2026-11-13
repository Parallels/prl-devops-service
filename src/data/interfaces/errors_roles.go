package interfaces

// Sentinels relocated from data/roles.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrRoleEmptyNameOrId        = errors.NewWithCode("no role specified", 500)
	ErrRoleEmptyName            = errors.NewWithCode("role name cannot be empty", 500)
	ErrRoleNotFound             = errors.NewWithCode("role not found", 404)
	ErrRemoveInternalRole       = errors.NewWithCode("role is internal and cannot be removed", 400)
	ErrUpdateInternalRole       = errors.NewWithCode("role is internal and cannot be updated", 400)
	ErrRoleAlreadyContainsClaim = errors.NewWithCode("role already contains claim", 400)
	ErrRoleDoesNotContainClaim  = errors.NewWithCode("role does not contain claim", 404)
	ErrRoleHasUsers             = errors.NewWithCode("role has users assigned and cannot be removed", 400)
)
