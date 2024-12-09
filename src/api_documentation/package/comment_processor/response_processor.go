package comment_processor

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/api_documentation/package/cache"
	"github.com/Parallels/prl-devops-service/api_documentation/package/helpers"
	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

var (
	successRegex = regexp.MustCompile(`^//\s*@Success\s+(.*)$`)
	failureRegex = regexp.MustCompile(`^//\s*@Failure\s+(.*)$`)
)

type ResponseCommentProcessor struct{}

func (p *ResponseCommentProcessor) ProcessComment(endpoint *models.Endpoint, filename, comment string) error {
	var matches []string
	if successMatches := successRegex.FindStringSubmatch(comment); successMatches != nil {
		matches = successMatches
	}
	if failureMatches := failureRegex.FindStringSubmatch(comment); failureMatches != nil {
		matches = failureMatches
	}
	if matches == nil || len(matches) < 2 {
		return nil
	}

	log.Printf("Response comment %v found in %s, processing", matches[1], filename)
	separator := " "
	if strings.Contains(matches[1], "\t") {
		separator = "\t"
	}

	value := strings.ReplaceAll(matches[1], "  ", separator)
	parts := strings.Split(value, separator)
	parts = helpers.RemoveEmptyStrings(parts)
	if len(parts) > 1 && parts[1] == "{object}" {
		modelName := parts[2]
		if modelName != "" {
			httpCode, err := strconv.Atoi(parts[0])
			if err != nil {
				httpCode = 200
			}

			codeBlock := models.CodeBlock{
				Title:           modelName,
				Code:            parts[0],
				CodeDescription: http.StatusText(httpCode),
				Language:        "json",
			}

			modelNameParts := strings.Split(modelName, ".")
			cachedModel := cache.Get(modelNameParts[len(modelNameParts)-1])
			if cachedModel != cache.CacheObjectNotFound {
				codeBlock.CodeBlock = cachedModel
			} else {
				jsonOutput := helpers.ConvertModelToJson(filename, modelName)
				codeBlock.CodeBlock = jsonOutput
				cache.Push(modelNameParts[len(modelNameParts)-1], jsonOutput)
			}
			endpoint.ResponseBlocks = append(endpoint.ResponseBlocks, codeBlock)
		}
	}

	return nil
}
