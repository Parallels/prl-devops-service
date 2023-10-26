package system

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/errors"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/service_provider/interfaces"
	"os"
	"strings"

	"github.com/cjlapao/common-go/commands"
)

var globalSystemService *SystemService
var logger = common.Logger

type SystemService struct {
	brewExecutable string
	installed      bool
	dependencies   []interfaces.Service
}

func Get() *SystemService {
	if globalSystemService != nil {
		return globalSystemService
	}
	return New()
}

func New() *SystemService {
	globalSystemService = &SystemService{}
	if globalSystemService.FindPath() == "" {
		logger.Warn("Running without support for brew")
		return nil
	} else {
		globalSystemService.installed = true
	}

	globalSystemService.SetDependencies([]interfaces.Service{})
	return globalSystemService
}

func (s *SystemService) Name() string {
	return "system"
}

func (s *SystemService) FindPath() string {
	logger.Info("Getting brew executable")
	out, err := commands.ExecuteWithNoOutput("which", "packer")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		logger.Warn("Brew executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.brewExecutable = path
		logger.Info("Brew found at: %s", s.brewExecutable)
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/brew"); err == nil {
			s.brewExecutable = "/opt/homebrew/bin/brew"
		} else {
			logger.Warn("Brew executable not found")
			return s.brewExecutable
		}

		logger.Info("Brew found at: %s", s.brewExecutable)
	}

	return s.brewExecutable
}

func (s *SystemService) Version() string {
	cmd := helpers.Command{
		Command: s.brewExecutable,
		Args:    []string{"--version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(cmd)
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

func (s *SystemService) Install(asUser, version string, flags map[string]string) error {
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
			logger.Info("Installing dependency %s", dependency.Name())
			if err := dependency.Install(asUser, "latest", flags); err != nil {
				return err
			}
		}
	}

	cmd := helpers.Command{
		Command: "/bin/bash",
		Args:    []string{"-c", "\"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""},
	}

	logger.Info("Installing %s with command: %v", s.Name(), cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	s.installed = true
	return nil
}

func (s *SystemService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		logger.Info("Uninstalling %s", s.Name())

		cmd := helpers.Command{
			Command: "/bin/bash",
			Args:    []string{"-c", "\"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/uninstall.sh)\""},
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
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

func (s *SystemService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *SystemService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *SystemService) Installed() bool {
	return s.installed && s.brewExecutable != ""
}

func (s *SystemService) GetSystemUsers() ([]models.SystemUser, error) {
	result := make([]models.SystemUser, 0)
	out, err := commands.ExecuteWithNoOutput("dscl", ".", "list", "/Users")
	if err != nil {
		return nil, err
	}

	users := strings.Split(out, "\n")
	for _, user := range users {
		user = strings.TrimSpace(user)
		if user == "" {
			continue
		}
		userHomeDir := "/Users/" + user
		if _, err := os.Stat(userHomeDir); os.IsNotExist(err) {
			continue
		} else {
			result = append(result, models.SystemUser{
				Username: user,
				Home:     userHomeDir,
			})
		}
	}

	return result, nil
}
