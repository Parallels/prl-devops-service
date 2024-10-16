package brew

import (
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

var globalBrewService *BrewService

type BrewService struct {
	ctx            basecontext.ApiContext
	brewExecutable string
	installed      bool
	dependencies   []interfaces.Service
}

func Get() *BrewService {
	if globalBrewService != nil {
		return globalBrewService
	}

	ctx := basecontext.NewBaseContext()

	return New(ctx)
}

func New(ctx basecontext.ApiContext) *BrewService {
	globalBrewService = &BrewService{
		ctx: ctx,
	}
	svc := system.Get()

	if svc.GetOperatingSystem() == "macos" && globalBrewService.FindPath() == "" {
		ctx.LogWarnf("Running without support for brew")
		return globalBrewService
	} else {
		globalBrewService.installed = true
	}

	globalBrewService.SetDependencies([]interfaces.Service{})
	return globalBrewService
}

func (s *BrewService) Name() string {
	return "system"
}

func (s *BrewService) FindPath() string {
	s.ctx.LogInfof("Getting brew executable")
	cmd := helpers.Command{
		Command: "which",
		Args:    []string{"brew"},
	}
	out, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		s.ctx.LogWarnf("Brew executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.brewExecutable = path
		s.ctx.LogInfof("Brew found at: %s", s.brewExecutable)
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/brew"); err == nil {
			s.brewExecutable = "/opt/homebrew/bin/brew"
		} else if _, err := os.Stat("/usr/local/bin/brew"); err == nil {
			s.brewExecutable = "/usr/local/bin/brew"
		} else {
			s.ctx.LogWarnf("Brew executable not found")
			return s.brewExecutable
		}

		s.ctx.LogInfof("Brew found at: %s", s.brewExecutable)
	}

	return s.brewExecutable
}

func (s *BrewService) Version() string {
	cmd := helpers.Command{
		Command: s.brewExecutable,
		Args:    []string{"--version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return "unknown"
	}

	vParts := strings.Split(stdout, " ")
	if len(vParts) > 0 {
		return strings.TrimSpace(vParts[1])
	} else {
		return stdout
	}
}

func (s *BrewService) Install(asUser, version string, flags map[string]string) error {
	if s.installed {
		s.ctx.LogInfof("%s already installed", s.Name())
		return nil
	}

	// Installing service dependency
	if s.dependencies != nil {
		for _, dependency := range s.dependencies {
			if dependency == nil {
				return errors.New("Dependency is nil")
			}
			s.ctx.LogInfof("Installing dependency %s", dependency.Name())
			if err := dependency.Install(asUser, "latest", flags); err != nil {
				return err
			}
		}
	}

	cmd := helpers.Command{
		Command: "/bin/bash",
		Args:    []string{"-c", "\"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""},
	}

	s.ctx.LogInfof("Installing %s with command: %v", s.Name(), cmd.String())
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	s.installed = true
	return nil
}

func (s *BrewService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		s.ctx.LogInfof("Uninstalling %s", s.Name())

		cmd := helpers.Command{
			Command: "/bin/bash",
			Args:    []string{"-c", "\"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/uninstall.sh)\""},
		}

		_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
		if err != nil {
			return err
		}
	}

	if !uninstallDependencies {
		// Uninstall service dependency
		if s.dependencies != nil {
			for _, dependency := range s.dependencies {
				if dependency == nil {
					continue
				}
				s.ctx.LogInfof("Uninstalling dependency %s for %s", dependency.Name(), s.Name())
				if err := dependency.Uninstall(asUser, uninstallDependencies); err != nil {
					return err
				}
			}
		}
	}

	s.installed = false
	return nil
}

func (s *BrewService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *BrewService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *BrewService) Installed() bool {
	return s.installed && s.brewExecutable != ""
}
