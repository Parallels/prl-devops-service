package comment_processor

import (
	"log"
	"regexp"

	"github.com/Parallels/prl-devops-service/api_documentation/package/helpers"
	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var rolesRegex = regexp.MustCompile(`^//\s*@Roles\s+(.*)$`)

type RolesCommentProcessor struct{}

func (p *RolesCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename string, comment string) error {
	if matches := rolesRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Roles comment %s found in %s, processing", matches[1], filename)
			if endpoint.Roles == nil {
				endpoint.Roles = []string{}
			}
			endpoint.Roles = append(endpoint.Roles, helpers.CleanString(matches[1]))
		}
	}

	return nil
}
