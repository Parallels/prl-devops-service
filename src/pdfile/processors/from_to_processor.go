package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type FromToCommandProcessor struct{}

const (
	FromCommand = "FROM"
	ToCommand   = "TO"
)

func (p FromToCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != FromCommand && command.Command != ToCommand {
		return false, diag
	}

	if command.Command == FromCommand {
		dest.Command = "pull"
	}
	if command.Command == ToCommand {
		dest.Command = "push"
	}

	if command.Argument == "" {
		diag.AddError(errors.New("run command is missing argument"))
	}

	dest.Host = command.Argument
	if command.Command == FromCommand {
		dest.From = command.Argument
	}
	if command.Command == ToCommand {
		dest.To = command.Argument
	}

	ctx.LogDebugf("Processed by FromToCommandProcessor, line %v", line)
	return true, diag
}
