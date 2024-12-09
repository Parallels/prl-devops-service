package examples_processor

import (
	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

type GoLangExampleProcessor struct{}

func (cep *GoLangExampleProcessor) Process(endpoint *models.Endpoint) error {
	result := models.CodeBlock{
		Title:    "Go",
		Language: "go",
	}
	codeBlock := "package main\n\n"
	codeBlock += "import (\n"
	codeBlock += "  \"fmt\"\n"
	codeBlock += "  \"net/http\"\n"
	codeBlock += "  \"strings\"\n"
	codeBlock += "  \"io\"\n"
	codeBlock += ")\n\n"

	codeBlock += "func main() {\n"
	codeBlock += "  url := \"" + endpoint.HostUrl + endpoint.ApiPrefix + endpoint.Path + "\"\n"
	codeBlock += "  method := \"" + endpoint.Method + "\"\n"
	for _, param := range endpoint.Parameters {
		if param.Body != "" {
			codeBlock += "  payload := strings.NewReader(`" + param.Body + "`)\n"
		}
	}

	codeBlock += "  client := &http.Client{}\n"
	codeBlock += "  req, err := http.NewRequest(method, url, payload)\n"
	codeBlock += "  if err != nil {\n"
	codeBlock += "    fmt.Println(err)\n"
	codeBlock += "    return\n"
	codeBlock += "  }\n"
	codeBlock += "  req.Header.Add(\"Content-Type\", \"application/json\")\n"
	if endpoint.RequiresAuth {
		codeBlock += "\n  req.Header.Add(\"Authorization\", \"••••••\")\n"
	}
	codeBlock += "  res, err := client.Do(req)\n"
	codeBlock += "  if err != nil {\n"
	codeBlock += "    fmt.Println(err)\n"
	codeBlock += "    return\n"
	codeBlock += "  }\n"
	codeBlock += "  defer res.Body.Close()\n"
	codeBlock += "  body, err := io.ReadAll(res.Body)\n"
	codeBlock += "  if err != nil {\n"
	codeBlock += "    fmt.Println(err)\n"
	codeBlock += "    return\n"
	codeBlock += "  }\n"
	codeBlock += "  fmt.Println(string(body))\n"
	codeBlock += "}\n"

	result.CodeBlock = codeBlock

	if endpoint.ExamplesBlocks == nil {
		endpoint.ExamplesBlocks = []models.CodeBlock{}
	}

	endpoint.ExamplesBlocks = append(endpoint.ExamplesBlocks, result)
	return nil
}
