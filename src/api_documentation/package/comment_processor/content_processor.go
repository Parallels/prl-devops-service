package comment_processor

import (
	"log"
	"regexp"

	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var contentRegex = regexp.MustCompile(`^//\s*@Content\s+(.*)$`)

type ContentCommentProcessor struct{}

func (p *ContentCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename string, comment string) error {
	if matches := contentRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Content comment %s found in %s, processing", matches[1], filename)
			if endpoint.Content == nil {
				endpoint.Content = []string{}
			}
			endpoint.Content = append(endpoint.Content, matches[1])
		}
	}

	return nil
}
