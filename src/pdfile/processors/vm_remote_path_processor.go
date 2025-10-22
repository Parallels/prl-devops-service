package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type VmRemotePathCommandProcessor struct{}

func (p VmRemotePathCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "VM_REMOTE_PATH" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("VM_REMOTE_PATH command is missing argument"))
	}

	dest.VMRemotePath = command.Argument

	ctx.LogDebugf("Processed by VmRemotePathCommandProcessor, line %v", line)
	return true, diag
}
