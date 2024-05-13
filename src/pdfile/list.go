package pdfile

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

func (p *PDFileService) runList(ctx basecontext.ApiContext, url string) (interface{}, *diagnostics.PDFileDiagnostics) {
	ctx.DisableLog()

	diag := diagnostics.NewPDFileDiagnostics()
	client := apiclient.NewHttpClient(ctx)
	if p.pdfile.Authentication != nil {
		authorization := apiclient.HttpClientServiceAuthorization{
			Username: p.pdfile.Authentication.Username,
			Password: p.pdfile.Authentication.Password,
			ApiKey:   p.pdfile.Authentication.ApiKey,
		}
		client.SetAuthorization(authorization)
	}

	var response []map[string][]models.CatalogManifest
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
		output := ""
		maxCatalogIdSize := 0
		maxVersionSize := 0
		for _, catalog := range response {
			for key := range catalog {
				currentCatalogIdLength := len(key) + 2
				if currentCatalogIdLength > maxCatalogIdSize {
					maxCatalogIdSize = currentCatalogIdLength
				}
				for _, catalog := range catalog[key] {
					currentVersionLength := len(catalog.Version) + 2
					if currentVersionLength > maxVersionSize {
						maxVersionSize = currentVersionLength
					}
				}
			}
		}

		headerLine := "| "
		catalogIdHeader := "Catalog ID"
		if maxCatalogIdSize > len("Catalog ID") {
			catalogIdHeader += strings.Repeat(" ", maxCatalogIdSize-len("Catalog ID"))
		}
		headerLine += catalogIdHeader
		headerLine += " | "
		versionHeader := "Version"
		if maxVersionSize > len("Version") {
			versionHeader += strings.Repeat(" ", maxVersionSize-len("Version"))
		}
		headerLine += versionHeader + " "
		headerLine += " | "
		architectureHeader := "Architecture"
		headerLine += architectureHeader
		headerLine += " |\n"
		headerLine1 := "|"
		headerLine1 += strings.Repeat("-", len(headerLine)-3)
		headerLine1 += "|\n"
		output += headerLine
		output += headerLine1
		for _, catalog := range response {
			for _, value := range catalog {
				for _, catalog := range value {
					line := fmt.Sprintf("| %s", catalog.CatalogId)
					if len(catalog.CatalogId) < len(catalogIdHeader)+2 {
						line += strings.Repeat(" ", len(catalogIdHeader)-len(catalog.CatalogId)+1)
					}
					line += fmt.Sprintf("| %s", catalog.Version)
					if len(catalog.Version) < len(versionHeader)+2 {
						line += strings.Repeat(" ", len(versionHeader)-len(catalog.Version)+2)
					}
					line += fmt.Sprintf("| %s ", catalog.Architecture)
					if len(catalog.Architecture) < len(architectureHeader)+2 {
						line += strings.Repeat(" ", len(architectureHeader)-len(catalog.Architecture))
					}
					line += "|\n"
					output += line
				}
			}
		}
		out = []byte(output)
	}
	if err != nil {
		diag.AddError(err)
		return nil, diag
	}

	ctx.EnableLog()
	return string(out), diag
}
