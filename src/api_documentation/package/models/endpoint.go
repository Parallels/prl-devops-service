package models

type Endpoint struct {
	ApiPrefix             string      `yaml:"-"`
	HostUrl               string      `yaml:"-"`
	Title                 string      `yaml:"title,omitempty"`
	Description           string      `yaml:"description,omitempty"`
	RequiresAuth          bool        `yaml:"requires_authorization,omitempty"`
	Category              string      `yaml:"category,omitempty"`
	CategoryPath          string      `yaml:"category_path,omitempty"`
	Path                  string      `yaml:"path,omitempty"`
	Method                string      `yaml:"method,omitempty"`
	Headers               []Parameter `yaml:"headers,omitempty"`
	Parameters            []Parameter `yaml:"parameters,omitempty"`
	Content               []string    `yaml:"-"`
	Roles                 []string    `yaml:"default_required_roles,omitempty"`
	Claims                []string    `yaml:"default_required_claims,omitempty"`
	MarkdownContent       string      `yaml:"content_markdown,omitempty"`
	ExampleRequestPayload []string    `yaml:"-"`
	ResponseBlocks        []CodeBlock `yaml:"response_blocks,omitempty"`
	ExamplesBlocks        []CodeBlock `yaml:"example_blocks,omitempty"`
}
