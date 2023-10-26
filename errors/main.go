package errors

import (
	"Parallels/pd-api-service/common"
	"fmt"
)

type SystemError struct {
	message     string
	code        int
	description string
}

func (e *SystemError) Error() string {
	msg := ""
	if e.code != 0 {
		msg = fmt.Sprintf("error code: %v, %v", e.code, e.message)
	} else {
		msg = fmt.Sprintf("error: %v", e.message)
	}
	if e.description != "" {
		msg = fmt.Sprintf("%v, description: %v", msg, e.description)
	}

	return msg
}

func New(message string) *SystemError {
	err := &SystemError{
		message: message,
	}

	common.Logger.Error(err.Error())
	return err
}

func Newf(format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
	}

	common.Logger.Error(err.Error())
	return err
}

func NewFromError(err error) *SystemError {
	return New(err.Error())
}

func NewFromErrorf(err error, format string, a ...interface{}) *SystemError {
	return Newf("%v: %v", fmt.Sprintf(format, a...), err.Error())
}

func NewFromErrorWithCode(err error, code int) *SystemError {
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
	common.Logger.Error(err.Error())
	return err
}

func NewWithCodef(code int, format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
		code:    code,
	}
	common.Logger.Error(err.Error())
	return err
}

func NewWithCodeAndDescription(message string, code int, description string) *SystemError {
	err := &SystemError{
		message:     message,
		code:        code,
		description: description,
	}
	common.Logger.Error(err.Error())
	return err
}

func NewWithCodeAndDescriptionf(code int, format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
		code:    code,
	}
	common.Logger.Error(err.Error())
	return err
}

func NewWithDescription(message string, description string) *SystemError {
	err := &SystemError{
		message:     message,
		description: description,
	}
	common.Logger.Error(err.Error())
	return err
}

func NewWithDescriptionf(format string, a ...interface{}) *SystemError {
	err := &SystemError{
		message: fmt.Sprintf(format, a...),
	}
	common.Logger.Error(err.Error())
	return err
}
