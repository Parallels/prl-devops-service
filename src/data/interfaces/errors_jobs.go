package interfaces

// Sentinels relocated from data/jobs.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrJobNotFound = errors.NewWithCode("job not found", 404)
)
