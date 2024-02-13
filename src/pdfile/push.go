package pdfile

import (
	"encoding/json"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog"
	"github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

func (p *PDFile) runPush(ctx basecontext.ApiContext) (interface{}, *PDFileDiagnostics) {
	ctx.LogInfof("Starting push...")
	ctx.DisableLog()

	diag := NewPDFileDiagnostics()
	body := models.PushCatalogManifestRequest{
		CatalogId:      p.CatalogId,
		Version:        p.Version,
		Architecture:   p.Architecture,
		LocalPath:      p.LocalPath,
		RequiredRoles:  p.Roles,
		RequiredClaims: p.Claims,
		Description:    p.Description,
		Tags:           p.Tags,
		Connection:     p.GetConnectionString(),
	}

	manifest := catalog.NewManifestService(ctx)
	resultManifest := manifest.Push(ctx, &body)
	if resultManifest.HasErrors() {
		errorMessage := "Error pushing manifest: \n"
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
