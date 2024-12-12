package examples_processor

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type CSharpExampleProcessor struct{}

func (cep *CSharpExampleProcessor) Process(endpoint *models.Endpoint) error {
	result := models.CodeBlock{
		Title:    "C#",
		Language: "csharp",
	}
	codeBlock := "var client = new HttpClient();\n\n"

	methodCase := cases.Title(language.Und)
	method := methodCase.String(endpoint.Method)
	codeBlock += fmt.Sprintf("var request = new HttpRequestMessage(HttpMethod.%v, \"%v\");\n", method, endpoint.HostUrl+endpoint.ApiPrefix+endpoint.Path)
	if endpoint.RequiresAuth {
		codeBlock += "request.Headers.Add(\"Authorization\", \"••••••\");\n"
	}
	for _, param := range endpoint.Parameters {
		if param.Body != "" {
			codeBlock += "request.Headers.Add(\"Content-Type\", \"application/json\");\n"
			codeBlock += fmt.Sprintf("request.Content = new StringContent(\"%v\");\n", param.Body)
			codeBlock += "request.Content = content;\n"
		}
	}
	codeBlock += "var response = await client.SendAsync(request);\n"
	codeBlock += "response.EnsureSuccessStatusCode();\n"
	codeBlock += "var responseString = await response.Content.ReadAsStringAsync();\n"
	result.CodeBlock = codeBlock

	if endpoint.ExamplesBlocks == nil {
		endpoint.ExamplesBlocks = []models.CodeBlock{}
	}

	endpoint.ExamplesBlocks = append(endpoint.ExamplesBlocks, result)
	return nil
}
