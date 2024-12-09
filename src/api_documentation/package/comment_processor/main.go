package comment_processor

import (
	"github.com/Parallels/prl-devops-service/api_documentation/package/models"
)

type ICommentProcessor interface {
	ProcessComment(endpoint *models.Endpoint, filename string, comment string) error
}

type CommentProcessor struct {
	processors []ICommentProcessor
}

func NewCommentProcessor() *CommentProcessor {
	return &CommentProcessor{
		processors: []ICommentProcessor{
			&ClaimsCommentProcessor{},
			&ContentCommentProcessor{},
			&DescriptionCommentProcessor{},
			&ExamplesCommentProcessor{},
			&HeaderCommentProcessor{},
			&ParamCommentProcessor{},
			&ResponseCommentProcessor{},
			&RolesCommentProcessor{},
			&RouterCommentProcessor{},
			&SummaryCommentProcessor{},
			&TagsCommentProcessor{},
		},
	}
}

func (p *CommentProcessor) ProcessComment(endpoint *models.Endpoint, filename string, comment string) error {
	for _, processor := range p.processors {
		err := processor.ProcessComment(endpoint, filename, comment)
		if err != nil {
			return err
		}
	}

	return nil
}
