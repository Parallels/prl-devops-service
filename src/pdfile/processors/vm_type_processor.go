package processors

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type VmTypeCommandProcessor struct{}

func (p VmTypeCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "VM_TYPE" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("VM_TYPE command is missing argument"))
	}

	switch command.Argument {
	case "pvm":
		dest.VMType = "pvm"
	case "macvm":
		dest.VMType = "macvm"
	default:
		diag.AddError(errors.New("VM_TYPE command has invalid argument, allowed values are 'pvm' or 'macvm'"))
		return false, diag
	}

	ctx.LogDebugf("Processed by VmTypeCommandProcessor, line %v", line)
	return true, diag
}
