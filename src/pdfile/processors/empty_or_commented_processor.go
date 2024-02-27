package processors

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type EmptyOrCommentedCommandProcessor struct{}

func (p EmptyOrCommentedCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	if line == "" || strings.HasPrefix(line, "#") {
		return true, diag
	}

	if strings.HasPrefix(line, "#") {
		ctx.LogDebugf("Ignored line %v", line)
	}
	return false, diag
}
