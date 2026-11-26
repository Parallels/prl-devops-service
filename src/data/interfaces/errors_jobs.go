package interfaces

// Sentinels relocated from data/jobs.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrJobNotFound = errors.NewWithCode("job not found", 404)

	// ErrNodeIdentityUnusable refuses a node-identity rename that could only be
	// a mistake. An empty side would match the ownerless rows — the ones
	// recovery deliberately adopts — and rewrite them as someone's work, so the
	// case is refused rather than treated as a no-op.
	ErrNodeIdentityUnusable = errors.NewWithCode("node identity is empty or unchanged", 400)
)
