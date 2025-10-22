package processors

import (
	"errors"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type CompressPackLevelCommandProcessor struct{}

func (p CompressPackLevelCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "COMPRESS_PACK_LEVEL" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("compress pack level command is missing argument"))
	}

	compressLevel, err := helpers.ConvertCompressRatioFromString(strings.ToLower(command.Argument))
	if err != nil {
		diag.AddError(errors.New("compress pack level command has invalid argument, allowed values are 'best_speed', 'balanced', 'best_compression', 'default', 'no_compression'"))
		return false, diag
	}

	dest.CompressPackLevel = compressLevel

	ctx.LogDebugf("Processed by CompressPackLevelCommandProcessor, line %v", line)
	return true, diag
}
