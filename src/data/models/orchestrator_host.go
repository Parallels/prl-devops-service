package models

import (
	"net/url"
	"strings"

	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

type OrchestratorHost struct {
	ID                        string                          `json:"id" gorm:"primaryKey;column:id;type:varchar(255)"`
	Enabled                   bool                            `json:"enabled" gorm:"column:enabled;type:boolean;default:false;not null"`
	Host                      string                          `json:"host" gorm:"column:host;type:varchar(255)"`
	Architecture              string                          `json:"architecture" gorm:"column:architecture;type:varchar(255)"`
	CpuModel                  string                          `json:"cpu_model,omitempty" gorm:"column:cpu_model;type:varchar(255)"`
	OsVersion                 string                          `json:"os_version,omitempty" gorm:"column:os_version;type:varchar(255)"`
	OsName                    string                          `json:"os_name,omitempty" gorm:"column:os_name;type:varchar(255)"`
	ExternalIpAddress         string                          `json:"external_ip_address,omitempty" gorm:"column:external_ip_address;type:varchar(255)"`
	DevOpsVersion             string                          `json:"devops_version,omitempty" gorm:"column:devops_version;type:varchar(255)"`
	Description               string                          `json:"description,omitempty" gorm:"column:description;type:text"`
	Tags                      []string                        `json:"tags,omitempty" gorm:"column:tags;type:json;serializer:json"`
	Port                      string                          `json:"port,omitempty" gorm:"column:port;type:varchar(255)"`
	Schema                    string                          `json:"schema,omitempty" gorm:"column:schema;type:varchar(255)"`
	PathPrefix                string                          `json:"path_prefix,omitempty" gorm:"column:path_prefix;type:varchar(255)"`
	ParallelsDesktopVersion   string                          `json:"parallels_desktop_version,omitempty" gorm:"column:parallels_desktop_version;type:varchar(255)"`
	ParallelsDesktopLicensed  bool                            `json:"parallels_desktop_licensed,omitempty" gorm:"column:parallels_desktop_licensed;type:boolean;default:false;not null"`
	HasWebsocketEvents        bool                            `json:"has_websocket_events" gorm:"column:has_websocket_events;type:boolean;default:false;not null"`
	IsLocal                   bool                            `json:"is_local,omitempty" gorm:"column:is_local;type:boolean;default:false;not null"`
	Authentication            *OrchestratorHostAuthentication `json:"authentication,omitempty" gorm:"column:authentication;type:json;serializer:json"`
	Resources                 *HostResources                  `json:"resources,omitempty" gorm:"column:resources;type:json;serializer:json"`
	State                     string                          `json:"state,omitempty" gorm:"column:state;type:varchar(255)"`
	LastUnhealthy             string                          `json:"last_seen,omitempty" gorm:"column:last_seen;type:varchar(255)"`
	LastUnhealthyErrorMessage string                          `json:"last_seen_error_message,omitempty" gorm:"column:last_seen_error_message;type:varchar(255)"`
	HealthCheck               *models.ApiHealthCheck          `json:"-" gorm:"-"`
	VirtualMachines           []VirtualMachine                `json:"virtual_machines,omitempty" gorm:"foreignKey:HostId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	IsReverseProxyEnabled     bool                            `json:"is_reverse_proxy_enabled,omitempty" gorm:"column:is_reverse_proxy_enabled;type:boolean;default:false;not null"`
	IsLogStreamingEnabled     bool                            `json:"is_log_streaming_enabled,omitempty" gorm:"column:is_log_streaming_enabled;type:boolean;default:false;not null"`
	EnabledModules            []string                        `json:"enabled_modules,omitempty" gorm:"column:enabled_modules;type:json;serializer:json"`
	ReverseProxy              *ReverseProxy                   `json:"reverse_proxy,omitempty" gorm:"column:reverse_proxy;type:json;serializer:json"`
	ReverseProxyHosts         []*ReverseProxyHost             `json:"reverse_proxy_hosts,omitempty" gorm:"-"`
	CacheConfig               *models.CatalogCacheConfig      `json:"cache_config,omitempty" gorm:"column:cache_config;type:json;serializer:json"`
	CacheItems                []models.HostCatalogCacheItem   `json:"cache_items,omitempty" gorm:"column:cache_items;type:json;serializer:json"`
	CreatedAt                 string                          `json:"created_at,omitempty" gorm:"column:created_at;type:timestamp"`
	UpdatedAt                 string                          `json:"updated_at,omitempty" gorm:"column:updated_at;type:timestamp"`
	RequiredClaims            []string                        `json:"required_claims,omitempty" gorm:"column:required_claims;type:json;serializer:json"`
	RequiredRoles             []string                        `json:"required_roles,omitempty" gorm:"column:required_roles;type:json;serializer:json"`
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

func (o OrchestratorHost) GetHostUrl() string {
	url, err := url.Parse(o.Host)
	if err != nil {
		return ""
	} else {
		return url.Path
	}
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
		found := false
		for _, vm2 := range source.VirtualMachines {
			if vm.ID == vm2.ID {
				found = true
				if vm.Diff(vm2) {
					return true
				}
				break
			}
		}
		if !found {
			return true
		}
	}

	if o.IsReverseProxyEnabled != source.IsReverseProxyEnabled {
		return true
	}

	if o.IsLogStreamingEnabled != source.IsLogStreamingEnabled {
		return true
	}

	if len(o.EnabledModules) != len(source.EnabledModules) {
		return true
	}
	for i, m := range o.EnabledModules {
		if m != source.EnabledModules[i] {
			return true
		}
	}

	if o.HasWebsocketEvents != source.HasWebsocketEvents {
		return true
	}

	if o.IsLocal != source.IsLocal {
		return true
	}

	if o.ReverseProxy != nil && source.ReverseProxy == nil {
		return true
	}

	if o.ReverseProxy == nil && source.ReverseProxy != nil {
		return true
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

	if o.CacheConfig != nil && source.CacheConfig == nil {
		return true
	}
	if o.CacheConfig == nil && source.CacheConfig != nil {
		return true
	}
	if o.CacheConfig != nil && source.CacheConfig != nil {
		if o.CacheConfig.Enabled != source.CacheConfig.Enabled || o.CacheConfig.MaxSize != source.CacheConfig.MaxSize {
			return true
		}
	}

	if len(o.CacheItems) != len(source.CacheItems) {
		return true
	}

	for _, item := range o.CacheItems {
		found := false
		for _, sourceItem := range source.CacheItems {
			if item.CatalogId == sourceItem.CatalogId && item.Version == sourceItem.Version {
				found = true
				break
			}
		}
		if !found {
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
