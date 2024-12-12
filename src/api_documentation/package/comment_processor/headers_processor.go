package comment_processor

import (
	"log"
	"regexp"
	"strings"

	"github.com/Parallels/prl-devops-service/api_documentation/package/helpers"
	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var headerRegex = regexp.MustCompile(`^//\s*@HeaderParam\s+(.*)$`)

type HeaderCommentProcessor struct{}

func (p *HeaderCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename, comment string) error {
	if matches := headerRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Header comment %v found in %s, processing", matches[1], filename)
			separator := " "
			if strings.Contains(matches[1], "\t") {
				separator = "\t"
			}
			value := strings.ReplaceAll(matches[1], "  ", separator)
			parts := strings.Split(value, separator)
			parts = helpers.RemoveEmptyStrings(parts)
			parameter := models.Parameter{}
			parameter.Name = parts[0]
			parameter.Type = "header"
			if len(parts) > 2 {
				parameter.ValueType = parts[1]
			}
			if len(parts) > 2 {
				parameter.Required = parts[2] == "true"
			}
			if len(parts) > 3 {
				parameter.Description = helpers.CleanString(parts[3])
			}

			endpoint.Headers = append(endpoint.Headers, parameter)
		}
	}

	return nil
}
