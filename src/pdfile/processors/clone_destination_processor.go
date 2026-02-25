package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type CloneDestinationCommandProcessor struct{}

func (p CloneDestinationCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "CLONE_DESTINATION" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("CLONE_DESTINATION requires an argument"))
		return false, diag
	}
	dest.CloneToDestination = command.Argument

	ctx.LogDebugf("Processed by CloneDestinationCommandProcessor, line %v", line)
	return true, diag
}
