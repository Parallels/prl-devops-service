package vagrant

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/errors"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/service_provider/interfaces"

	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/common-go/commands"
)

var globalVagrantService *VagrantService
var logger = common.Logger

type VagrantService struct {
	executable   string
	installed    bool
	dependencies []interfaces.Service
}

func Get() *VagrantService {
	if globalVagrantService != nil {
		return globalVagrantService
	}
	return New()
}

func New() *VagrantService {
	globalVagrantService = &VagrantService{}
	if globalVagrantService.FindPath() == "" {
		logger.Warn("Running without support for Vagrant")
	} else {
		globalVagrantService.installed = true
	}

	globalVagrantService.SetDependencies([]interfaces.Service{})
	return globalVagrantService
}

func (s *VagrantService) Name() string {
	return "vagrant"
}

func (s *VagrantService) FindPath() string {
	logger.Info("Getting vagrant executable")
	out, err := commands.ExecuteWithNoOutput("which", "vagrant")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		logger.Warn("Vagrant executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.executable = path
		logger.Info("Vagrant found at: %s", s.executable)
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/vagrant"); err == nil {
			s.executable = "/opt/homebrew/bin/vagrant"
		} else {
			logger.Warn("Vagrant executable not found, trying to install it")
			return s.executable
		}

		logger.Info("Vagrant found at: %s", s.executable)
	}

	return s.executable
}

func (s *VagrantService) Version() string {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(cmd)
	if err != nil {
		return "unknown"
	}

	return strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(stdout, "Vagrant ", "")), "\n", "")
}

func (s *VagrantService) Install(asUser, version string, flags map[string]string) error {
	if s.installed {
		logger.Info("%s already installed", s.Name())
		return nil
	}

	// Installing service dependency
	if s.dependencies != nil {
		for _, dependency := range s.dependencies {
			if dependency == nil {
				return errors.New("Dependency is nil")
			}
			logger.Info("Installing dependency %s for %s", dependency.Name(), s.Name())
			if err := dependency.Install(asUser, "latest", flags); err != nil {
				return err
			}
		}
	}

	var cmd helpers.Command
	if asUser == "" {
		cmd = helpers.Command{
			Command: "brew",
		}
	} else {
		cmd = helpers.Command{
			Command: "sudo",
			Args:    []string{"-u", asUser, "brew"},
		}
	}

	if version == "" || version == "latest" {
		cmd.Args = append(cmd.Args, "install", "hashicorp-vagrant")
	} else {
		cmd.Args = append(cmd.Args, "install", "hashicorp-vagrant@"+version)
	}

	logger.Info("Installing %s with command: %v", s.Name(), cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	s.installed = true
	return nil
}

func (s *VagrantService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		logger.Info("Uninstalling %s", s.Name())
		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "brew",
				Args:    []string{"uninstall", "hashicorp-vagrant"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "brew", "uninstall", "hashicorp-vagrant"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}

	if uninstallDependencies {
		// Uninstall service dependency
		if s.dependencies != nil {
			for _, dependency := range s.dependencies {
				if dependency == nil {
					continue
				}
				logger.Info("Uninstalling dependency %s for %s", dependency.Name(), s.Name())
				if err := dependency.Uninstall(asUser, uninstallDependencies); err != nil {
					return err
				}
			}
		}
	}

	s.installed = false
	return nil
}

func (s *VagrantService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *VagrantService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *VagrantService) Installed() bool {
	return s.installed && s.executable != ""
}

func (s *VagrantService) Init(path string) error {
	stdout, err := helpers.ExecuteWithNoOutput(helpers.Command{
		Command:          s.executable,
		WorkingDirectory: path,
		Args:             []string{"init", "."},
	})
	if err != nil {
		println(stdout)
		buildError := fmt.Errorf("There was an error init packer folder %v, error: %v", path, err.Error())
		return buildError
	}
	return nil
}

func (s *VagrantService) Up(path string, variableFile string) error {
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
