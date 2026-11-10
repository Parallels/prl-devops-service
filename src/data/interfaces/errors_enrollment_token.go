package interfaces

// Sentinels relocated from data/orchestrator_enrollment_token.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrEnrollmentTokenNotFound = errors.NewWithCode("enrollment token not found", 404)
	ErrEnrollmentTokenExpired  = errors.NewWithCode("enrollment token has expired", 401)
	ErrEnrollmentTokenUsed     = errors.NewWithCode("enrollment token has already been used", 401)
)
