package models

type PDFileCommand struct {
	Command  string `json:"COMMAND,omitempty" yaml:"COMMAND,omitempty"`
	Argument string `json:"ARGUMENT,omitempty" yaml:"ARGUMENT,omitempty"`
}
