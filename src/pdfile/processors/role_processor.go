package processors

import (
	"errors"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type RoleCommandProcessor struct{}

func (p RoleCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "ROLE" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("run command is missing argument"))
	}

	rolesStr := command.Argument
	roles := strings.Split(rolesStr, ",")
	for i, role := range roles {
		roles[i] = strings.TrimSpace(role)
	}

	// removing duplicates
	dedupedRoles := make([]string, 0, len(roles))
	for _, role := range roles {
		found := false
		for _, dedupedRole := range dedupedRoles {
			if role == dedupedRole {
				found = true
				break
			}
		}
		if !found {
			dedupedRoles = append(dedupedRoles, role)
		}
	}

	dest.Roles = append(dest.Roles, dedupedRoles...)
	ctx.LogDebugf("Processed by RoleCommandProcessor, line %v", line)
	return true, diag
}
