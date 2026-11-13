package interfaces

// Sentinels relocated from data/users.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrUserNotFound              = errors.NewWithCode("user not found", 404)
	ErrUserAlreadyExists         = errors.NewWithCode("user already exists", 400)
	ErrUserEmailCannotBeEmpty    = errors.NewWithCode("user email cannot be empty", 400)
	ErrUserUsernameCannotBeEmpty = errors.NewWithCode("user username cannot be empty", 400)
	ErrUserNameCannotBeEmpty     = errors.NewWithCode("user name cannot be empty", 400)
	ErrUserIDCannotBeEmpty       = errors.NewWithCode("user ID cannot be empty", 400)
	ErrRoleIDCannotBeEmpty       = errors.NewWithCode("role ID cannot be empty", 400)
	ErrClaimIDCannotBeEmpty      = errors.NewWithCode("claim ID cannot be empty", 400)
	ErrCannotUpdateRootUser      = errors.NewWithCode("cannot update root user", 400)
	ErrCannotRemoveRootUser      = errors.NewWithCode("cannot remove root user", 400)
	ErrUserAlreadyContainsRole   = errors.NewWithCode("user already contains role", 400)
	ErrUserAlreadyContainsClaim  = errors.NewWithCode("user already contains claim", 400)
)
