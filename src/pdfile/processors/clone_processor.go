package processors

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
	"github.com/Parallels/prl-devops-service/security"
)

type CloneCommandProcessor struct{}

func (p CloneCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "CLONE" {
		return false, diag
	}
	if command.Argument == "" {
		cloneName, err := security.GenerateCryptoRandomString(20)
		if err != nil {
			diag.AddError(err)
			return false, diag
		}
		dest.CloneTo = cloneName
	}

	dest.Clone = true
	dest.CloneTo = command.Argument

	ctx.LogDebugf("Processed by CloneCommandProcessor, line %v", line)
	return true, diag
}
