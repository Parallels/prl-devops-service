package comment_processor

import (
	"log"
	"regexp"

	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var summaryRegex = regexp.MustCompile(`^//\s*@Summary\s+(.*)$`)

type SummaryCommentProcessor struct{}

func (p *SummaryCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename string, comment string) error {
	if matches := summaryRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Summary comment %s found in %s, processing", matches[1], filename)
			endpoint.Title = matches[1]
		}
	}

	return nil
}
