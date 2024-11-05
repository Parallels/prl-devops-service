package models

import (
	"regexp"

	"github.com/Parallels/prl-devops-service/errors"
)

type ReverseProxyHostHttpRoute struct {
	ID              string            `json:"id,omitempty" yaml:"id,omitempty"`
	Path            string            `json:"path,omitempty" yaml:"path,omitempty"`
	TargetVmId      string            `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
	TargetHost      string            `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetPort      string            `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	Schema          string            `json:"schema,omitempty" yaml:"scheme,omitempty"`
	Pattern         string            `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	RegexpPattern   *regexp.Regexp    `json:"-" yaml:"-"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty" yaml:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty" yaml:"response_headers,omitempty"`
}

func (r *ReverseProxyHostHttpRoute) Validate() error {
	if r.TargetHost == "" && r.TargetVmId == "" {
		return errors.NewWithCode("missing target host or target vm id for TCP route", 400)
	}
	if r.TargetPort == "" {
		return errors.NewWithCode("missing target port for TCP route", 400)
	}
	if r.Path == "" && r.Pattern == "" {
		return errors.NewWithCode("missing path or pattern for HTTP route", 400)
	}

	if r.Path != "" && r.Pattern != "" {
		return errors.NewWithCode("HTTP route cannot have both path and pattern", 400)
	}

	return nil
}

type ReverseProxyHostHttpRouteCreateRequest struct {
	Path            string            `json:"path,omitempty" yaml:"path,omitempty"`
	TargetVmId      string            `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
	TargetHost      string            `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetPort      string            `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	Schema          string            `json:"schema,omitempty" yaml:"scheme,omitempty"`
	Pattern         string            `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty" yaml:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty" yaml:"response_headers,omitempty"`
}

func (r *ReverseProxyHostHttpRouteCreateRequest) Validate() error {
	if r.TargetPort == "" {
		return errors.NewWithCode("missing target port for TCP route", 400)
	}

	if r.TargetHost == "" && r.TargetVmId == "" {
		return errors.NewWithCode("missing target host or target vm id for TCP route", 400)
	}
	if r.Path == "" && r.Pattern == "" {
		return errors.NewWithCode("missing path or pattern for HTTP route", 400)
	}

	if r.Path != "" && r.Pattern != "" {
		return errors.NewWithCode("HTTP route cannot have both path and pattern", 400)
	}

	if r.Pattern != "" {
		_, err := regexp.Compile(r.Pattern)
		if err != nil {
			return errors.NewWithCode("invalid pattern for HTTP route", 400)
		}
	}

	return nil
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
