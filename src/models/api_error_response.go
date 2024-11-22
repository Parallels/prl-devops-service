package models

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/errors"
)

type ApiErrorStack struct {
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path,omitempty"`
	Code        int    `json:"code"`
}

type ApiErrorResponse struct {
	Message string          `json:"message"`
	Stack   []ApiErrorStack `json:"stack,omitempty"`
	Code    int             `json:"code"`
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
	fmt.Println("[TEST]: %v", err)
	message := ApiErrorResponse{
		Code: code,
	}
	if IsSystemError(err) {
		sysError, ok := err.(errors.SystemError)
		if ok {
			message.Message = sysError.Message()
			if len(sysError.NestedError) > 0 {
				for _, nestedError := range sysError.NestedError {
					stack := ApiErrorStack{
						Message: nestedError.Message,
					}
					if nestedError.Path != "" {
						stack.Path = nestedError.Path
					}
					if nestedError.Code != 0 {
						stack.Code = nestedError.Code
					}
					if nestedError.Description != "" {
						stack.Description = nestedError.Description
					}

					message.Stack = append(message.Stack, stack)
				}
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
		err.NestedError = make([]errors.NestedError, 0)
		for _, stack := range r.Stack {
			nestedError := errors.NestedError{
				Message:     stack.Message,
				Code:        stack.Code,
				Path:        stack.Path,
				Description: stack.Description,
			}
			err.NestedError = append(err.NestedError, nestedError)
		}
	}

	return err
}
