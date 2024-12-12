package comment_processor

import (
	"log"
	"regexp"

	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var tagsRegex = regexp.MustCompile(`^//\s*@Tags\s+(.*)$`)

type TagsCommentProcessor struct{}

func (p *TagsCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename string, comment string) error {
	if matches := tagsRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Tags comment %s found in %s, processing", matches[1], filename)
			endpoint.Category = matches[1]
		}
	}

	return nil
}
