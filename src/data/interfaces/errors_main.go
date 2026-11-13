package interfaces

// Sentinels relocated from data/main.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrDatabaseNotConnected = errors.NewWithCode("the database is not connected", 500)
	ErrNotAuthorized        = errors.NewWithCode("not authorized to view record", 403)
)
