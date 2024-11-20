package errors

import (
	"fmt"
)

type SystemError struct {
	message     string
	NestedError []NestedError
	Path        string
	code        int
	description string
}

func (e SystemError) Error() string {
	msg := ""
	if e.code != 0 {
		msg = fmt.Sprintf("error %v: %v", e.code, e.message)
	} else {
		msg = fmt.Sprintf("error: %v", e.message)
	}
	if e.description != "" {
		msg = fmt.Sprintf("%v, description: %v", msg, e.description)
	}
	if e.Path != "" {
		msg = fmt.Sprintf("%v, path: %v", msg, e.Path)
	}
	if len(e.NestedError) > 0 {
		msg = fmt.Sprintf("%v, nested errors: %v", msg, e.NestedError)
	}

	return msg
}

func (e SystemError) Message() string {
	return e.message
}

func (e SystemError) Description() string {
	return e.description
}

func (e SystemError) Code() int {
	return e.code
}

func New(message string) *SystemError {
	err := &SystemError{
		message: message,
	}

	return err
}

func Newf(format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
	}

	return err
}

func NewFromError(err error) *SystemError {
	if err == nil {
		return nil
	}

	return New(err.Error())
}

func NewFromErrorf(err error, format string, a ...interface{}) *SystemError {
	return Newf("%v: %v", fmt.Sprintf(format, a...), err.Error())
}

func NewFromErrorWithCode(err error, code int) *SystemError {
	if err == nil {
		return nil
	}

	return NewWithCode(err.Error(), code)
}

func NewFromErrorWithCodef(err error, code int, format string, a ...interface{}) *SystemError {
	return NewWithCodef(code, "%v: %v", fmt.Sprintf(format, a...), err.Error())
}

func NewWithCode(message string, code int) *SystemError {
	err := &SystemError{
		message: message,
		code:    code,
	}

	return err
}

func NewWithCodef(code int, format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
		code:    code,
	}

	return err
}

func NewWithCodeAndDescription(message string, code int, description string) *SystemError {
	err := &SystemError{
		message:     message,
		code:        code,
		description: description,
	}

	return err
}

func NewWithCodeAndDescriptionf(code int, format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
		code:    code,
	}

	return err
}

func NewWithDescription(message string, description string) *SystemError {
	err := &SystemError{
		message:     message,
		description: description,
	}

	return err
}

func NewWithDescriptionf(format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
	}

	return err
}

func NewWithCodeAndNestedErrorf(sysError SystemError, code int, format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
		code:    code,
	}
	err.NestedError = make([]NestedError, 0)
	nestedError := NestedError{
		Message:     sysError.Message(),
		Code:        sysError.Code(),
		Path:        sysError.Path,
		Description: sysError.Description(),
	}

	err.NestedError = append(err.NestedError, nestedError)
	if len(sysError.NestedError) > 0 {
		for _, nestedError := range sysError.NestedError {
			nestedError := NestedError{
				Message:     nestedError.Message,
				Code:        nestedError.Code,
				Path:        nestedError.Path,
				Description: nestedError.Description,
			}
			err.NestedError = append(err.NestedError, nestedError)
		}
	}

	return err
}
