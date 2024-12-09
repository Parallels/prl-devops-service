package comment_processor

import (
	"log"
	"regexp"

	"github.com/Parallels/prl-devops-service/api_documentation/package/helpers"
	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var claimsRegex = regexp.MustCompile(`^//\s*@Claims\s+(.*)$`)

type ClaimsCommentProcessor struct{}

func (p *ClaimsCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename string, comment string) error {
	if matches := claimsRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Claims comment %s found in %s, processing", matches[1], filename)
			if endpoint.Claims == nil {
				endpoint.Claims = []string{}
			}
			endpoint.Claims = append(endpoint.Claims, helpers.CleanString(matches[1]))
		}
	}

	return nil
}
