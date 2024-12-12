package comment_processor

import (
	"log"
	"regexp"

	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var descriptionRegex = regexp.MustCompile(`^//\s*@Description\s+(.*)$`)

type DescriptionCommentProcessor struct{}

func (p *DescriptionCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename string, comment string) error {
	if matches := descriptionRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Description comment %s found in %s, processing", matches[1], filename)
			endpoint.Description = matches[1]
		}
	}

	return nil
}
