package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type ExecuteCommandProcessor struct{}

func (p ExecuteCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "EXECUTE" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("run command is missing argument"))
	}

	if dest.Execute == nil {
		dest.Execute = make([]string, 0)
	}

	found := false
	for _, existingCommand := range dest.Execute {
		if existingCommand == command.Argument {
			found = true
			break
		}
	}
	if !found {
		dest.Execute = append(dest.Execute, command.Argument)
	}

	ctx.LogDebugf("Processed by ExecuteCommandProcessor, line %v", line)
	return true, diag
}
