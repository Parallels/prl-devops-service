package pdfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

func (p *PDFileService) runPull(ctx basecontext.ApiContext) (interface{}, *diagnostics.PDFileDiagnostics) {
	ctx.DisableLog()
	serviceprovider.InitServices(ctx)

	diag := diagnostics.NewPDFileDiagnostics()

	if !p.pdfile.HasAuthentication() {
		diag.AddError(errors.New("Username and password or apikey are required for authentication"))
		return nil, diag
	}

	body := catalog_models.PullCatalogManifestRequest{
		CatalogId:      p.pdfile.CatalogId,
		Version:        p.pdfile.Version,
		Architecture:   p.pdfile.Architecture,
		Owner:          p.pdfile.Owner,
		MachineName:    p.pdfile.MachineName,
		Path:           p.pdfile.Destination,
		Connection:     p.pdfile.GetHostConnection(),
		StartAfterPull: p.pdfile.StartAfterPull,
	}

	if err := body.Validate(); err != nil {
		diag.AddError(err)
		return nil, diag
	}

	manifest := catalog.NewManifestService(ctx)
	resultManifest := manifest.Pull(ctx, &body)
	if resultManifest.HasErrors() {
		errorMessage := "Error pulling manifest:"
		for _, err := range resultManifest.Errors {
			errorMessage += "\n" + err.Error() + " "
		}

		diag.AddError(errors.New(errorMessage))
		return nil, diag
	}

	if len(p.pdfile.Execute) > 0 {
		for _, command := range p.pdfile.Execute {
			provider := serviceprovider.Get()
			if provider.ParallelsDesktopService == nil {
				diag.AddError(errors.New("parallels Desktop service is not available"))
				return nil, diag
			}
			commandRequest := api_models.VirtualMachineExecuteCommandRequest{
				Command: command,
			}
			r, err := provider.ParallelsDesktopService.ExecuteCommandOnVm(ctx, resultManifest.ID, &commandRequest)
			if err != nil {
				diag.AddError(err)
				return nil, diag
			}
			if r.ExitCode != 0 {
				diag.AddError(fmt.Errorf("Error executing command: %v, err: %v"+command, r.Error))
				return nil, diag
			}
		}
	}

	response := models.PullResponse{
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
