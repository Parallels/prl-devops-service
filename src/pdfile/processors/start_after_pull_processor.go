package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type StartAfterPullCommandProcessor struct{}

func (p StartAfterPullCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "START_AFTER_PULL" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("run command is missing argument"))
	}

	dest.StartAfterPull = getBoolValue(command.Argument)
	ctx.LogDebugf("Processed by StartAfterPullCommandProcessor, line %v", line)
	return true, diag
}
