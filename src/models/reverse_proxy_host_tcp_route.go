package models

import (
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/errors"
)

type ReverseProxyHostTcpRoute struct {
	ID              string                      `json:"id,omitempty" yaml:"id,omitempty"`
	TargetPort      string                      `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	TargetHost      string                      `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetVmId      string                      `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
	TargetVmDetails *ReverseProxyRouteVmDetails `json:"target_vm_details,omitempty" yaml:"target_vm_details,omitempty"`
}

func (r *ReverseProxyHostTcpRoute) Validate(diag *errors.Diagnostics) {
	if r.TargetHost == "" && r.TargetVmId == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing target host or target vm id for TCP route", "")
		return
	}
	if r.TargetPort == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing target port for TCP route", "")
		return
	}
}

type ReverseProxyHostTcpRouteCreateRequest struct {
	TargetPort string `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	TargetHost string `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetVmId string `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
}

func (r *ReverseProxyHostTcpRouteCreateRequest) Validate(diag *errors.Diagnostics) {
	if r.TargetHost == "" && r.TargetVmId == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing target host or target vm id for TCP route", "")
		return
	}
	if r.TargetPort == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing target port for TCP route", "")
		return
	}

	return
}
