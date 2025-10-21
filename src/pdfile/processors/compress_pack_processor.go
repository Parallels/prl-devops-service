package processors

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type CompressPackCommandProcessor struct{}

func (p CompressPackCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "COMPRESS_PACK" {
		return false, diag
	}
	if command.Argument == "" {
		command.Argument = "true"
	}

	dest.CompressPack = getBoolValue(command.Argument)
	ctx.LogDebugf("Processed by CompressPackCommandProcessor, line %v", line)
	return true, diag
}
