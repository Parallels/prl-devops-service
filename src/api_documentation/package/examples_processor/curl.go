package examples_processor

import (
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

type CurlExampleProcessor struct{}

func (cep *CurlExampleProcessor) Process(endpoint *models.Endpoint) error {
	result := models.CodeBlock{
		Title:    "cURL",
		Language: "powershell",
	}
	codeBlock := fmt.Sprintf("curl --location '%v' \n", endpoint.HostUrl+endpoint.ApiPrefix+endpoint.Path)
	if endpoint.RequiresAuth {
		codeBlock += "--header 'Authorization ••••••'\n"
	}
	for _, param := range endpoint.Parameters {
		if param.Body != "" {
			codeBlock += "--header 'Content-Type: application/json' \n"
			if len(endpoint.ExampleRequestPayload) > 0 {
				codeBlock += fmt.Sprintf("--data '%v'\n", strings.Join(endpoint.ExampleRequestPayload, "  "))
			} else {
				codeBlock += fmt.Sprintf("--data '%v'\n", param.Body)
			}
		}
	}
	result.CodeBlock = codeBlock

	if endpoint.ExamplesBlocks == nil {
		endpoint.ExamplesBlocks = []models.CodeBlock{}
	}

	endpoint.ExamplesBlocks = append(endpoint.ExamplesBlocks, result)
	return nil
}
