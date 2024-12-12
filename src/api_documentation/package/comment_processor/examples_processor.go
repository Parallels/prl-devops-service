package comment_processor

import (
	"fmt"
	"log"
	"regexp"

	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var examplesRegex = regexp.MustCompile(`^//\s+@Examples\s+(.*)$`)

type ExamplesCommentProcessor struct{}

func (p *ExamplesCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename, comment string) error {
	if matches := examplesRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Examples comment %v found in %s, processing", matches[1], filename)
			if endpoint.ExampleRequestPayload == nil {
				endpoint.ExampleRequestPayload = []string{}
			}
			exampleLine := fmt.Sprintf("%v\n", matches[1])
			endpoint.ExampleRequestPayload = append(endpoint.ExampleRequestPayload, exampleLine)
		}
	}

	return nil
}
