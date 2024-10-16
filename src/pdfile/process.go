package pdfile

import (
	"fmt"
	"os"
	"regexp"
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
	var envVars []string
	for _, line := range result.Raw {
		start := 0
		for {
			line = line[start:]
			start = strings.Index(line, "{{")
			if start == -1 {
				break
			}
			end := strings.Index(line[start:], "}}")
			if end == -1 {
				break
			}
			envVar := line[start : start+end+2]
			envVars = append(envVars, envVar)
			start += end + 2
		}
	}
	for _, envVar := range envVars {
		envVarName := strings.ReplaceAll(envVar, "{{", "")
		envVarName = strings.ReplaceAll(envVarName, "{{", "")
		envVarName = strings.ReplaceAll(envVarName, "}}", "")
		re := regexp.MustCompile(`(?i)\.env\.`)
		envVarName = re.ReplaceAllString(envVarName, "")

		envVarName = strings.TrimSpace(envVarName)
		if _, exists := os.LookupEnv(envVarName); exists {
			fileContent = strings.ReplaceAll(fileContent, envVar, os.Getenv(envVarName))
		}
	}
	result.Raw = strings.Split(fileContent, "\n")

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
