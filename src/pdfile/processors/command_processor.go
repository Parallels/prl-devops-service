package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type CommandCommandProcessor struct{}

var availableCommands = []string{"RUN", "DO", "IMPORT", "PUSH", "PULL", "LIST"}

func (p CommandCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	foundCommand := false
	cmd := ""
	for _, availableCommand := range availableCommands {
		if command.Command == availableCommand {
			foundCommand = true
			break
		}
	}
	if !foundCommand {
		return false, diag
	}
	if command.Command == "RUN" || command.Command == "DO" {
		cmd = command.Argument
		if command.Argument == "" {
			diag.AddError(errors.New("run command is missing argument"))
			return false, diag
		}
	} else {
		cmd = command.Command
	}

	dest.Command = cmd
	ctx.LogDebugf("Processed by CommandCommandProcessor, line %v", line)

	return true, diag
}
