package services

import (
	"Parallels/pd-api-service/helpers"
	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/common-go/commands"
)

var globalPackerService *PackerService

type PackerService struct {
	executable string
}

func NewPackerService() *PackerService {
	if globalPackerService != nil {
		return globalPackerService
	}

	globalPackerService = &PackerService{}

	globalServices.Logger.Info("Getting packer executable")
	out, err := commands.ExecuteWithNoOutput("which", "packer")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		globalServices.Logger.Warn("Packer executable not found, trying to find it in the default locations")
	}

	if path != "" {
		globalPackerService.executable = path
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/packer"); err == nil {
			globalPackerService.executable = "/opt/homebrew/bin/packer"
		} else {
			globalServices.Logger.Warn("Packer executable not found")
			return nil
		}

		globalServices.Logger.Info("Packer executable: " + globalPackerService.executable)
	}

	return globalPackerService
}

func (s *PackerService) Build(path string, variableFile string) error {
	stdout, err := helpers.ExecuteWithNoOutput(helpers.Command{
		Command:          s.executable,
		WorkingDirectory: path,
		Args:             []string{"build", "-var-file", variableFile, "."},
	})
	if err != nil {
		println(stdout)
		buildError := fmt.Errorf("There was an error building packer folder %v, error: %v", path, err.Error())
		return buildError
	}
	return nil
}
