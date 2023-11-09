package packer

import (
	"os"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/serviceprovider/interfaces"

	"github.com/cjlapao/common-go/commands"
)

var globalPackerService *PackerService
var logger = common.Logger

type PackerService struct {
	executable   string
	installed    bool
	dependencies []interfaces.Service
}

func Get() *PackerService {
	if globalPackerService != nil {
		return globalPackerService
	}
	return New()
}

func New() *PackerService {
	globalPackerService = &PackerService{}
	if globalPackerService.FindPath() == "" {
		logger.Warn("Running without support for packer")
	} else {
		globalPackerService.installed = true
	}

	globalPackerService.SetDependencies([]interfaces.Service{})
	return globalPackerService
}

func (s *PackerService) Name() string {
	return "packer"
}

func (s *PackerService) FindPath() string {
	logger.Info("Getting packer executable")
	out, err := commands.ExecuteWithNoOutput("which", "packer")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		logger.Warn("Packer executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.executable = path
		logger.Info("Packer found at: %s", s.executable)
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/packer"); err == nil {
			s.executable = "/opt/homebrew/bin/packer"
		} else {
			logger.Warn("Packer executable not found")
			return s.executable
		}

		logger.Info("Packer found at: %s", s.executable)
	}

	return s.executable
}

func (s *PackerService) Version() string {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"--version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(cmd)
	if err != nil {
		return "unknown"
	}

	return strings.ReplaceAll(strings.TrimSpace(stdout), "\n", "")
}

func (s *PackerService) Install(asUser, version string, flags map[string]string) error {
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
		cmd.Args = append(cmd.Args, "install", "packer")
	} else {
		cmd.Args = append(cmd.Args, "install", "packer@"+version)
	}

	logger.Info("Installing %s with command: %v", s.Name(), cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}
	s.installed = true
	return nil
}

func (s *PackerService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		logger.Info("Uninstalling %s", s.Name())

		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "brew",
				Args:    []string{"uninstall", "packer"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "brew", "uninstall", "packer"},
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

func (s *PackerService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *PackerService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *PackerService) Installed() bool {
	return s.installed && s.executable != ""
}

func (s *PackerService) Init(ctx basecontext.ApiContext, owner string, path string) error {
	var cmd helpers.Command
	if owner == "" {
		cmd = helpers.Command{
			Command:          s.executable,
			WorkingDirectory: path,
			Args:             []string{"init", "."},
		}
	} else {
		cmd = helpers.Command{
			Command:          "sudo",
			WorkingDirectory: path,
			Args:             []string{"-u", owner, s.executable, "init", "."},
		}
	}

	stdout, _, _, err := helpers.ExecuteAndWatch(cmd)
	if err != nil {
		println(stdout)
		buildError := errors.NewWithCodef(400, "There was an error init packer folder %v, error: %v", path, err.Error())
		return buildError
	}

	ctx.LogInfo("Packer folder %v initialized", path)
	return nil
}

func (s *PackerService) Build(ctx basecontext.ApiContext, owner string, path string, variableFile string) error {
	var cmd helpers.Command
	if owner == "" {
		cmd = helpers.Command{
			Command:          s.executable,
			WorkingDirectory: path,
			Args:             []string{"build", "-var-file", variableFile, "."},
		}
	} else {
		cmd = helpers.Command{
			Command:          "sudo",
			WorkingDirectory: path,
			Args:             []string{"-u", owner, s.executable, "build", "-var-file", variableFile, "."},
		}
	}

	stdout, _, _, err := helpers.ExecuteAndWatch(cmd)
	if err != nil {
		println(stdout)
		buildError := errors.NewWithCodef(400, "There was an error building packer folder %v, error: %v", path, err.Error())
		return buildError
	}

	ctx.LogInfo("Packer folder %v built", path)
	return nil
}
