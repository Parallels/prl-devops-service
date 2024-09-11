package processors

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type MinimumSpecsRequirementsCommandProcessor struct{}

func (p MinimumSpecsRequirementsCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "MINIMUM_REQUIREMENT" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("requirements command is missing argument"))
	}

	if dest.MinimumSpecRequirements == nil {
		dest.MinimumSpecRequirements = &models.PdFileMinimumSpecRequirement{}
	}

	argumentParts := strings.Split(command.Argument, " ")
	switch strings.ToUpper(argumentParts[0]) {
	case "CPU":
		cpuString := strings.TrimSpace(strings.Join(argumentParts[1:], " "))
		cpu, err := strconv.Atoi(cpuString)
		if err != nil {
			diag.AddError(errors.New("invalid cpu minimum requirement value"))
		} else {
			dest.MinimumSpecRequirements.Cpu = cpu
		}
	case "MEMORY":
		memoryString := strings.TrimSpace(strings.Join(argumentParts[1:], " "))
		memory, err := strconv.Atoi(memoryString)
		if err != nil {
			diag.AddError(errors.New("invalid memory minimum requirement value"))
		} else {
			dest.MinimumSpecRequirements.Memory = memory
		}
	case "DISK":
		diskString := strings.TrimSpace(strings.Join(argumentParts[1:], " "))
		disk, err := strconv.Atoi(diskString)
		if err != nil {
			diag.AddError(errors.New("invalid disk minimum requirement value"))
		} else {
			dest.MinimumSpecRequirements.Disk = disk
		}
	default:
		diag.AddError(errors.New("invalid minimum requirement type"))
	}
	ctx.LogDebugf("Processed by MinimumSpecsRequirementsCommandProcessor, line %v", line)
	return true, diag
}
