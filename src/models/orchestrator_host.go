package models

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Parallels/prl-devops-service/errors"
)

type HostResourceItem struct {
	TotalAppleVms    int64   `json:"total_apple_vms,omitempty"`
	PhysicalCpuCount int64   `json:"physical_cpu_count,omitempty"`
	LogicalCpuCount  int64   `json:"logical_cpu_count"`
	MemorySize       float64 `json:"memory_size"`
	DiskSize         float64 `json:"disk_size,omitempty"`
	FreeDiskSize     float64 `json:"free_disk_size,omitempty"`
}

type HostResources struct {
	TotalAppleVms  int64            `json:"total_apple_vms,omitempty"`
	SystemReserved HostResourceItem `json:"system_reserved,omitempty"`
	Total          HostResourceItem `json:"total,omitempty"`
	TotalAvailable HostResourceItem `json:"total_available,omitempty"`
	TotalInUse     HostResourceItem `json:"total_in_use,omitempty"`
	TotalReserved  HostResourceItem `json:"total_reserved,omitempty"`
}

type HostReverseProxy struct {
	Enabled bool               `json:"enabled,omitempty"`
	Host    string             `json:"host,omitempty"`
	Port    string             `json:"port,omitempty"`
	Hosts   []ReverseProxyHost `json:"hosts,omitempty"`
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
	ID                       string              `json:"id"`
	Enabled                  bool                `json:"enabled"`
	Host                     string              `json:"host"`
	Architecture             string              `json:"architecture"`
	CpuModel                 string              `json:"cpu_model"`
	OsVersion                string              `json:"os_version,omitempty"`
	OsName                   string              `json:"os_name,omitempty"`
	ExternalIpAddress        string              `json:"external_ip_address,omitempty"`
	DevOpsVersion            string              `json:"devops_version,omitempty"`
	Description              string              `json:"description,omitempty"`
	Tags                     []string            `json:"tags,omitempty"`
	State                    string              `json:"state,omitempty"`
	ParallelsDesktopVersion  string              `json:"parallels_desktop_version,omitempty"`
	ParallelsDesktopLicensed bool                `json:"parallels_desktop_licensed,omitempty"`
	IsReverseProxyEnabled    bool                `json:"is_reverse_proxy_enabled"`
	ReverseProxy             *HostReverseProxy   `json:"reverse_proxy,omitempty"`
	ReverseProxyHosts        []*ReverseProxyHost `json:"reverse_proxy_hosts,omitempty"`
	Resources                HostResourceItem    `json:"resources,omitempty"`
	DetailedResources        *HostResources      `json:"detailed_resources,omitempty"`
	RequiredClaims           []string            `json:"required_claims,omitempty"`
	RequiredRoles            []string            `json:"required_roles,omitempty"`
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
