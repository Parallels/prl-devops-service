package pdfile

import (
	"encoding/json"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

func (p *PDFileService) runPush(ctx basecontext.ApiContext) (interface{}, *diagnostics.PDFileDiagnostics) {
	ctx.LogInfof("Starting push...")
	ctx.DisableLog()

	diag := diagnostics.NewPDFileDiagnostics()
	body := models.PushCatalogManifestRequest{
		CatalogId:      p.pdfile.CatalogId,
		Version:        p.pdfile.Version,
		Architecture:   p.pdfile.Architecture,
		LocalPath:      p.pdfile.LocalPath,
		RequiredRoles:  p.pdfile.Roles,
		RequiredClaims: p.pdfile.Claims,
		Description:    p.pdfile.Description,
		Tags:           p.pdfile.Tags,
		Connection:     p.pdfile.GetConnectionString(),
	}

	manifest := catalog.NewManifestService(ctx)
	resultManifest := manifest.Push(ctx, &body)
	if resultManifest.HasErrors() {
		errorMessage := "Error pushing manifest:"
		for _, err := range resultManifest.Errors {
			errorMessage += "\n" + err.Error() + " "
		}
		return nil, diag
	}

	resultData := mappers.DtoCatalogManifestToApi(mappers.CatalogManifestToDto(*resultManifest))
	resultData.ID = resultManifest.ID

	var out []byte
	var err error

	format := helper.GetFlagValue(constants.PD_FILE_OUTPUT_FLAG, "")
	switch strings.ToLower(format) {
	case "json":
		out, err = json.MarshalIndent(resultData, "", "  ")
	case "yaml":
		out, err = yaml.Marshal(resultData)
	default:
		out, err = json.MarshalIndent(resultData, "", "  ")
	}
	if err != nil {
		diag.AddError(err)
		return nil, diag
	}

	ctx.EnableLog()
	return string(out), diag
}
