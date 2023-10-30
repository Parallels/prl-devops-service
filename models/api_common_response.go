package models

type ApiCommonResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code,omitempty"`
	Operation string      `json:"operation,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
}
