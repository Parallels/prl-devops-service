package interfaces

// Sentinels relocated from data/catalog.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrCatalogManifestNotFound = errors.NewWithCode("catalog manifest not found", 404)
	ErrCatalogAlreadyExists    = errors.NewWithCode("Catalog Manifest already exists", 404)
)
