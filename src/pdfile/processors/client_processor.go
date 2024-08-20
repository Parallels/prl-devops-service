package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type ClientCommandProcessor struct{}

func (p ClientCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "CLIENT" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("client command is missing argument"))
	}

	dest.Client = command.Argument
	ctx.LogDebugf("Processed by ClientCommandProcessor, line %v", line)
	return true, diag
}
