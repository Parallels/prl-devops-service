package models

import "github.com/Parallels/prl-devops-service/errors"

type ReverseProxyHostTcpRoute struct {
	ID              string                      `json:"id,omitempty" yaml:"id,omitempty"`
	TargetPort      string                      `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	TargetHost      string                      `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetVmId      string                      `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
	TargetVmDetails *ReverseProxyRouteVmDetails `json:"target_vm_details,omitempty" yaml:"target_vm_details,omitempty"`
}

func (r *ReverseProxyHostTcpRoute) Validate() error {
	if r.TargetHost == "" && r.TargetVmId == "" {
		return errors.NewWithCode("missing target host or target vm id for TCP route", 400)
	}
	if r.TargetPort == "" {
		return errors.NewWithCode("missing target port for TCP route", 400)
	}

	return nil
}

type ReverseProxyHostTcpRouteCreateRequest struct {
	TargetPort string `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	TargetHost string `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetVmId string `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
}

func (r *ReverseProxyHostTcpRouteCreateRequest) Validate() error {
	if r.TargetHost == "" && r.TargetVmId == "" {
		return errors.NewWithCode("missing target host or target vm id for TCP route", 400)
	}
	if r.TargetPort == "" {
		return errors.NewWithCode("missing target port for TCP route", 400)
	}

	return nil
}
