package processors

import (
	"errors"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type ProviderCommandProcessor struct{}

func (p ProviderCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "PROVIDER" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("run command is missing argument"))
	}

	if dest.Provider == nil {
		dest.Provider = &models.PDFileProvider{}
	}

	if strings.HasPrefix(command.Argument, "NAME") {
		parts := strings.Split(command.Argument, "=")
		if len(parts) != 2 {
			parts = strings.Split(command.Argument, " ")
		}
		if len(parts) != 2 {
			diag.AddError(errors.New("Provider name is missing argument"))
		}
		dest.Provider.Name = parts[1]
	} else {
		provider, err := dest.ParseProvider(line)
		if err != nil {
			diag.AddError(err)
		}

		if dest.Provider.Attributes == nil {
			dest.Provider.Attributes = make(map[string]string)
		}

		if provider.Name != "" {
			dest.Provider.Name = provider.Name
		}
		for key, value := range provider.Attributes {
			dest.Provider.Attributes[key] = value
		}
	}

	ctx.LogDebugf("Processed by ProviderCommandProcessor, line %v", line)
	return true, diag
}
