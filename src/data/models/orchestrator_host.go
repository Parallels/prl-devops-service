package models

import (
	"strings"

	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
)

type OrchestratorHost struct {
	ID                        string                          `json:"id"`
	Host                      string                          `json:"host"`
	Description               string                          `json:"description,omitempty"`
	Tags                      []string                        `json:"tags,omitempty"`
	Port                      string                          `json:"port,omitempty"`
	Schema                    string                          `json:"schema,omitempty"`
	PathPrefix                string                          `json:"path_prefix,omitempty"`
	Authentication            *OrchestratorHostAuthentication `json:"authentication,omitempty"`
	Resources                 *HostResources                  `json:"resources,omitempty"`
	State                     string                          `json:"state,omitempty"`
	LastUnhealthy             string                          `json:"last_seen,omitempty"`
	LastUnhealthyErrorMessage string                          `json:"last_seen_error_message,omitempty"`
	HealthCheck               *models.ApiHealthCheck          `json:"-"`
	VirtualMachines           []VirtualMachine                `json:"virtual_machines,omitempty"`
	CreatedAt                 string                          `json:"created_at,omitempty"`
	UpdatedAt                 string                          `json:"updated_at,omitempty"`
	RequiredClaims            []string                        `json:"required_claims,omitempty"`
	RequiredRoles             []string                        `json:"required_roles,omitempty"`
}

func (o OrchestratorHost) GetHost() string {
	port := ""
	if o.Port != "" {
		port = ":" + o.Port
	}
	schema := o.Schema
	if schema != "" && !strings.HasSuffix(schema, "://") {
		schema = schema + "://"
	}
	url, err := helpers.JoinUrl([]string{schema, o.Host, port, o.PathPrefix})
	if err != nil {
		return ""
	}

	return url.String()
}

func (o *OrchestratorHost) SetUnhealthy(reason string) {
	o.State = "unhealthy"
	o.LastUnhealthy = helpers.GetUtcCurrentDateTime()
	o.LastUnhealthyErrorMessage = reason
}

func (o *OrchestratorHost) SetHealthy() {
	o.State = "healthy"
	o.LastUnhealthy = ""
	o.LastUnhealthyErrorMessage = ""
}

func (o *OrchestratorHost) Diff(source OrchestratorHost) bool {
	if o.Host != source.Host {
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

	return false
}

func (o OrchestratorHost) GetRequiredClaims() []string {
	return o.RequiredClaims
}

func (o OrchestratorHost) GetRequiredRoles() []string {
	return o.RequiredRoles
}
