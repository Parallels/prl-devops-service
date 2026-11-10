package interfaces

// Sentinels relocated from data/orchestrator.go (PBI #421 D-3): shared values so both engines
// return identical pointers; data/ keeps identity-preserving aliases.
import (
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	ErrOrchestratorHostEmptyIdOrHost          = errors.NewWithCode("no host specified", 500)
	ErrOrchestratorHostEmptyName              = errors.NewWithCode("host name cannot be empty", 500)
	ErrOrchestratorHostNotFound               = errors.NewWithCode("host not found", 404)
	ErrOrchestratorReverseProxyHostNotFound   = errors.NewWithCode("reverse proxy host not found", 404)
	ErrOrchestratorHostVirtualMachineNotFound = errors.NewWithCode("host virtual machine not found", 404)
	ErrOrchestratorHostsVMSnapshotsNotFound   = errors.NewWithCode("VM snapshots not found for host", 404)
)
