package processors

import (
	"errors"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type ClaimCommandProcessor struct{}

func (p ClaimCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "CLAIM" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("run command is missing argument"))
	}

	claimStr := command.Argument
	claims := strings.Split(claimStr, ",")
	for i, claim := range claims {
		claims[i] = strings.TrimSpace(claim)
	}

	// removing duplicates
	dedupedClaims := make([]string, 0, len(claims))
	for _, claim := range claims {
		found := false
		for _, dedupedClaim := range dedupedClaims {
			if claim == dedupedClaim {
				found = true
				break
			}
		}
		if !found {
			dedupedClaims = append(dedupedClaims, claim)
		}
	}

	dest.Claims = append(dest.Claims, dedupedClaims...)
	ctx.LogDebugf("Processed by ClaimCommandProcessor, line %v", line)
	return true, diag
}
