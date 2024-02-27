package pdfile

import (
	"encoding/json"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
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
