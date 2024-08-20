package pdfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
	"github.com/Parallels/prl-devops-service/security"
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
		AmplitudeEvent: p.pdfile.Client,
	}

	if p.pdfile.Clone {
		body.StartAfterPull = false
	}

	if err := body.Validate(); err != nil {
		diag.AddError(err)
		return nil, diag
	}

	ctx.LogInfof("Pulling catalog machine %v", p.pdfile.MachineName)

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

	ctx.LogInfof("Machine %v pulled successfully", p.pdfile.MachineName)
	if p.pdfile.Clone {
		ctx.LogInfof("Cloning machine %v", p.pdfile.MachineName)
		if p.pdfile.CloneTo == "" {
			cloneName, err := security.GenerateCryptoRandomString(20)
			if err != nil {
				diag.AddError(err)
				return nil, diag
			}

			p.pdfile.CloneTo = cloneName
		}

		provider := serviceprovider.Get()
		if provider.ParallelsDesktopService == nil {
			diag.AddError(errors.New("parallels Desktop service is not available"))
			return nil, diag
		}

		err := provider.ParallelsDesktopService.CloneVm(ctx, resultManifest.MachineID, p.pdfile.CloneTo)
		if err != nil {
			diag.AddError(err)
			return nil, diag
		}
		vm, err := provider.ParallelsDesktopService.GetVmSync(ctx, p.pdfile.CloneTo)
		if err != nil {
			diag.AddError(err)
			return nil, diag
		}

		p.pdfile.CloneId = vm.ID
		ctx.LogInfof("Machine %v cloned successfully to %v", p.pdfile.MachineName, p.pdfile.CloneTo)
	}

	if len(p.pdfile.Execute) > 0 {
		ctx.LogInfof("Executing commands on machine %v", resultManifest.MachineName)
		provider := serviceprovider.Get()
		executeMachine := resultManifest.ID
		if p.pdfile.Clone {
			executeMachine = p.pdfile.CloneTo
		}

		vm, err := provider.ParallelsDesktopService.GetVmSync(ctx, executeMachine)
		if err != nil {
			diag.AddError(err)
			return nil, diag
		}

		if vm.State == "stopped" {
			ctx.LogInfof("Starting machine %v", executeMachine)
			err := provider.ParallelsDesktopService.StartVm(ctx, executeMachine)
			if err != nil {
				diag.AddError(err)
				return nil, diag
			}

			counter := 0
			for {
				resp, err := provider.ParallelsDesktopService.ExecuteCommandOnVm(ctx, executeMachine, &api_models.VirtualMachineExecuteCommandRequest{
					Command: "echo 'Waiting for machine to start'",
				})
				if err == nil && resp.ExitCode == 0 {
					break
				}

				time.Sleep(1 * time.Second)
				counter++
				if counter > 60 {
					diag.AddError(errors.New("Timeout waiting for machine to start"))
					return nil, diag
				}
			}

		}

		for _, command := range p.pdfile.Execute {

			if provider.ParallelsDesktopService == nil {
				diag.AddError(errors.New("parallels Desktop service is not available"))
				return nil, diag
			}
			commandRequest := api_models.VirtualMachineExecuteCommandRequest{
				Command: command,
			}
			r, err := provider.ParallelsDesktopService.ExecuteCommandOnVm(ctx, executeMachine, &commandRequest)
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

	if p.pdfile.Clone {
		response.MachineId = p.pdfile.CloneId
		response.MachineName = p.pdfile.CloneTo
		if p.pdfile.StartAfterPull {
			provider := serviceprovider.Get()
			err := provider.ParallelsDesktopService.StartVm(ctx, p.pdfile.CloneTo)
			if err != nil {
				diag.AddError(err)
				return nil, diag
			}
		}
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
