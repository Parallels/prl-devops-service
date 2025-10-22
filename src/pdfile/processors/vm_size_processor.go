package processors

import (
	"errors"
	"strconv"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type VmSizeCommandProcessor struct{}

func (p VmSizeCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "VM_SIZE" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("VM_SIZE command is missing argument"))
	}

	vmSize, err := strconv.Atoi(command.Argument)
	if err != nil || vmSize <= 0 {
		diag.AddError(errors.New("VM_SIZE command has invalid argument, it should be a positive integer representing size in MB"))
		return false, diag
	}

	dest.VMSize = int64(vmSize)

	ctx.LogDebugf("Processed by VmSizeCommandProcessor, line %v", line)
	return true, diag
}
