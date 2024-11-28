package models

import (
	"github.com/Parallels/prl-devops-service/errors"
)

type ApiErrorStack struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path,omitempty"`
	Code        int    `json:"code,omitempty"`
}

type ApiErrorResponse struct {
	Message string          `json:"message"`
	Stack   []ApiErrorStack `json:"stack,omitempty"`
	Code    int             `json:"code,omitempty"`
}

func IsSystemError(err error) bool {
	_, ok := err.(*errors.SystemError)
	if !ok {
		_, ok = err.(errors.SystemError)
		return ok
	}

	return ok
}

func GetSystemErrorCode(err error) int {
	_, ok := err.(*errors.SystemError)
	if !ok {
		_, ok = err.(errors.SystemError)
		if ok {
			sysError := err.(errors.SystemError)
			code := sysError.Code()
			if code == 0 {
				return 400
			}
			return code
		}

		return 400
	}

	code := err.(*errors.SystemError).Code()
	if code == 0 {
		return 400
	}
	return code
}

func NewFromError(err error) ApiErrorResponse {
	if IsSystemError(err) {

		code := GetSystemErrorCode(err)
		result := ApiErrorResponse{
			Message: err.Error(),
			Code:    code,
		}
		return result
	} else {
		return NewFromErrorWithCode(err, 404)
	}
}

func NewFromErrorWithCode(err error, code int) ApiErrorResponse {
	message := ApiErrorResponse{
		Code: code,
	}
	if IsSystemError(err) {
		sysError := extractSystemError(err)
		message.Message = sysError.Message()
		if len(sysError.Stack) > 0 {
			for _, nestedError := range sysError.Stack {
				stack := ApiErrorStack{
					Error: nestedError.Error,
				}
				if nestedError.Path != nil {
					stack.Path = *nestedError.Path
				}
				if nestedError.Code != nil {
					stack.Code = *nestedError.Code
				}
				if nestedError.Description != nil {
					stack.Description = *nestedError.Description
				}

				message.Stack = append(message.Stack, stack)
			}
		}
	} else {
		message.Message = err.Error()
	}

	return message
}

func (r *ApiErrorResponse) ToError() *errors.SystemError {
	err := errors.NewWithCode(r.Message, r.Code)
	if len(r.Stack) > 0 {
		err.Stack = make([]errors.StackItem, 0)
		for _, stack := range r.Stack {
			nestedError := errors.StackItem{
				Error: stack.Error,
			}
			if stack.Path != "" {
				nestedError.Path = &stack.Path
			}
			if stack.Code != 0 {
				nestedError.Code = &stack.Code
			}
			if stack.Description != "" {
				nestedError.Description = &stack.Description
			}
			err.Stack = append(err.Stack, nestedError)
		}
	}

	return err
}

func extractSystemError(err error) *errors.SystemError {
	if IsSystemError(err) {
		sysError, ok := err.(errors.SystemError)
		if !ok {
			sysErrorP, ok := err.(*errors.SystemError)
			if ok {
				sysError = *sysErrorP
			}
		}
		return &sysError
	}

	return nil
}
