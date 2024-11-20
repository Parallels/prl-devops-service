package models

import "github.com/Parallels/prl-devops-service/errors"

type ApiNestedError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type ApiErrorResponse struct {
	Message     string           `json:"message"`
	NestedError []ApiNestedError `json:"nested_error,omitempty"`
	Code        int              `json:"code"`
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
			code := err.(errors.SystemError).Code()
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
	return ApiErrorResponse{
		Message: err.Error(),
		Code:    code,
	}
}
