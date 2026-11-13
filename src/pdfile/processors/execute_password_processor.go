package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type ExecutePasswordCommandProcessor struct{}

func (p ExecutePasswordCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "EXECUTE_PASSWORD" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("execute password is missing argument"))
	}

	dest.ExecutePassword = command.Argument
	ctx.LogDebugf("Processed by ExecutePasswordCommandProcessor, line %v", line)
	return true, diag
}
