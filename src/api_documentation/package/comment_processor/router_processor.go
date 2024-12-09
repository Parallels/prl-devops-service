package comment_processor

import (
	"log"
	"regexp"

	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var routerRegex = regexp.MustCompile(`^//\s*@Router\s+(.*)\s+\[(.*)\]$`)

type RouterCommentProcessor struct{}

func (p *RouterCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename string, comment string) error {
	if matches := routerRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Router comment %s found in %s, processing", matches[1], filename)
		}
		if len(matches) >= 2 {
			endpoint.Path = matches[1]
		}
		if len(matches) >= 3 {
			endpoint.Method = matches[2]
		}
	}

	return nil
}
