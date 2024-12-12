package examples_processor

import "github.com/Parallels/prl-devops-service/api_documentation/package/models"

type IExampleProcessor interface {
	Process(endpoint *models.Endpoint) error
}

type ExamplesProcessor struct {
	processors []IExampleProcessor
}

func NewExamplesProcessor() *ExamplesProcessor {
	return &ExamplesProcessor{
		processors: []IExampleProcessor{
			&CurlExampleProcessor{},
			&CSharpExampleProcessor{},
			&GoLangExampleProcessor{},
		},
	}
}

func (p *ExamplesProcessor) AddProcessor(processor IExampleProcessor) {
	p.processors = append(p.processors, processor)
}

func (p *ExamplesProcessor) Process(endpoint *models.Endpoint) error {
	for _, processor := range p.processors {
		if err := processor.Process(endpoint); err != nil {
			return err
		}
	}

	return nil
}
