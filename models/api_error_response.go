package models

type ApiErrorResponse struct {
	Message string `json:"message"`
	Code    int32  `json:"code"`
}

func NewFromError(err error) ApiErrorResponse {
	return NewFromErrorWithCode(err, 500)
}

func NewFromErrorWithCode(err error, code int32) ApiErrorResponse {
	return ApiErrorResponse{
		Message: err.Error(),
		Code:    code,
	}
}
