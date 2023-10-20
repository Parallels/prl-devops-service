package models

type ApiErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewFromError(err error) ApiErrorResponse {
	return NewFromErrorWithCode(err, 500)
}

func NewFromErrorWithCode(err error, code int) ApiErrorResponse {
	return ApiErrorResponse{
		Message: err.Error(),
		Code:    code,
	}
}
