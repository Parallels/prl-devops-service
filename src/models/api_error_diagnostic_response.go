package models

import (
	"github.com/Parallels/prl-devops-service/errors"
)

type ApiErrorDiagnosticsResponse struct {
	Message     string              `json:"message"`
	Diagnostics *errors.Diagnostics `json:"diagnostics"`
	Code        int                 `json:"code,omitempty"`
}

func NewDiagnosticsWithCode(diag *errors.Diagnostics, code int) ApiErrorDiagnosticsResponse {

	apiRspDiag := errors.NewDiagnostics("api error response")
	if diag != nil {
		apiRspDiag.Append(diag)
	}
	apiRspDiag.Complete()
	message := ApiErrorDiagnosticsResponse{
		Code:        code,
		Diagnostics: apiRspDiag,
		Message:     apiRspDiag.GetSummary(),
	}
	return message
}
