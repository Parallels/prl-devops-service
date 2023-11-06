package models

import (
	"bytes"
	"encoding/json"
	"fmt"

	log "github.com/cjlapao/common-go-logger"
)

// OAuthErrorType Enum
type OAuthErrorType int64

const (
	OAuthInvalidRequestError OAuthErrorType = iota
	OAuthInvalidClientError
	OAuthInvalidGrant
	OAuthInvalidScope
	OAuthUserExists
	OAuthUnauthorizedClient
	OAuthUnsupportedGrantType
	OAuthPasswordMismatch
	OAuthPasswordValidation
	OAuthUserValidation
	OAuthEmailNotVerified
	OAuthUserBlocked
	UnknownError
)

func (oAuthErrorType OAuthErrorType) String() string {
	return toOAuthErrorTypeString[oAuthErrorType]
}

func (oAuthErrorType OAuthErrorType) FromString(keyType string) OAuthErrorType {
	return toOAuthErrorTypeID[keyType]
}

var toOAuthErrorTypeString = map[OAuthErrorType]string{
	OAuthInvalidRequestError:  "invalid_request",
	OAuthInvalidClientError:   "invalid_client",
	OAuthInvalidGrant:         "invalid_grant",
	OAuthInvalidScope:         "invalid_scope",
	OAuthUnauthorizedClient:   "unauthorized_client",
	OAuthUnsupportedGrantType: "unsupported_grant_type",
	OAuthPasswordMismatch:     "password_mismatch",
	OAuthPasswordValidation:   "password_validation",
	OAuthUserValidation:       "user_validation",
	OAuthUserExists:           "user_exists",
	OAuthEmailNotVerified:     "email_not_verified",
	OAuthUserBlocked:          "user_blocked",
	UnknownError:              "unknown_error",
}

var toOAuthErrorTypeID = map[string]OAuthErrorType{
	"invalid_request":        OAuthInvalidRequestError,
	"invalid_client":         OAuthInvalidClientError,
	"invalid_grant":          OAuthInvalidGrant,
	"invalid_scope":          OAuthInvalidScope,
	"unauthorized_client":    OAuthUnauthorizedClient,
	"unsupported_grant_type": OAuthUnsupportedGrantType,
	"password_mismatch":      OAuthPasswordMismatch,
	"password_validation":    OAuthPasswordValidation,
	"user_validation":        OAuthUserValidation,
	"user_exists":            OAuthUserExists,
	"email_not_verified":     OAuthEmailNotVerified,
	"user_blocked":           OAuthUserBlocked,
	"unknown_error":          UnknownError,
}

func (oAuthErrorType OAuthErrorType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toOAuthErrorTypeString[oAuthErrorType])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (oAuthErrorType *OAuthErrorType) UnmarshalJSON(b []byte) error {
	var key string
	err := json.Unmarshal(b, &key)
	if err != nil {
		return err
	}

	*oAuthErrorType = toOAuthErrorTypeID[key]
	return nil
}

// OAuthErrorResponse entity
type OAuthErrorResponse struct {
	Error            OAuthErrorType `json:"error"`
	ErrorDescription string         `json:"error_description,omitempty"`
	ErrorUri         string         `json:"error_uri,omitempty"`
}

func NewOAuthErrorResponse(err OAuthErrorType, description string) OAuthErrorResponse {
	errorResponse := OAuthErrorResponse{
		Error:            err,
		ErrorDescription: description,
	}
	return errorResponse

}

func (err OAuthErrorResponse) String() string {
	return fmt.Sprintf("An error occurred, %v: %v", err.Error.String(), err.ErrorDescription)
}

func (err OAuthErrorResponse) Log(extraLogs ...string) {
	logger := log.Get()
	if len(extraLogs) == 0 {
		logger.Error(err.String())
	} else {
		logger.Error("%v %v", err.String(), extraLogs[0])
	}
}
