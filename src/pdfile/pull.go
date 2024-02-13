package pdfile

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

func (p *PDFile) runPull(ctx basecontext.ApiContext) (interface{}, *PDFileDiagnostics) {
	ctx.DisableLog()
	serviceprovider.InitServices(ctx)

	diag := NewPDFileDiagnostics()

	body := models.PullCatalogManifestRequest{
		CatalogId:      p.CatalogId,
		Version:        p.Version,
		Architecture:   p.Architecture,
		Owner:          p.Owner,
		MachineName:    p.MachineName,
		Path:           p.Destination,
		Connection:     p.GetHostConnection(),
		StartAfterPull: p.StartAfterPull,
	}

	if err := body.Validate(); err != nil {
		diag.AddError(err)
		return nil, diag
	}

	manifest := catalog.NewManifestService(ctx)
	resultManifest := manifest.Pull(ctx, &body)
	if resultManifest.HasErrors() {
		errorMessage := "Error pulling manifest: \n"
		for _, err := range resultManifest.Errors {
			errorMessage += "\n" + err.Error() + " "
		}

		diag.AddError(errors.New(errorMessage))
		return nil, diag
	}

	response := PullResponse{
		MachineId:    resultManifest.ID,
		MachineName:  resultManifest.MachineName,
		CatalogId:    resultManifest.Manifest.CatalogId,
		Version:      resultManifest.Manifest.Version,
		Architecture: resultManifest.Manifest.Architecture,
		Type:         resultManifest.Manifest.Type,
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
