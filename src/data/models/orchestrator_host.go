package models

import (
	"net/url"
	"strings"

	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

type OrchestratorHost struct {
	ID                        string                          `json:"id"`
	Enabled                   bool                            `json:"enabled"`
	Host                      string                          `json:"host"`
	Architecture              string                          `json:"architecture"`
	CpuModel                  string                          `json:"cpu_model,omitempty"`
	OsVersion                 string                          `json:"os_version,omitempty"`
	OsName                    string                          `json:"os_name,omitempty"`
	ExternalIpAddress         string                          `json:"external_ip_address,omitempty"`
	DevOpsVersion             string                          `json:"devops_version,omitempty"`
	Description               string                          `json:"description,omitempty"`
	Tags                      []string                        `json:"tags,omitempty"`
	Port                      string                          `json:"port,omitempty"`
	Schema                    string                          `json:"schema,omitempty"`
	PathPrefix                string                          `json:"path_prefix,omitempty"`
	ParallelsDesktopVersion   string                          `json:"parallels_desktop_version,omitempty"`
	ParallelsDesktopLicensed  bool                            `json:"parallels_desktop_licensed,omitempty"`
	Authentication            *OrchestratorHostAuthentication `json:"authentication,omitempty"`
	Resources                 *HostResources                  `json:"resources,omitempty"`
	State                     string                          `json:"state,omitempty"`
	LastUnhealthy             string                          `json:"last_seen,omitempty"`
	LastUnhealthyErrorMessage string                          `json:"last_seen_error_message,omitempty"`
	HealthCheck               *models.ApiHealthCheck          `json:"-"`
	VirtualMachines           []VirtualMachine                `json:"virtual_machines,omitempty"`
	IsReverseProxyEnabled     bool                            `json:"is_reverse_proxy_enabled,omitempty"`
	ReverseProxy              *ReverseProxy                   `json:"reverse_proxy,omitempty"`
	ReverseProxyHosts         []*ReverseProxyHost             `json:"reverse_proxy_hosts,omitempty"`
	CreatedAt                 string                          `json:"created_at,omitempty"`
	UpdatedAt                 string                          `json:"updated_at,omitempty"`
	RequiredClaims            []string                        `json:"required_claims,omitempty"`
	RequiredRoles             []string                        `json:"required_roles,omitempty"`
}

func (o OrchestratorHost) GetHost() string {
	if strings.HasPrefix(o.Host, "http") || strings.HasPrefix(o.Host, "https") {
		hostUrl, err := url.Parse(o.Host)
		if err == nil {
			return o.Host
		}

		o.Host = hostUrl.Hostname()
		o.Port = hostUrl.Port()
		o.Schema = hostUrl.Scheme
	}

	port := ""
	if o.Port != "" {
		port = ":" + o.Port
	}
	schema := o.Schema
	if schema != "" && !strings.HasSuffix(schema, "://") {
		schema += "://"
	}
	url, err := helpers.JoinUrl([]string{schema, o.Host, port, o.PathPrefix})
	if err != nil {
		return ""
	}

	return url.String()
}

func (o *OrchestratorHost) SetUnhealthy(reason string) {
	if o.State == "unhealthy" {
		return
	}

	o.State = "unhealthy"
	o.LastUnhealthy = helpers.GetUtcCurrentDateTime()
	o.LastUnhealthyErrorMessage = reason
}

func (o *OrchestratorHost) SetHealthy() {
	if o.State == "healthy" {
		return
	}

	o.State = "healthy"
	o.LastUnhealthy = ""
	o.LastUnhealthyErrorMessage = ""
}

func (o *OrchestratorHost) Diff(source OrchestratorHost) bool {
	if o.Host != source.Host {
		return true
	}
	if o.Enabled != source.Enabled {
		return true
	}
	if o.OsVersion != source.OsVersion {
		return true
	}
	if o.OsName != source.OsName {
		return true
	}
	if o.DevOpsVersion != source.DevOpsVersion {
		return true
	}
	if o.ExternalIpAddress != source.ExternalIpAddress {
		return true
	}
	if o.Architecture != source.Architecture {
		return true
	}
	if o.CpuModel != source.CpuModel {
		return true
	}
	if o.Description != source.Description {
		return true
	}
	if o.Tags != nil && source.Tags != nil {
		if len(o.Tags) != len(source.Tags) {
			return true
		}
		for i, tag := range o.Tags {
			if tag != source.Tags[i] {
				return true
			}
		}
	}
	if o.Authentication != nil && source.Authentication != nil {
		if o.Authentication.Diff(*source.Authentication) {
			return true
		}
	}
	if o.RequiredClaims != nil && source.RequiredClaims != nil {
		if len(o.RequiredClaims) != len(source.RequiredClaims) {
			return true
		}
		for i, claim := range o.RequiredClaims {
			if claim != source.RequiredClaims[i] {
				return true
			}
		}
	}
	if o.RequiredRoles != nil && source.RequiredRoles != nil {
		if len(o.RequiredRoles) != len(source.RequiredRoles) {
			return true
		}
		for i, role := range o.RequiredRoles {
			if role != source.RequiredRoles[i] {
				return true
			}
		}
	}

	if o.Resources != nil && source.Resources != nil {
		if o.Resources.Diff(*source.Resources) {
			return true
		}
	}

	if o.Authentication != nil && source.Authentication != nil {
		if o.Authentication.Diff(*source.Authentication) {
			return true
		}
	}

	if o.Port != source.Port {
		return true
	}

	if o.Schema != source.Schema {
		return true
	}

	if o.PathPrefix != source.PathPrefix {
		return true
	}

	if o.State != source.State {
		return true
	}

	if o.LastUnhealthy != source.LastUnhealthy {
		return true
	}

	if o.LastUnhealthyErrorMessage != source.LastUnhealthyErrorMessage {
		return true
	}

	// check if we have the same number of VMs
	if len(o.VirtualMachines) != len(source.VirtualMachines) {
		return true
	}

	for _, vm := range o.VirtualMachines {
		for _, vm2 := range source.VirtualMachines {
			if vm.ID == vm2.ID {
				if vm.Diff(vm2) {
					return true
				}
			}
		}
	}

	if o.ReverseProxy != nil && source.ReverseProxy != nil {
		if o.ReverseProxy.Diff(*source.ReverseProxy) {
			return true
		}
	}

	if len(o.ReverseProxyHosts) != len(source.ReverseProxyHosts) {
		return true
	}

	for i, rpHost := range o.ReverseProxyHosts {
		if rpHost.Diff(*source.ReverseProxyHosts[i]) {
			return true
		}
	}

	return false
}

func (o OrchestratorHost) GetRequiredClaims() []string {
	return o.RequiredClaims
}

func (o OrchestratorHost) GetRequiredRoles() []string {
	return o.RequiredRoles
}
