package interfaces

// Sentinels relocated from data/packer_template.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrPackerTemplateNotFound         = errors.NewWithCode("packer template not found", 404)
	ErrPackerTemplateAlreadyExists    = errors.NewWithCode("machine template already exists", 400)
	ErrRemovingInternalPackerTemplate = errors.NewWithCode("cannot remove internal machine template", 400)
	ErrUpdatingInternalPackerTemplate = errors.NewWithCode("cannot update internal machine template", 400)
)
