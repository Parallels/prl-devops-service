package interfaces

// Sentinels relocated from data/catalog_managers.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"errors"
)

var (
	ErrCatalogManagerNotFound       = errors.New("catalog manager not found")
	ErrRemoveInternalCatalogManager = errors.New("catalog manager is internal and cannot be removed")
	ErrUpdateInternalCatalogManager = errors.New("catalog manager is internal and cannot be updated")
)
