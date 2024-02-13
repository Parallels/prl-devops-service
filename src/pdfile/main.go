package pdfile

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

func (PDFile) ParseProvider(value string) (PDFileProvider, error) {
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
		if i == 0 && strings.EqualFold(providerPart, "name") || strings.EqualFold(providerPart, "provider") {
			result.Name = strings.TrimSpace(providerPart)
			continue
		}

		providerSegmentParts := strings.Split(providerPart, "=")
		if len(providerSegmentParts) == 2 {
			result.Attributes[strings.ToLower(providerSegmentParts[0])] = strings.TrimSpace(providerSegmentParts[1])
		} else {
			if len(providerParts) == 2 {
				result.Attributes[strings.ToLower(providerParts[0])] = strings.TrimSpace(providerParts[1])
			}
		}
	}

	return result, nil
}

func (p *PDFile) GetHostUrl() string {
	host := ""
	if p.Insecure {
		host = fmt.Sprintf("http://%s", p.From)
	} else {
		host = fmt.Sprintf("https://%s", p.From)
	}
	host = fmt.Sprintf("%s/api/catalog", host)

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
			host = fmt.Sprintf("%s@%s", p.Authentication.ApiKey, p.From)
		} else if p.Authentication.Username != "" && p.Authentication.Password != "" {
			host = fmt.Sprintf("%s:%s@%s", p.Authentication.Username, p.Authentication.Password, p.From)
		}
	} else {
		host = p.From
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
		if len(result) > 0 {
			result = fmt.Sprintf("%s;%s", result, provider)
		} else {
			result = provider
		}
	}

	return result
}

func (p *PDFile) Run(ctx basecontext.ApiContext) (interface{}, *PDFileDiagnostics) {
	diag := NewPDFileDiagnostics()

	if strings.EqualFold(p.Command, "list") {
		url := p.GetHostCatalogUrl()
		out, runDiag := p.runList(ctx, url)
		diag.Append(runDiag)
		return out, diag
	}

	if strings.EqualFold(p.Command, "push") {
		out, runDiag := p.runPush(ctx)
		diag.Append(runDiag)
		return out, diag
	}

	if strings.EqualFold(p.Command, "pull") {
		out, runDiag := p.runPull(ctx)
		diag.Append(runDiag)
		return out, diag
	}

	return nil, diag
}

func (p *PDFile) runList(ctx basecontext.ApiContext, url string) (interface{}, *PDFileDiagnostics) {
	ctx.DisableLog()

	diag := NewPDFileDiagnostics()
	client := apiclient.NewHttpClient(ctx)
	if p.Authentication != nil {
		authorization := apiclient.HttpClientServiceAuthorization{
			Username: p.Authentication.Username,
			Password: p.Authentication.Password,
			ApiKey:   p.Authentication.ApiKey,
		}
		client.SetAuthorization(authorization)
	}

	var response interface{}
	_, err := client.Get(url, &response)
	if err != nil {
		diag.AddError(err)
		return nil, diag
	}

	var out []byte

	format := helper.GetFlagValue(constants.PD_FILE_OUTPUT_FLAG, "")
	switch strings.ToLower(format) {
	case "json":
		out, err = json.MarshalIndent(response, "", "  ")
	case "yaml":
		out, err = yaml.Marshal(response)
	default:
		out, err = json.MarshalIndent(response, "", "  ")
	}
	if err != nil {
		diag.AddError(err)
		return nil, diag
	}

	ctx.EnableLog()
	return string(out), diag
}
