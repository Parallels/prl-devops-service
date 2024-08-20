package models

import (
	"fmt"
	"strings"
)

type PDFile struct {
	Raw            []string              `json:"-" yaml:"-"`
	Insecure       bool                  `json:"-" yaml:"-"`
	Host           string                `json:"-" yaml:"-"`
	From           string                `json:"FROM,omitempty" yaml:"FROM,omitempty"`
	To             string                `json:"TO,omitempty" yaml:"TO,omitempty"`
	Prefix         string                `json:"PREFIX,omitempty" yaml:"PREFIX,omitempty"`
	Authentication *PDFileAuthentication `json:"AUTHENTICATION,omitempty" yaml:"AUTHENTICATION,omitempty"`
	Description    string                `json:"DESCRIPTION,omitempty" yaml:"DESCRIPTION,omitempty"`
	CatalogId      string                `json:"CATALOG_ID,omitempty" yaml:"CATALOG_ID,omitempty"`
	Version        string                `json:"VERSION,omitempty" yaml:"VERSION,omitempty"`
	Architecture   string                `json:"ARCHITECTURE,omitempty" yaml:"ARCHITECTURE,omitempty"`
	LocalPath      string                `json:"LOCAL_PATH,omitempty" yaml:"LOCAL_PATH,omitempty"`
	Destination    string                `json:"DESTINATION,omitempty" yaml:"DESTINATION,omitempty"`
	MachineName    string                `json:"MACHINE_NAME,omitempty" yaml:"MACHINE_NAME,omitempty"`
	Owner          string                `json:"OWNER,omitempty" yaml:"OWNER,omitempty"`
	StartAfterPull bool                  `json:"START_AFTER_PULL,omitempty" yaml:"START_AFTER_PULL,omitempty"`
	Roles          []string              `json:"ROLES,omitempty" yaml:"ROLES,omitempty"`
	Claims         []string              `json:"CLAIMS,omitempty" yaml:"CLAIMS,omitempty"`
	Tags           []string              `json:"TAGS,omitempty" yaml:"TAGS,omitempty"`
	Provider       *PDFileProvider       `json:"PROVIDER,omitempty" yaml:"PROVIDER,omitempty"`
	Command        string                `json:"COMMAND,omitempty" yaml:"COMMAND,omitempty"`
	Execute        []string              `json:"EXECUTE,omitempty" yaml:"EXECUTE,omitempty"`
	Clone          bool                  `json:"CLONE,omitempty" yaml:"CLONE,omitempty"`
	CloneTo        string                `json:"CLONE_TO,omitempty" yaml:"CLONE_TO,omitempty"`
	CloneId        string                `json:"CLONE_ID,omitempty" yaml:"CLONE_ID,omitempty"`
	Operation      string                `json:"RUN,omitempty" yaml:"RUN,omitempty"`
	Client         string                `json:"CLIENT,omitempty" yaml:"CLIENT,omitempty"`
}

func NewPdFile() *PDFile {
	return &PDFile{
		Raw:     []string{},
		Roles:   []string{},
		Claims:  []string{},
		Tags:    []string{},
		Execute: []string{},
	}
}

func (p *PDFile) HasAuthentication() bool {
	if p.Authentication == nil {
		return false
	}
	if p.Authentication.ApiKey != "" || (p.Authentication.Username != "" && p.Authentication.Password != "") {
		return true
	}

	return false
}

func (p *PDFile) ParseProvider(value string) (PDFileProvider, error) {
	result := PDFileProvider{}
	if result.Attributes == nil {
		result.Attributes = make(map[string]string)
	}
	separator := " "
	if strings.Contains(value, ";") {
		separator = ";"
	}

	providerParts := strings.Split(value, separator)
	if len(providerParts) == 1 {
		result.Attributes[strings.ToLower(value)] = value
		return result, nil
	}

	for i, providerPart := range providerParts {
		if i == 0 && (strings.HasPrefix(strings.ToLower(providerPart), "provider") ||
			strings.HasPrefix(strings.ToLower(providerPart), "name")) {
			providerNameParts := strings.Split(providerPart, " ")
			if len(providerNameParts) == 2 {
				result.Name = strings.TrimSpace(providerNameParts[1])
				continue
			} else {
				providerNameParts = strings.Split(providerPart, "=")
				if len(providerNameParts) == 2 {
					result.Name = strings.TrimSpace(providerNameParts[1])
					continue
				}
			}
		}

		if strings.Contains(providerPart, "=") {
			providerSegmentParts := strings.Split(providerPart, "=")
			if len(providerSegmentParts) == 2 {
				if strings.ToLower(providerSegmentParts[0]) == "name" {
					result.Name = strings.TrimSpace(providerSegmentParts[1])
					continue
				}

				result.Attributes[strings.ToLower(providerSegmentParts[0])] = strings.TrimSpace(providerSegmentParts[1])
			} else {
				if len(providerParts) == 2 {
					if strings.ToLower(providerParts[0]) == "name" {
						result.Name = strings.TrimSpace(providerParts[1])
						continue
					}

					result.Attributes[strings.ToLower(providerParts[0])] = strings.TrimSpace(providerParts[1])
				}
				if len(providerParts) == 3 {
					if strings.ToLower(providerParts[1]) == "name" {
						result.Name = strings.TrimSpace(providerParts[2])
						continue
					}
					result.Attributes[strings.ToLower(providerParts[1])] = strings.TrimSpace(providerParts[2])
				}
			}
		} else if len(providerParts) == 3 {
			if strings.ToLower(providerParts[1]) == "name" {
				result.Name = strings.TrimSpace(providerParts[2])
				continue
			}

			result.Attributes[strings.ToLower(providerParts[1])] = strings.TrimSpace(providerParts[2])
		}
	}

	return result, nil
}

func (p *PDFile) GetHostUrl() string {
	host := ""
	if p == nil {
		return host
	}
	if p.Host == "" && (p.From != "" || p.To != "") {
		if p.From != "" {
			p.Host = p.From
		} else {
			p.Host = p.To
		}
	}

	if p.Insecure || (p.From == "localhost" || p.To == "localhost") {
		host = fmt.Sprintf("http://%s", p.Host)
	} else {
		host = fmt.Sprintf("https://%s", p.Host)
	}
	prefix := "api"
	if p.Prefix != "" {
		prefix = strings.ReplaceAll(p.Prefix, "/", "")
	}
	host = fmt.Sprintf("%s/%s/catalog", host, prefix)

	return host
}

func (p *PDFile) GetHostCatalogUrl() string {
	host := p.GetHostUrl()

	if p.CatalogId != "" {
		host = fmt.Sprintf("%s/%s", host, p.CatalogId)
		if p.Version != "" {
			host = fmt.Sprintf("%s/%s", host, p.Version)
			if p.Architecture != "" {
				host = fmt.Sprintf("%s/%s", host, p.Architecture)
			}
		}
	}

	return host
}

func (p *PDFile) GetHostConnection() string {
	host := ""
	if p.Authentication != nil {
		if p.Authentication.ApiKey != "" {
			host = fmt.Sprintf("%s@%s", p.Authentication.ApiKey, p.Host)
		} else if p.Authentication.Username != "" && p.Authentication.Password != "" {
			host = fmt.Sprintf("%s:%s@%s", p.Authentication.Username, p.Authentication.Password, p.Host)
		}
	} else {
		host = p.Host
	}

	return fmt.Sprintf("host=%v", host)
}

func (p *PDFile) GetProviderConnectionString() string {
	provider := ""
	if p.Provider != nil {
		provider = fmt.Sprintf("provider=%v", p.Provider.Name)

		for key, value := range p.Provider.Attributes {
			provider += fmt.Sprintf(";%s=%s", key, value)
		}
	}

	return provider
}

func (p *PDFile) GetConnectionString() string {
	result := ""
	provider := p.GetProviderConnectionString()

	host := p.GetHostConnection()
	if host != "" {
		result = host
	}

	if provider != "" {
		if result != "" {
			result = fmt.Sprintf("%s;%s", result, provider)
		} else {
			result = provider
		}
	}

	return result
}
