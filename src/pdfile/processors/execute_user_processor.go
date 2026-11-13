package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type ExecuteUserCommandProcessor struct{}

func (p ExecuteUserCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "EXECUTE_USER" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("execute user is missing argument"))
	}

	dest.ExecuteUser = command.Argument
	ctx.LogDebugf("Processed by ExecuteUserCommandProcessor, line %v", line)
	return true, diag
}
