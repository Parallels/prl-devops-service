package pdfile

import (
	"encoding/json"
	"errors"
	"fmt"
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
	ctx.DisableLog()

	progressChannel := make(chan int)
	fileNameChannel := make(chan string)
	stepChannel := make(chan string)

	defer close(progressChannel)
	defer close(fileNameChannel)

	progress := 0
	currentProgress := 0
	fileName := ""

	diag := diagnostics.NewPDFileDiagnostics()
	body := models.PushCatalogManifestRequest{
		CatalogId:       p.pdfile.CatalogId,
		Version:         p.pdfile.Version,
		Architecture:    p.pdfile.Architecture,
		LocalPath:       p.pdfile.LocalPath,
		RequiredRoles:   p.pdfile.Roles,
		RequiredClaims:  p.pdfile.Claims,
		Description:     p.pdfile.Description,
		Tags:            p.pdfile.Tags,
		CompressPack:    p.pdfile.CompressPack,
		ProgressChannel: progressChannel,
		FileNameChannel: fileNameChannel,
		StepChannel:     stepChannel,
		Connection:      p.pdfile.GetConnectionString(),
	}

	if p.pdfile.MinimumSpecRequirements != nil {
		body.MinimumSpecRequirements = models.MinimumSpecRequirement{
			Cpu:    p.pdfile.MinimumSpecRequirements.Cpu,
			Memory: p.pdfile.MinimumSpecRequirements.Memory,
			Disk:   p.pdfile.MinimumSpecRequirements.Disk,
		}
	}

	go func() {
		for {
			fileName = <-fileNameChannel
		}
	}()

	go func() {
		for {
			step := <-stepChannel
			clearLine()
			fmt.Printf("\r%s", step)
		}
	}()

	go func() {
		for {
			currentProgress = <-progressChannel
			if currentProgress > progress {
				progress = currentProgress
				clearLine()
				fmt.Printf("\rUploading %s: %d%%", fileName, progress)
			}
		}
	}()

	manifest := catalog.NewManifestService(ctx)
	resultManifest := manifest.Push(&body)
	if resultManifest.HasErrors() {
		errorMessage := "Error pushing manifest:"
		for _, err := range resultManifest.Errors {
			errorMessage += "\n" + err.Error() + " "
		}
		diag.AddError(errors.New(errorMessage))
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

	clearLine()
	fmt.Printf("\rFinished pushing manifest\n")
	ctx.EnableLog()
	return string(out), diag
}

func clearLine() {
	fmt.Printf("\r\033[K")
}
