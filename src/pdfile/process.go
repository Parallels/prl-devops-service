package pdfile

import (
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

const (
	RunCommand = "RUN"
)

func Process(ctx basecontext.ApiContext, fileContent string) (*models.PDFile, *diagnostics.PDFileDiagnostics) {
	result := models.NewPdFile()
	result.Raw = strings.Split(fileContent, "\n")
	diag := diagnostics.NewPDFileDiagnostics()
	svc := NewPDFileService(ctx, result)

	for i, line := range result.Raw {
		executed := false
		for _, processor := range svc.processors {
			processExecuted, processDiag := processor.Process(ctx, line, result)
			if processDiag.HasErrors() {
				diag.Append(processDiag)
			}
			if processExecuted {
				executed = true
				break
			}
		}

		if !executed {
			diag.AddError(fmt.Errorf("invalid command %v at line %d", line, i))
		}
	}

	if result.Command == "" {
		if result.From != "" {
			result.Command = "pull"
		}
		if result.To != "" {
			result.Command = "push"
		}
	}

	return result, diag
}
