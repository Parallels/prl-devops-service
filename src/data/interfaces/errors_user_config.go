package interfaces

// Sentinels relocated from data/user_config.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrUserConfigNotFound      = errors.NewWithCode("user config not found", 404)
	ErrUserConfigAlreadyExists = errors.NewWithCode("user config already exists", 500)
)
