package interfaces

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
)

// MigrationWorker interface defines the contract for a migration worker
type MigrationWorker interface {
	GetName() string
	GetDescription() string
	GetVersion() int
	GetOrder() int
	Up(ctx basecontext.BaseContext) *errors.Diagnostics
	Down(ctx basecontext.BaseContext) *errors.Diagnostics
}
