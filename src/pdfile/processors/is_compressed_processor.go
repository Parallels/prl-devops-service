package processors

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type IsCompressedCommandProcessor struct{}

func (p IsCompressedCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "IS_COMPRESSED" {
		return false, diag
	}
	if command.Argument == "" {
		command.Argument = "true"
	}

	dest.IsCompressed = getBoolValue(command.Argument)
	ctx.LogDebugf("Processed by IsCompressedCommandProcessor, line %v", line)
	return true, diag
}
