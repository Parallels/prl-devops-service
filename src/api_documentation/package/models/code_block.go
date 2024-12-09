package models

type CodeBlock struct {
	Type            string `yaml:"type,omitempty"`
	CodeBlock       string `yaml:"code_block,omitempty"`
	Code            string `yaml:"code,omitempty"`
	CodeDescription string `yaml:"code_description,omitempty"`
	Title           string `yaml:"title,omitempty"`
	Language        string `yaml:"language,omitempty"`
}
