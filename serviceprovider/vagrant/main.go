package vagrant

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/errors"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/serviceprovider/interfaces"
	"Parallels/pd-api-service/serviceprovider/system"
	"path/filepath"

	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/common-go/commands"
	"github.com/cjlapao/common-go/helper"
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
		} else if _, err := os.Stat("/usr/local/bin/vagrant"); err == nil {
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
		// logger.Info("Updating %s plugins", s.Name())
		// s.updatePlugins(asUser)
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
	logger.Info("Installing %s plugins", s.Name())
	s.InstallParallelsDesktopPlugin(asUser)
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

func (s *VagrantService) InstallParallelsDesktopPlugin(asUser string) error {
	if s.installed {
		logger.Info("Updating Parallels Desktop Plugin %s", s.Name())
		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "vagrant",
				Args:    []string{"plugin", "install", "vagrant-parallels"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "vagrant", "plugin", "install", "vagrant-parallels"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VagrantService) UpdatePlugins(asUser string) error {
	if s.installed {
		logger.Info("Updating Parallels Desktop Plugin %s", s.Name())
		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "vagrant",
				Args:    []string{"plugin", "update"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "vagrant", "plugin", "update"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VagrantService) getVagrantFolderPath(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) (string, error) {
	system := system.Get()
	rootDir, err := system.GetUserHome(ctx, request.Owner)
	if err != nil {
		return "", err
	}
	userId, err := system.GetUserId(ctx, request.Owner)
	if err != nil {
		return "", err
	}

	vagrantFileFolderName := ""
	if request.Name != "" {
		vagrantFileFolderName = helpers.NormalizeString(request.Name)
	} else if request.Box != "" {
		vagrantFileFolderName = helpers.NormalizeString(request.Box)
	} else {
		return "", errors.NewWithCode("Box or Name must be provided", 500)
	}

	vagrantFileFolder := filepath.Join(rootDir, fmt.Sprintf("vagrant_%s", vagrantFileFolderName))
	if err := helpers.CreateDirIfNotExist(vagrantFileFolder); err != nil {
		return "", err
	}

	if err := os.Chown(vagrantFileFolder, userId, -1); err != nil {
		return "", err
	}

	return vagrantFileFolder, nil
}

func (s *VagrantService) getVagrantFilePath(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) (string, error) {
	vagrantFileFolder, err := s.getVagrantFolderPath(ctx, request)
	if err != nil {
		return "", err
	}

	vagrantFilePath := filepath.Join(vagrantFileFolder, "Vagrantfile")

	if helper.FileExists(vagrantFilePath) {
		if err := helper.DeleteFile(vagrantFilePath); err != nil {
			return "", err
		}
	}

	return vagrantFilePath, nil
}

func (s *VagrantService) GenerateVagrantFile(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) (string, error) {
	vagrantFileContent := "Vagrant.configure(\"2\") do |config|\n"
	vagrantFileContent += fmt.Sprintf("  config.vm.box = \"%s\"\n", request.Box)
	if request.Version != "" {
		vagrantFileContent += fmt.Sprintf("config.vm.box_version = \"%s\"\n", request.Version)
	}
	if request.CustomVagrantConfig != "" {
		vagrantFileContent += request.CustomVagrantConfig
	}

	if request.Name != "" || request.CustomParallelsConfig != "" {
		vagrantFileContent += "\n"
		vagrantFileContent += "  config.vm.provider \"parallels\" do |prl|\n"
		if request.Name != "" {
			vagrantFileContent += fmt.Sprintf("    prl.name = \"%s\"\n", request.Name)
		}
		if request.CustomParallelsConfig != "" {
			vagrantFileContent += request.CustomParallelsConfig
		}
		vagrantFileContent += "    end\n"
		vagrantFileContent += "end\n"
	}

	vagrantFilePath, err := s.getVagrantFilePath(ctx, request)
	if err != nil {
		return "", err
	}

	if err := helper.WriteToFile(vagrantFileContent, vagrantFilePath); err != nil {
		return "", err
	}

	return vagrantFileContent, nil
}

func (s *VagrantService) Init(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) error {
	vagrantFileFolder, err := s.getVagrantFolderPath(ctx, request)
	if err != nil {
		return err
	}

	if content, err := s.GenerateVagrantFile(ctx, request); err != nil {
		ctx.LogError("Error generating vagrant file: %v", err)
		ctx.LogError("Vagrant file content: %v", content)
		return err
	}

	cmd := helpers.Command{
		Command:          "sudo",
		WorkingDirectory: vagrantFileFolder,
		Args:             make([]string, 0),
	}

	if request.Owner != "" {
		cmd.Args = append(cmd.Args, "-u", request.Owner, s.executable)
	} else {
		cmd.Args = append(cmd.Args, s.executable)
	}

	cmd.Args = append(cmd.Args, "init", request.Box)

	ctx.LogInfo("Initializing vagrant folder with command: %v", cmd.String())
	stdout, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		println(stdout)
		buildError := errors.Newf("There was an error init vagrant folder %v, error: %v", vagrantFileFolder, err.Error())
		return buildError
	}

	return nil
}

func (s *VagrantService) Up(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) error {
	vagrantFileFolder, err := s.getVagrantFolderPath(ctx, request)
	if err != nil {
		return err
	}

	cmd := helpers.Command{
		Command:          "sudo",
		WorkingDirectory: vagrantFileFolder,
		Args:             make([]string, 0),
	}

	if request.Owner != "" {
		cmd.Args = append(cmd.Args, "-u", request.Owner, s.executable)
	} else {
		cmd.Args = append(cmd.Args, s.executable)
	}

	cmd.Args = append(cmd.Args, "up")

	ctx.LogInfo("Bringing vagrant box %s up with command: %v", request.Box, cmd.String())
	stdout, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		println(stdout)
		buildError := errors.Newf("There was an error init vagrant folder %v, error: %v", vagrantFileFolder, err.Error())
		return buildError
	}

	return nil
}
