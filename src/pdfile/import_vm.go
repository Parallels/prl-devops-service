package pdfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

func (p *PDFileService) runImportVM(ctx basecontext.ApiContext) (interface{}, *diagnostics.PDFileDiagnostics) {
	ctx.DisableLog()
	serviceprovider.InitServices(ctx)

	diag := diagnostics.NewPDFileDiagnostics()

	body := catalog_models.ImportVmRequest{
		CatalogId:         p.pdfile.CatalogId,
		Version:           p.pdfile.Version,
		Architecture:      p.pdfile.Architecture,
		Description:       p.pdfile.Description,
		IsCompressed:      p.pdfile.IsCompressed,
		Type:              p.pdfile.VMType,
		Size:              p.pdfile.VMSize,
		MachineRemotePath: p.pdfile.VMRemotePath,
		Tags:              p.pdfile.Tags,
		RequiredClaims:    p.pdfile.Claims,
		RequiredRoles:     p.pdfile.Roles,
		Force:             p.pdfile.Force,
		Connection:        p.pdfile.GetProviderConnectionString(),
	}

	if err := body.Validate(); err != nil {
		diag.AddError(err)
		return nil, diag
	}

	client := apiclient.NewHttpClient(ctx)
	if p.pdfile.Authentication != nil {
		if p.pdfile.Authentication.ApiKey != "" {
			client.AuthorizeWithApiKey(p.pdfile.Authentication.ApiKey)
		}
		if p.pdfile.Authentication.Username != "" && p.pdfile.Authentication.Password != "" {
			client.AuthorizeWithUsernameAndPassword(p.pdfile.Authentication.Username, p.pdfile.Authentication.Password)
		}
	}

	host := p.pdfile.GetHostUrl() + "/import-vm"
	var response catalog_models.ImportVmResponse
	apiResponse, apiErr := client.Put(host, &body, &response)
	if apiErr != nil {
		diag.AddError(apiErr)
		return nil, diag
	}
	if apiResponse.StatusCode != 200 {
		diag.AddError(errors.New("Error pulling manifest: " + fmt.Sprintf("%v", apiResponse.StatusCode)))
		return nil, diag
	}

	var out []byte
	var err error

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
