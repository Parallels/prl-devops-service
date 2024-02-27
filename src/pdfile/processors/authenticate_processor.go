package processors

import (
	"errors"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type AuthenticateCommandProcessor struct{}

func (p AuthenticateCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "AUTHENTICATE" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("run command is missing argument"))
	}

	if dest.Authentication == nil {
		dest.Authentication = &models.PDFileAuthentication{}
	}

	argumentParts := strings.Split(command.Argument, " ")
	switch strings.ToUpper(argumentParts[0]) {
	case "USERNAME":
		dest.Authentication.Username = strings.TrimSpace(strings.Join(argumentParts[1:], " "))
	case "PASSWORD":
		dest.Authentication.Password = strings.TrimSpace(strings.Join(argumentParts[1:], " "))
	case "API_KEY":
		dest.Authentication.ApiKey = strings.TrimSpace(strings.Join(argumentParts[1:], " "))
	default:
		diag.AddError(errors.New("invalid authentication type"))
	}
	ctx.LogDebugf("Processed by AuthenticateCommandProcessor, line %v", line)
	return true, diag
}
