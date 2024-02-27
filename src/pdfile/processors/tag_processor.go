package processors

import (
	"errors"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type TagCommandProcessor struct{}

func (p TagCommandProcessor) Process(ctx basecontext.ApiContext, line string, dest *models.PDFile) (bool, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()
	command := getCommand(line)
	if command == nil {
		return false, diag
	}
	if command.Command != "TAG" {
		return false, diag
	}
	if command.Argument == "" {
		diag.AddError(errors.New("run command is missing argument"))
	}

	tagStr := command.Argument
	tags := strings.Split(tagStr, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	// removing duplicates
	dedupedTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		found := false
		for _, dedupedTag := range dedupedTags {
			if tag == dedupedTag {
				found = true
				break
			}
		}
		if !found {
			dedupedTags = append(dedupedTags, tag)
		}
	}

	dest.Tags = append(dest.Tags, dedupedTags...)
	ctx.LogDebugf("Processed by TagCommandProcessor, line %v", line)
	return true, diag
}
