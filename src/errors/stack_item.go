package errors

type StackItem struct {
	Error       string
	Path        *string
	Code        *int
	Description *string
}

func NewStackItem(err SystemError) StackItem {
	stackItem := StackItem{
		Error: err.Message(),
	}
	if err.Code() != 0 {
		stackItem.Code = &err.code
	}
	if err.Description() != "" {
		stackItem.Description = &err.description
	}
	if err.Path != "" {
		stackItem.Path = &err.Path
	}

	return stackItem
}
