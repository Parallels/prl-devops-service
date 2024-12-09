package comment_processor

import (
	"log"
	"regexp"
	"strings"

	"github.com/Parallels/prl-devops-service/api_documentation/package/cache"
	"github.com/Parallels/prl-devops-service/api_documentation/package/helpers"
	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var paramRegex = regexp.MustCompile(`^//\s*@Param\s+(.*)$`)

type ParamCommentProcessor struct{}

func (p *ParamCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename, comment string) error {
	if matches := paramRegex.FindStringSubmatch(comment); matches != nil {
		if len(matches) > 1 {
			log.Printf("Param comment %v found in %s, processing", matches[1], filename)
			separator := " "
			if strings.Contains(matches[1], "\t") {
				separator = "\t"
			}

			value := strings.ReplaceAll(matches[1], "  ", separator)
			parts := strings.Split(value, separator)
			parts = helpers.RemoveEmptyStrings(parts)
			parameter := models.Parameter{}
			parameter.Name = parts[0]
			parameter.Type = "path"
			if len(parts) > 1 {
				parameter.Type = parts[1]
			}
			switch parameter.Type {
			case "path":
				processPathParameter(parts, &parameter)
			case "body":
				processBodyParameter(parts, filename, &parameter)
			}

			endpoint.Parameters = append(endpoint.Parameters, parameter)
		}
	}

	return nil
}

func processPathParameter(parts []string, parameter *models.Parameter) {
	if len(parts) > 2 {
		parameter.ValueType = parts[2]
	}
	if len(parts) > 3 {
		parameter.Required = parts[3] == "true"
	}
	if len(parts) > 4 {
		parameter.Description = strings.ReplaceAll(parts[4], "\"", "")
	}
}

func processBodyParameter(parts []string, filename string, parameter *models.Parameter) {
	parameter.ValueType = "object"
	if len(parts) > 2 {
		modelName := parts[2]
		if modelName != "" {
			modelNameParts := strings.Split(modelName, ".")
			cachedModel := cache.Get(modelNameParts[len(modelNameParts)-1])
			if cachedModel != cache.CacheObjectNotFound {
				parameter.Body = cachedModel
			} else {
				jsonOutput := helpers.ConvertModelToJson(filename, modelName)
				parameter.Body = jsonOutput
				cache.Push(modelNameParts[len(modelNameParts)-1], jsonOutput)
			}
		}
	}
	if len(parts) > 4 {
		parameter.Description = strings.ReplaceAll(parts[4], "\"", "")
	}
}
