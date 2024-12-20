package pdfile

import (
	"encoding/base64"
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
	"github.com/Parallels/prl-devops-service/notifications"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
	"github.com/Parallels/prl-devops-service/security"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/telemetry"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

func (p *PDFileService) runPull(ctx basecontext.ApiContext) (interface{}, *diagnostics.PDFileDiagnostics) {
	ctx.DisableLog()
	serviceprovider.InitServices(ctx)
	ns := notifications.Get()
	ns.EnableSingleLineOutput()

	diag := diagnostics.NewPDFileDiagnostics()

	if !p.pdfile.HasAuthentication() {
		diag.AddError(errors.New("username and password or apikey are required for authentication"))
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
	sendTelemetry := false
	var amplitudeEvent api_models.AmplitudeEvent
	telemetryItem := telemetry.TelemetryItem{}
	if body.AmplitudeEvent != "" {
		decodedBytes, err := base64.StdEncoding.DecodeString(body.AmplitudeEvent)
		if err == nil {
			err := json.Unmarshal(decodedBytes, &amplitudeEvent)
			if err == nil {
				telemetryItem.Type = amplitudeEvent.EventType
				if telemetryItem.Type == "" {
					telemetryItem.Type = "DEVOPS:PULL_MANIFEST"
				}
				telemetryItem.Properties = amplitudeEvent.EventProperties
				telemetryItem.Options = amplitudeEvent.UserProperties
				telemetryItem.UserID = amplitudeEvent.AppId
				telemetryItem.DeviceId = amplitudeEvent.DeviceId
				if amplitudeEvent.Origin != "" {
					telemetryItem.Properties["origin"] = amplitudeEvent.Origin
				}
				sendTelemetry = true
			} else {
				ctx.LogErrorf("Error unmarshalling amplitude event", err)
			}
		} else {
			ctx.LogErrorf("Error decoding amplitude event", err)
		}
	}

	resultManifest := manifest.Pull(&body)
	if resultManifest.HasErrors() {
		errorMessage := "Error pulling manifest:"
		for _, err := range resultManifest.Errors {
			errorMessage += "\n" + err.Error() + " "
		}

		diag.AddError(errors.New(errorMessage))
		if sendTelemetry {
			sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
		}
		return nil, diag
	}

	fmt.Printf("\n")
	fmt.Printf("Successfully pulled machine %v\n", p.pdfile.MachineName)
	if p.pdfile.Clone {
		clearLine()
		fmt.Printf("\rCloning machine %v", p.pdfile.MachineName)
		if p.pdfile.CloneTo == "" {
			cloneName, err := security.GenerateCryptoRandomString(20)
			if err != nil {
				diag.AddError(err)
				if sendTelemetry {
					sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
				}
				return nil, diag
			}

			p.pdfile.CloneTo = cloneName
		}

		provider := serviceprovider.Get()
		if provider.ParallelsDesktopService == nil {
			diag.AddError(errors.New("parallels Desktop service is not available"))
			if sendTelemetry {
				sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
			}
			return nil, diag
		}

		err := provider.ParallelsDesktopService.CloneVm(ctx, resultManifest.MachineID, p.pdfile.CloneTo)
		if err != nil {
			diag.AddError(err)
			if sendTelemetry {
				sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
			}
			return nil, diag
		}
		vm, err := provider.ParallelsDesktopService.GetVmSync(ctx, p.pdfile.CloneTo)
		if err != nil {
			diag.AddError(err)
			if sendTelemetry {
				sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
			}
			return nil, diag
		}

		p.pdfile.CloneId = vm.ID
		ctx.LogInfof("Machine %v cloned successfully to %v", p.pdfile.MachineName, p.pdfile.CloneTo)
	}

	if len(p.pdfile.Execute) > 0 {
		clearLine()
		fmt.Printf("\rExecuting commands on machine %v", resultManifest.MachineName)
		provider := serviceprovider.Get()
		executeMachine := resultManifest.ID
		if p.pdfile.Clone {
			executeMachine = p.pdfile.CloneTo
		}

		vm, err := provider.ParallelsDesktopService.GetVmSync(ctx, executeMachine)
		if err != nil {
			diag.AddError(err)
			if sendTelemetry {
				sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
			}
			return nil, diag
		}

		if vm.State == "stopped" {
			ctx.LogInfof("Starting machine %v", executeMachine)
			err := provider.ParallelsDesktopService.StartVm(ctx, executeMachine)
			if err != nil {
				diag.AddError(err)
				if sendTelemetry {
					sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
				}
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
					diag.AddError(errors.New("timeout waiting for machine to start"))
					if sendTelemetry {
						sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
					}
					return nil, diag
				}
			}

		}

		for _, command := range p.pdfile.Execute {
			clearLine()
			fmt.Printf("\rExecuting command %v on machine %v", command, resultManifest.MachineName)
			if provider.ParallelsDesktopService == nil {
				diag.AddError(errors.New("parallels Desktop service is not available"))
				if sendTelemetry {
					sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
				}
				return nil, diag
			}
			commandRequest := api_models.VirtualMachineExecuteCommandRequest{
				Command: command,
			}
			r, err := provider.ParallelsDesktopService.ExecuteCommandOnVm(ctx, executeMachine, &commandRequest)
			if err != nil {
				diag.AddError(err)
				if sendTelemetry {
					sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
				}
				return nil, diag
			}
			if r.ExitCode != 0 {
				diag.AddError(fmt.Errorf("Error executing command: %v, err: %v"+command, r.Error))
				if sendTelemetry {
					sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
				}
				return nil, diag
			}
		}
	}

	fmt.Printf("\rFinished pulling manifest\n")
	response := models.PullResponse{
		MachineId:      resultManifest.ID,
		MachineName:    resultManifest.MachineName,
		CatalogId:      resultManifest.Manifest.CatalogId,
		Version:        resultManifest.Manifest.Version,
		Architecture:   resultManifest.Manifest.Architecture,
		LocalCachePath: resultManifest.LocalCachePath,
		Type:           resultManifest.Manifest.Type,
	}

	if p.pdfile.Clone {
		response.MachineId = p.pdfile.CloneId
		response.MachineName = p.pdfile.CloneTo
		if p.pdfile.StartAfterPull {
			provider := serviceprovider.Get()
			err := provider.ParallelsDesktopService.StartVm(ctx, p.pdfile.CloneTo)
			if err != nil {
				diag.AddError(err)
				if sendTelemetry {
					sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
				}
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
		if sendTelemetry {
			sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
		}
		return nil, diag
	}

	if sendTelemetry {
		sendTelemetryEvent(amplitudeEvent, telemetryItem, diag)
	}

	ctx.EnableLog()
	return string(out), diag
}

func sendTelemetryEvent(amplitudeEvent api_models.AmplitudeEvent, telemetryItem telemetry.TelemetryItem, diag *diagnostics.PDFileDiagnostics) {
	if amplitudeEvent.EventProperties == nil {
		telemetryItem.Properties["success"] = "true"
		telemetry.TrackEvent(telemetryItem)
		telemetry.Get().Flush()
		return
	}

	if diag == nil {
		telemetryItem.Properties["success"] = "true"
		telemetry.TrackEvent(telemetryItem)
		telemetry.Get().Flush()
		return
	}

	if diag.HasErrors() {
		if amplitudeEvent.EventProperties != nil {
			telemetryItem.Properties["success"] = "false"
			telemetry.TrackEvent(telemetryItem)
			telemetry.Get().Flush()
		}
	} else {
		if amplitudeEvent.EventProperties != nil {
			telemetryItem.Properties["success"] = "true"
			telemetry.TrackEvent(telemetryItem)
			telemetry.Get().Flush()
		}
	}
}
