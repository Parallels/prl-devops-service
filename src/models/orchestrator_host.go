package models

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Parallels/prl-devops-service/errors"
)

type HostResourceItem struct {
	PhysicalCpuCount int64   `json:"physical_cpu_count,omitempty"`
	LogicalCpuCount  int64   `json:"logical_cpu_count"`
	MemorySize       float64 `json:"memory_size,omitempty"`
	DiskSize         float64 `json:"disk_size,omitempty"`
	FreeDiskSize     float64 `json:"free_disk_size,omitempty"`
}

type OrchestratorAuthentication struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	ApiKey   string `json:"api_key,omitempty"`
}

type OrchestratorHostRequest struct {
	HostUrl        *url.URL                    `json:"-"`
	Host           string                      `json:"host"`
	HostName       string                      `json:"-"`
	Port           string                      `json:"port"`
	Schema         string                      `json:"schema"`
	Prefix         string                      `json:"prefix"`
	Description    string                      `json:"description,omitempty"`
	Tags           []string                    `json:"tags,omitempty"`
	Authentication *OrchestratorAuthentication `json:"authentication,omitempty"`
	RequiredClaims []string                    `json:"required_claims,omitempty"`
	RequiredRoles  []string                    `json:"required_roles,omitempty"`
}

func (o *OrchestratorHostRequest) Validate() error {
	if o.Host == "" {
		return errors.NewWithCode("Host cannot be empty", 400)
	}
	if !strings.Contains(o.Host, "http://") && !strings.Contains(o.Host, "https://") {
		o.Host = "http://" + o.Host
	}
	hostUrl, err := url.Parse(o.Host)
	if err != nil {
		return errors.NewWithCode("Invalid host", 400)
	}
	o.HostUrl = hostUrl
	o.HostName = hostUrl.Hostname()
	o.Port = hostUrl.Port()
	o.Schema = hostUrl.Scheme

	if o.HostUrl.Path == "" {
		o.Prefix = "/api"
	} else {
		o.Prefix = o.HostUrl.Path
	}

	o.Host = fmt.Sprintf("%s://%s:%s", o.Schema, o.HostName, o.Port)
	if o.Prefix != "" {
		o.Host = fmt.Sprintf("%s%s", o.Host, o.Prefix)
	}

	if o.Authentication.Username == "" && o.Authentication.Password == "" && o.Authentication.ApiKey == "" {
		return errors.NewWithCode("Authentication cannot be empty", 400)
	}
	if o.Authentication.Username != "" && o.Authentication.Password == "" {
		return errors.NewWithCode("Password cannot be empty", 400)
	}

	if o.Tags == nil {
		o.Tags = []string{}
	}
	if o.RequiredClaims == nil {
		o.RequiredClaims = []string{}
	}
	if o.RequiredRoles == nil {
		o.RequiredRoles = []string{}
	}
	return nil
}

type OrchestratorHostResponse struct {
	ID             string           `json:"id"`
	Enabled        bool             `json:"enabled"`
	Host           string           `json:"host"`
	Architecture   string           `json:"architecture"`
	CpuModel       string           `json:"cpu_model"`
	Description    string           `json:"description,omitempty"`
	Tags           []string         `json:"tags,omitempty"`
	State          string           `json:"state,omitempty"`
	Resources      HostResourceItem `json:"resources,omitempty"`
	RequiredClaims []string         `json:"required_claims,omitempty"`
	RequiredRoles  []string         `json:"required_roles,omitempty"`
}

type OrchestratorHostUpdateRequest struct {
	HostUrl        *url.URL                    `json:"-"`
	Host           string                      `json:"host"`
	HostName       string                      `json:"-"`
	Port           string                      `json:"port"`
	Schema         string                      `json:"schema"`
	Prefix         string                      `json:"prefix"`
	Description    string                      `json:"description,omitempty"`
	Authentication *OrchestratorAuthentication `json:"authentication,omitempty"`
}

func (o *OrchestratorHostUpdateRequest) Validate() error {
	if o.Host != "" {
		if !strings.Contains(o.Host, "http://") && !strings.Contains(o.Host, "https://") {
			o.Host = "http://" + o.Host
		}
		hostUrl, err := url.Parse(o.Host)
		if err != nil {
			return errors.NewWithCode("Invalid host", 400)
		}
		o.HostUrl = hostUrl
		o.HostName = hostUrl.Hostname()
		o.Port = hostUrl.Port()
		o.Schema = hostUrl.Scheme

		if o.HostUrl.Path == "" {
			o.Prefix = "/api"
		} else {
			o.Prefix = o.HostUrl.Path
		}

		o.Host = fmt.Sprintf("%s://%s:%s", o.Schema, o.HostName, o.Port)
		if o.Prefix != "" {
			o.Host = fmt.Sprintf("%s%s", o.Host, o.Prefix)
		}
	}

	if o.Authentication != nil && o.Authentication.Username != "" && o.Authentication.Password == "" {
		return errors.NewWithCode("Authentication password cannot be empty", 400)
	}

	return nil
}
