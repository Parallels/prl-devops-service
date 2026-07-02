package models

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/Parallels/prl-devops-service/errors"
)

type ReverseProxyHostHttpRoute struct {
	ID              string                      `json:"id,omitempty" yaml:"id,omitempty"`
	Order           int                         `json:"order,omitempty" yaml:"order,omitempty"`
	Path            string                      `json:"path,omitempty" yaml:"path,omitempty"`
	TargetVmId      string                      `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
	TargetHost      string                      `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetPort      string                      `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	TargetVmDetails *ReverseProxyRouteVmDetails `json:"target_vm_details,omitempty" yaml:"target_vm_details,omitempty"`
	Schema          string                      `json:"schema,omitempty" yaml:"scheme,omitempty"`
	Pattern         string                      `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	RegexpPattern   *regexp.Regexp              `json:"-" yaml:"-"`
	RequestHeaders  map[string]string           `json:"request_headers,omitempty" yaml:"request_headers,omitempty"`
	ResponseHeaders map[string]string           `json:"response_headers,omitempty" yaml:"response_headers,omitempty"`
}

func (r *ReverseProxyHostHttpRoute) Validate(diag *errors.Diagnostics) {
	if r.TargetHost == "" && r.TargetVmId == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing target host or target vm id for TCP route", "")
		return
	}
	if r.TargetPort == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing target port for TCP route", "")
		return
	}
	if r.Path == "" && r.Pattern == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing path or pattern for HTTP route", "")
		return
	}

	if r.Path != "" && r.Pattern != "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "HTTP route cannot have both path and pattern", "")
		return
	}

	return
}

type ReverseProxyHostHttpRouteCreateRequest struct {
	Order           int               `json:"order,omitempty" yaml:"order,omitempty"`
	Path            string            `json:"path,omitempty" yaml:"path,omitempty"`
	TargetVmId      string            `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
	TargetHost      string            `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetPort      string            `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	Schema          string            `json:"schema,omitempty" yaml:"scheme,omitempty"`
	Pattern         string            `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty" yaml:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty" yaml:"response_headers,omitempty"`
}

func (r *ReverseProxyHostHttpRouteCreateRequest) Validate(diag *errors.Diagnostics) {
	if r.TargetPort == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing target port for TCP route", "")
		return
	}

	if r.TargetHost == "" && r.TargetVmId == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing target host or target vm id for TCP route", "")
		return
	}
	if r.Path == "" && r.Pattern == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing path or pattern for HTTP route", "")
		return
	}

	if r.Path != "" && r.Pattern != "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "HTTP route cannot have both path and pattern", "")
		return
	}

	if r.Pattern != "" {
		_, err := regexp.Compile(r.Pattern)
		if err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid pattern for HTTP route", "")
			return
		}
	}

	return
}

func (r *ReverseProxyHostHttpRouteCreateRequest) GetRoute() string {
	if r.Path != "" {
		return r.Path
	}
	if r.Pattern != "" {
		return r.Pattern
	}

	return ""
}

type ReverseProxyHostHttpRouteReorderRequest struct {
	ID    string `json:"id"`
	Order int    `json:"order"`
}

func (r *ReverseProxyHostHttpRouteReorderRequest) Validate(diag *errors.Diagnostics) {
	if r.ID == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing http route id", "")
		return
	}
	if r.Order < 1 {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "invalid order for HTTP route", "")
		return
	}

	return
}
